// FakeAdminServer mirrors com.openttd.server.FakeAdminServer: a minimal admin
// network server that accepts one connection, reads the packets the client
// sends and scripts the packets a server would send. It frames packets the way
// src/network/core/packet.cpp does (uint16 length including the header, then
// uint8 packet type, then payload), so NetworkAdmin's Socket.receive() sees
// exactly what OpenTTD would send.
//
// No OpenTTD process is needed, which makes the admin protocol decodable in a
// unit test - including the packets a real server never sends to an
// ADMIN_PACKET_ADMIN_JOIN client (FULL, BANNED, CMD_LOGGING, AUTH_REQUEST, ...).
package server

import "errors"
import "net"
import "strconv"
import "sync"
import "com.openttd"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/util"

var log = util.GetLogger("FakeAdminServer")

// ReceivedPacket is one packet sent by the admin client, framed and ready to
// decode.
type ReceivedPacket struct {
	wireTypeId int
	packet *core.Packet
}

func (p *ReceivedPacket) GetWireTypeId() int { return p.wireTypeId }

// GetType is getType: PacketAdminType.valueOf of the wire id, i.e. the NONE
// sentinel for a packet the client never sends.
func (p *ReceivedPacket) GetType() constant.PacketAdminType {
	return constant.PacketAdminTypeOf(p.wireTypeId)
}

// GetPacket is getPacket: positioned after the header, so the payload can be
// read with Packet's read* methods. Like the Java version it re-reads the
// header, so calling it twice starts the payload over.
func (p *ReceivedPacket) GetPacket() *core.Packet {
	p.packet.PrepareToRead()
	p.packet.ReadUint8()
	return p.packet
}

func (p *ReceivedPacket) ToString() string {
	return constant.PacketAdminTypeName(p.GetType())+" ("+strconv.Itoa(p.wireTypeId)+")"
}

type FakeAdminServer struct {
	listener *net.TCPListener
	port int
	received []*ReceivedPacket
	receivedLock sync.Mutex
	connectionLock sync.Mutex
	connection net.Conn
	connectionCount int
	clientDisconnected bool
	running bool
}

// NewFakeAdminServer is the constructor: bind an ephemeral port, then read the
// port the kernel picked. The accept loop runs on its own thread, as Java's
// daemon thread does.
func NewFakeAdminServer() (*FakeAdminServer, error) {
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
	var server = &FakeAdminServer{listener: listener, port: bound.Port, running: true}
	go acceptLoop(server)
	return server, nil
}

// GetPort is getPort.
func (s *FakeAdminServer) GetPort() int { return s.port }

// GetConnectionCount is getConnectionCount: how many admin connections have
// been accepted so far.
func (s *FakeAdminServer) GetConnectionCount() int {
	s.connectionLock.Lock()
	var count = s.connectionCount
	s.connectionLock.Unlock()
	return count
}

// IsClientDisconnected is isClientDisconnected: whether the admin client closed
// its connection (or hit EOF).
func (s *FakeAdminServer) IsClientDisconnected() bool {
	s.connectionLock.Lock()
	var disconnected = s.clientDisconnected
	s.connectionLock.Unlock()
	return disconnected
}

// acceptLoop is acceptLoop: accept, then decode packets until the connection
// dies. A broken connection is not an error here: the client reconnects every
// 5 s, so the loop just goes back to accepting.
func acceptLoop(s *FakeAdminServer) {
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
			var packet, readError = readPacket(conn)
			if readError != nil || packet == nil {
				break
			}
			s.receivedLock.Lock()
			s.received = append(s.received, packet)
			s.receivedLock.Unlock()
			answer(s, packet)
		}
		s.connectionLock.Lock()
		// Never write to a closed connection: send() waits for the client's
		// next one instead.
		s.connection = nil
		s.connectionLock.Unlock()
		s.clientDisconnected = true
	}
}

// answer is answer: the one reply a real server gives automatically.
// ADMIN_PACKET_ADMIN_PING is answered with ADMIN_PACKET_SERVER_PONG echoing the
// identifier (Receive_ADMIN_PING in network_admin.cpp).
func answer(s *FakeAdminServer, packet *ReceivedPacket) {
	if packet.GetWireTypeId() != int(constant.ADMIN_PACKET_ADMIN_PING) {
		return
	}
	var payload = packet.GetPacket()
	var pingId = payload.ReadUint32()
	s.writeNow(New(constant.ADMIN_PACKET_SERVER_PONG).WriteUint32(pingId))
}

// readPacket is readPacket: one framed packet off the wire, in the buffer shape
// Packet.packetToReceive expects.
func readPacket(conn net.Conn) (*ReceivedPacket, error) {
	var header, headerError = readExactly(conn, 2)
	if headerError != nil {
		return nil, headerError
	}
	var length = int(header[0]&0xFF) | int(header[1]&0xFF) << 8
	if length < 3 {
		return nil, errors.New("Illegal packet length: "+strconv.Itoa(length))
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
	return &ReceivedPacket{wireTypeId: wireTypeId, packet: core.NewForReceive(buffer)}, nil
}

// readExactly is readExactly: fill count bytes, or report a lost connection.
// Java's channel.read returns -1 at EOF; Go's Read returns 0 or an error.
func readExactly(conn net.Conn, count int) ([]byte, error) {
	var bytes = make([]byte, count)
	var read = 0
	for read < count {
		var n, err = conn.Read(bytes[read:count])
		if err != nil {
			return nil, err
		}
		if n <= 0 {
			return nil, errors.New("Connection lost, received "+strconv.Itoa(read)+" of "+strconv.Itoa(count)+" bytes")
		}
		read += n
	}
	return bytes, nil
}

// handshake is handshake: wait until the client has connected, then answer its
// ADMIN_PACKET_ADMIN_JOIN the way SendProtocol() + SendWelcome() do.
func (s *FakeAdminServer) Handshake() {
	s.awaitConnection(10)
	s.Send(ProtocolPacket())
	s.Send(WelcomePacket())
}

// awaitConnection is awaitConnection. The condition reads the connection under
// the same lock acceptLoop holds, as Java's synchronized lambda does.
func (s *FakeAdminServer) awaitConnection(timeoutSeconds int) {
	openttd.Await(func() bool { return s.hasConnection() }, timeoutSeconds,
		"the admin client to connect")
}

func (s *FakeAdminServer) hasConnection() bool {
	s.connectionLock.Lock()
	var connected = s.connection != nil
	s.connectionLock.Unlock()
	return connected
}

// AwaitConnectionCount is awaitConnectionCount.
func (s *FakeAdminServer) AwaitConnectionCount(count int, timeoutSeconds int) {
	openttd.Await(func() bool { return s.GetConnectionCount() >= count }, timeoutSeconds,
		"the admin client to reconnect ("+strconv.Itoa(count)+" connections)")
}

// AwaitClientDisconnect is awaitClientDisconnect: the admin client closes its
// connection when it receives a packet that ends the session (FULL, BANNED,
// SHUTDOWN, ...).
func (s *FakeAdminServer) AwaitClientDisconnect(timeoutSeconds int) {
	openttd.Await(func() bool { return s.IsClientDisconnected() }, timeoutSeconds,
		"the admin client to disconnect")
}

// Disconnect is disconnect: close the connection, as SendError() or a full or
// banned server does.
func (s *FakeAdminServer) Disconnect() {
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
func (s *FakeAdminServer) Send(packet *PacketWriter) {
	s.awaitConnection(10)
	s.writeLocked(packet)
}

// SendAll is sendAll.
func (s *FakeAdminServer) SendAll(packets []*PacketWriter) {
	for i := 0; i < len(packets); i++ {
		s.Send(packets[i])
	}
}

func (s *FakeAdminServer) writeLocked(packet *PacketWriter) {
	var bytes = packet.ToBytes()
	s.connectionLock.Lock()
	if s.connection == nil {
		s.connectionLock.Unlock()
		log.Warn("FakeAdminServer: connection closed before sending "+packet.ToString())
		return
	}
	var _, err = s.connection.Write(bytes)
	s.connectionLock.Unlock()
	if err != nil {
		log.Error(core.Message(err))
	}
}

// writeNow is what answer uses: the connection being read is known to be open,
// so there is nothing to wait for. Java routes the automatic PONG through send()
// as well; going straight to the socket keeps a panic out of the accept thread.
func (s *FakeAdminServer) writeNow(packet *PacketWriter) {
	s.connectionLock.Lock()
	if s.connection == nil {
		s.connectionLock.Unlock()
		return
	}
	var _, err = s.connection.Write(packet.ToBytes())
	s.connectionLock.Unlock()
	if err != nil {
		log.Error(core.Message(err))
	}
}

// Received is received(): every packet the client sent so far, oldest first.
func (s *FakeAdminServer) Received() []*ReceivedPacket {
	s.receivedLock.Lock()
	var copy []*ReceivedPacket
	for i := 0; i < len(s.received); i++ {
		copy = append(copy, s.received[i])
	}
	s.receivedLock.Unlock()
	return copy
}

// AwaitPacket is awaitPacket: wait for a specific command from the client.
func (s *FakeAdminServer) AwaitPacket(packetType constant.PacketAdminType, timeoutSeconds int) *ReceivedPacket {
	var wireTypeId = int(packetType)
	openttd.Await(func() bool { return s.findPacket(wireTypeId) != nil }, timeoutSeconds,
		// Java concatenates "ADMIN_PACKET_" onto the enum's name, which already
		// starts with it; the doubled prefix is kept so failure messages match.
		"ADMIN_PACKET_"+constant.PacketAdminTypeName(packetType))
	return s.findPacket(wireTypeId)
}

func (s *FakeAdminServer) findPacket(wireTypeId int) *ReceivedPacket {
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
func (s *FakeAdminServer) Packets(packetType constant.PacketAdminType) []*ReceivedPacket {
	var matching []*ReceivedPacket
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
func (s *FakeAdminServer) Close() {
	s.running = false
	s.Disconnect()
	s.listener.Close()
}
