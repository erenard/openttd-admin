// FakeGameServer is the game-port twin of FakeAdminServer: a minimal server on
// the client port (3979) that accepts one connection, reads the packets the
// client sends and scripts the packets a real server would send. It frames
// packets the way src/network/core/packet.cpp does (uint16 length including the
// header, then uint8 packet type, then payload), so NetworkClient's
// Socket.receive() sees exactly what OpenTTD would send.
//
// The value of the fake is the same as the admin one: a real 15.3 dedicated
// server refuses this client's join (its PacketGameType ids in the Java tree
// are stale, see tcp_game.go) and never sends the packets the listener has to
// decode - CHECK_NEWGRFS, NEED_GAME_PASSWORD, NEED_COMPANY_PASSWORD, WAIT,
// MAP_*, SYNC, COMMAND, EXTERNAL_CHAT, MOVE, CONFIG_UPDATE, ... Here every one
// of them can be scripted.
package server

import "errors"
import "net"
import "strconv"
import "sync"
import "com.openttd"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/util"

// gameLog is this file's logger. Package server already has one for
// FakeAdminServer, and a package-level var may not be declared twice.
var gameLog = util.GetLogger("FakeGameServer")

// GAME_SERVER_ID is the server id the handshake sends: NETWORK_SERVER_ID_LENGTH
// is 22 bytes including the terminator, and the company-password hash is salted
// with exactly this many bytes, so a full-length id exercises the whole salt.
const GAME_SERVER_ID = "openttd-client-test-01"

// GAME_MAP_SEED is the map seed the handshake sends.
const GAME_MAP_SEED = 123456789

// ReceivedGamePacket is one packet sent by the game client, framed and ready to
// decode.
type ReceivedGamePacket struct {
	wireTypeId int
	packet     *core.Packet
}

func (p *ReceivedGamePacket) GetWireTypeId() int { return p.wireTypeId }

// GetType is getType: PacketGameType.valueOf of the wire id, i.e. the NONE
// sentinel for a packet the client never sends.
func (p *ReceivedGamePacket) GetType() constant.PacketGameType {
	return constant.PacketGameTypeOf(p.wireTypeId)
}

// GetPacket is getPacket: positioned after the header, so the payload can be
// read with Packet's read* methods. Like the Java version it re-reads the
// header, so calling it twice starts the payload over.
func (p *ReceivedGamePacket) GetPacket() *core.Packet {
	p.packet.PrepareToRead()
	p.packet.ReadUint8()
	return p.packet
}

func (p *ReceivedGamePacket) ToString() string {
	return constant.PacketGameTypeName(p.GetType()) + " (" + strconv.Itoa(p.wireTypeId) + ")"
}

type FakeGameServer struct {
	listener           *net.TCPListener
	port               int
	received           []*ReceivedGamePacket
	receivedLock       sync.Mutex
	connectionLock     sync.Mutex
	connection         net.Conn
	connectionCount    int
	clientDisconnected bool
	running            bool
}

// NewFakeGameServer is the constructor: bind an ephemeral port, then read the
// port the kernel picked. The accept loop runs on its own thread, as Java's
// daemon thread does.
func NewFakeGameServer() (*FakeGameServer, error) {
	var address, addressError = net.ResolveTCPAddr("tcp", net.JoinHostPort("localhost", "0"))
	if addressError != nil {
		return nil, addressError
	}
	var listener, listenError = net.ListenTCP("tcp", address)
	if listenError != nil {
		return nil, listenError
	}
	var bound, boundError = net.ResolveTCPAddr("tcp", listener.Addr().String())
	if boundError != nil {
		listener.Close()
		return nil, boundError
	}
	var server = &FakeGameServer{listener: listener, port: bound.Port, running: true}
	go acceptGameLoop(server)
	return server, nil
}

// GetPort is getPort.
func (s *FakeGameServer) GetPort() int { return s.port }

// GetConnectionCount is getConnectionCount: how many game connections have
// been accepted so far. NetworkClient reconnects every 5 s, so a test that
// drops the connection watches this number grow.
func (s *FakeGameServer) GetConnectionCount() int {
	s.connectionLock.Lock()
	var count = s.connectionCount
	s.connectionLock.Unlock()
	return count
}

// IsClientDisconnected is isClientDisconnected: whether the game client closed
// its connection (or hit EOF).
func (s *FakeGameServer) IsClientDisconnected() bool {
	s.connectionLock.Lock()
	var disconnected = s.clientDisconnected
	s.connectionLock.Unlock()
	return disconnected
}

// acceptGameLoop is acceptLoop: accept, then decode packets until the connection
// dies. A broken connection is not an error here: the client reconnects every
// 5 s, so the loop just goes back to accepting.
func acceptGameLoop(s *FakeGameServer) {
	for s.running {
		var conn, acceptError = s.listener.Accept()
		if acceptError != nil {
			// close() closed the listener.
			return
		}
		s.connectionLock.Lock()
		s.connection = conn
		s.connectionCount += 1
		s.clientDisconnected = false
		s.connectionLock.Unlock()
		for s.running {
			var packet, readError = readGamePacket(conn)
			if readError != nil || packet == nil {
				break
			}
			s.receivedLock.Lock()
			s.received = append(s.received, packet)
			s.receivedLock.Unlock()
		}
		s.connectionLock.Lock()
		// Never write to a closed connection: send() waits for the client's
		// next one instead.
		s.connection = nil
		s.connectionLock.Unlock()
		s.clientDisconnected = true
	}
}

// readGamePacket is readPacket for the game port: one framed packet off the wire
// in the buffer shape Packet.packetToReceive expects. readExactly is shared with
// FakeAdminServer.
func readGamePacket(conn net.Conn) (*ReceivedGamePacket, error) {
	var header, headerError = readExactly(conn, 2)
	if headerError != nil {
		return nil, headerError
	}
	var length = int(header[0]&0xFF) | int(header[1]&0xFF)<<8
	if length < 3 {
		return nil, errors.New("Illegal packet length: " + strconv.Itoa(length))
	}
	var body, bodyError = readExactly(conn, length-2)
	if bodyError != nil {
		return nil, bodyError
	}
	var buffer = make([]byte, core.MTU)
	for i := 0; i < 2; i++ {
		buffer[i] = header[i]
	}
	for i := 0; i < len(body); i++ {
		buffer[2+i] = body[i]
	}
	var wireTypeId = int(body[0] & 0xFF)
	return &ReceivedGamePacket{wireTypeId: wireTypeId, packet: core.NewForReceive(buffer)}, nil
}

// ClientJoinPacket is the packet this client sends first, see
// NetworkClientSender.join: string version, uint32 newgrf version, string
// name, uint8 play-as, uint8 language. Named ClientJoinPacket because
// WelcomePacket and friends are already taken by upstream_admin_protocol.go,
// which builds the admin port's packets in this same package.
func ClientJoinPacket() *PacketWriter {
	var packet = New(constant.PACKET_CLIENT_JOIN)
	packet.WriteString(constant.OPENTTD)
	packet.WriteUint32(uint32(constant.NEWGRF))
	packet.WriteString("Strategyboard test client")
	packet.WriteUint8(255)
	packet.WriteUint8(0)
	return packet
}

// GameWelcomePacket is the SERVER_WELCOME a 15.3 server answers a join with, see
// tcp_game.h: uint32 client_id, uint32 map_seed, string server_id. Named with
// the Game prefix because WelcomePacket already belongs to
// upstream_admin_protocol.go, which builds the admin port's packets in this
// same package.
func GameWelcomePacket() *PacketWriter {
	var packet = New(constant.PACKET_SERVER_WELCOME)
	packet.WriteUint32(1)
	packet.WriteUint32(GAME_MAP_SEED)
	packet.WriteString(GAME_SERVER_ID)
	return packet
}

// handshake is handshake: wait until the client has connected, then answer its
// PACKET_CLIENT_JOIN the way a server that needs no newgrf and no game password
// would: straight with SERVER_WELCOME.
func (s *FakeGameServer) Handshake() {
	s.AwaitConnection(10)
	s.Send(GameWelcomePacket())
}

// AwaitConnection is awaitConnection: the client has opened its connection.
func (s *FakeGameServer) AwaitConnection(timeoutSeconds int) {
	openttd.Await(func() bool { return s.hasConnection() }, timeoutSeconds,
		"the game client to connect")
}

func (s *FakeGameServer) hasConnection() bool {
	s.connectionLock.Lock()
	var connected = s.connection != nil
	s.connectionLock.Unlock()
	return connected
}

// AwaitConnectionCount is awaitConnectionCount.
func (s *FakeGameServer) AwaitConnectionCount(count int, timeoutSeconds int) {
	openttd.Await(func() bool { return s.GetConnectionCount() >= count }, timeoutSeconds,
		"the game client to reconnect ("+strconv.Itoa(count)+" connections)")
}

// AwaitClientDisconnect is awaitClientDisconnect.
func (s *FakeGameServer) AwaitClientDisconnect(timeoutSeconds int) {
	openttd.Await(func() bool { return s.IsClientDisconnected() }, timeoutSeconds,
		"the game client to disconnect")
}

// Disconnect is disconnect: close the connection, as a server that refuses the
// join does after sending PACKET_SERVER_ERROR.
func (s *FakeGameServer) Disconnect() {
	s.connectionLock.Lock()
	if s.connection != nil {
		s.connection.Close()
		s.connection = nil
	}
	s.connectionLock.Unlock()
}

// Send is send: wait for a connection, then write one whole packet. A failed
// write is logged rather than raised: the harness's awaited conditions then
// report the missing packet, which is clearer than the IOException Java's
// caller propagates.
func (s *FakeGameServer) Send(packet *PacketWriter) {
	s.AwaitConnection(10)
	s.writeLocked(packet)
}

// SendAll is sendAll.
func (s *FakeGameServer) SendAll(packets []*PacketWriter) {
	for i := 0; i < len(packets); i++ {
		s.Send(packets[i])
	}
}

// SendRaw is bytes whose header has been edited, so the client sees a packet
// id the protocol does not define.
func (s *FakeGameServer) SendRaw(bytes []byte) {
	s.AwaitConnection(10)
	s.connectionLock.Lock()
	if s.connection == nil {
		s.connectionLock.Unlock()
		gameLog.Warn("FakeGameServer: connection closed before sending raw bytes")
		return
	}
	var _, err = s.connection.Write(bytes)
	s.connectionLock.Unlock()
	if err != nil {
		gameLog.Error(core.Message(err))
	}
}

func (s *FakeGameServer) writeLocked(packet *PacketWriter) {
	var bytes = packet.ToBytes()
	s.connectionLock.Lock()
	if s.connection == nil {
		s.connectionLock.Unlock()
		gameLog.Warn("FakeGameServer: connection closed before sending " + packet.ToString())
		return
	}
	var _, err = s.connection.Write(bytes)
	s.connectionLock.Unlock()
	if err != nil {
		gameLog.Error(core.Message(err))
	}
}

// Received is received(): every packet the client sent so far, oldest first.
func (s *FakeGameServer) Received() []*ReceivedGamePacket {
	s.receivedLock.Lock()
	var copy []*ReceivedGamePacket
	for i := 0; i < len(s.received); i++ {
		copy = append(copy, s.received[i])
	}
	s.receivedLock.Unlock()
	return copy
}

// AwaitPacket is awaitPacket: wait for a specific command from the client.
func (s *FakeGameServer) AwaitPacket(packetType constant.PacketGameType, timeoutSeconds int) *ReceivedGamePacket {
	var wireTypeId = int(packetType)
	openttd.Await(func() bool { return s.findPacket(wireTypeId) != nil }, timeoutSeconds,
		"PACKET_"+constant.PacketGameTypeName(packetType))
	return s.findPacket(wireTypeId)
}

func (s *FakeGameServer) findPacket(wireTypeId int) *ReceivedGamePacket {
	var all = s.Received()
	for i := 0; i < len(all); i++ {
		if all[i].GetWireTypeId() == wireTypeId {
			return all[i]
		}
	}
	return nil
}

// Packets is packets(): all packets of one command type, in the order they
// were sent.
func (s *FakeGameServer) Packets(packetType constant.PacketGameType) []*ReceivedGamePacket {
	var matching []*ReceivedGamePacket
	var all = s.Received()
	for i := 0; i < len(all); i++ {
		if all[i].GetWireTypeId() == int(packetType) {
			matching = append(matching, all[i])
		}
	}
	return matching
}

// Close is close: stop the loop, drop the connection, close the listener. The
// blocked Accept() returns an error, which ends the accept thread; Java
// interrupts it instead.
func (s *FakeGameServer) Close() {
	s.running = false
	s.Disconnect()
	s.listener.Close()
}
