// Harness self-test. NetworkAdmin's own socket is ported later, so this test
// talks to FakeAdminServer with plain TCP sockets: it proves the harness frames
// packets the way src/network/core/packet.cpp does, answers PING automatically,
// and scripts the PROTOCOL/WELCOME handshake, which is what every ported
// offline test leans on.

package server_test

import "errors"
import "strconv"
import "testing"
import "net"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/server"

// readPacket is the client side of the framing: uint16 length (header included),
// then the packet, copied into an MTU buffer so Packet can decode it.
func readPacket(conn net.Conn) (*core.Packet, error) {
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
	for i := 0; i < length; i++ {
		if i < 2 {
			buffer[i] = header[i]
		} else {
			buffer[i] = body[i-2]
		}
	}
	return core.NewForReceive(buffer), nil
}

func readExactly(conn net.Conn, count int) ([]byte, error) {
	var bytes = make([]byte, count)
	var read = 0
	for read < count {
		var n, err = conn.Read(bytes[read:count])
		if err != nil {
			return nil, err
		}
		if n <= 0 {
			return nil, errors.New("Connection lost, received " + strconv.Itoa(read) + " of " + strconv.Itoa(count) + " bytes")
		}
		read += n
	}
	return bytes, nil
}

// TestHandshakeSendsProtocolThenWelcome is the handshake the real server performs:
// PROTOCOL first, then WELCOME, both decodable by Packet.
func TestHandshakeSendsProtocolThenWelcome(t *testing.T) {
	var fake, fakeError = server.NewFakeAdminServer()
	if fakeError != nil {
		t.Fatalf("FakeAdminServer: %s", core.Message(fakeError))
	}
	defer fake.Close()

	var conn, connError = net.Dial("tcp", "localhost:"+strconv.Itoa(fake.GetPort()))
	if connError != nil {
		t.Fatalf("connect: %s", core.Message(connError))
	}
	defer conn.Close()

	// The client announces itself, then the harness answers.
	conn.Write(server.New(constant.ADMIN_PACKET_ADMIN_JOIN).WriteString("Strategyboard").ToBytes())
	fake.Handshake()

	var protocol, protocolError = readPacket(conn)
	if protocolError != nil {
		t.Fatalf("PROTOCOL: %s", core.Message(protocolError))
	}
	protocol.PrepareToRead()
	var protocolId = protocol.ReadUint8()
	if protocolId != constant.ADMIN_PACKET_SERVER_PROTOCOL.Id() {
		t.Errorf("expected the PROTOCOL packet first, got wire id %d", protocolId)
	}
	var version = protocol.ReadUint8()
	if version != server.GAME_ADMIN_VERSION {
		t.Errorf("expected admin protocol version %d, got %d", server.GAME_ADMIN_VERSION, version)
	}

	// One (bool8, uint16 type, uint16 mask) triple per update type, then false.
	var types = constant.AdminUpdateTypeValues()
	var seen = 0
	for protocol.ReadBool8() {
		var updateType = constant.AdminUpdateTypeOf(protocol.ReadUint16())
		var mask = protocol.ReadUint16()
		if mask != server.Frequency(updateType) {
			t.Errorf("%s: expected mask %d, got %d",
				constant.AdminUpdateTypeName(updateType), server.Frequency(updateType), mask)
		}
		if seen < len(types) {
			if updateType != types[seen] {
				t.Errorf("update type %d: expected %s, got %s", seen,
					constant.AdminUpdateTypeName(types[seen]), constant.AdminUpdateTypeName(updateType))
			}
		}
		seen += 1
	}
	// ADMIN_UPDATE_END is never advertised, so the server lists the other ten.
	if seen != len(types)-1 {
		t.Errorf("expected %d update types, got %d", len(types)-1, seen)
	}
	if protocol.Remaining() != 0 {
		t.Errorf("expected the PROTOCOL packet to be fully consumed, %d bytes left", protocol.Remaining())
	}

	var welcome, welcomeError = readPacket(conn)
	if welcomeError != nil {
		t.Fatalf("WELCOME: %s", core.Message(welcomeError))
	}
	welcome.PrepareToRead()
	var welcomeId = welcome.ReadUint8()
	if welcomeId != constant.ADMIN_PACKET_SERVER_WELCOME.Id() {
		t.Errorf("expected the WELCOME packet second, got wire id %d", welcomeId)
	}
	var message = welcome.ReadString()
	if message != "Strategyboard test server" {
		t.Errorf("expected the welcome message, got %s", message)
	}
	var gameVersion = welcome.ReadString()
	if gameVersion != "15.3" {
		t.Errorf("expected the version string 15.3, got %s", gameVersion)
	}
}

// TestAnswersPing is answer(): the one reply the harness gives on its own, so
// the PONG reaches the client without the test scripting it.
func TestAnswersPing(t *testing.T) {
	var fake, fakeError = server.NewFakeAdminServer()
	if fakeError != nil {
		t.Fatalf("FakeAdminServer: %s", core.Message(fakeError))
	}
	defer fake.Close()

	var conn, connError = net.Dial("tcp", "localhost:"+strconv.Itoa(fake.GetPort()))
	if connError != nil {
		t.Fatalf("connect: %s", core.Message(connError))
	}
	defer conn.Close()

	conn.Write(server.New(constant.ADMIN_PACKET_ADMIN_PING).WriteUint32(uint32(42)).ToBytes())

	var pong, pongError = readPacket(conn)
	if pongError != nil {
		t.Fatalf("PONG: %s", core.Message(pongError))
	}
	pong.PrepareToRead()
	var pongId = pong.ReadUint8()
	if pongId != constant.ADMIN_PACKET_SERVER_PONG.Id() {
		t.Errorf("expected a PONG, got wire id %d", pongId)
	}
	var pingId = pong.ReadUint32()
	if pingId != uint32(42) {
		t.Errorf("expected the PONG to echo the ping identifier, got %d", pingId)
	}
}

// TestRecordsClientPackets is received()/awaitPacket(): the harness keeps the
// packets the client sent, framed and positioned after the header.
func TestRecordsClientPackets(t *testing.T) {
	var fake, fakeError = server.NewFakeAdminServer()
	if fakeError != nil {
		t.Fatalf("FakeAdminServer: %s", core.Message(fakeError))
	}
	defer fake.Close()

	var conn, connError = net.Dial("tcp", "localhost:"+strconv.Itoa(fake.GetPort()))
	if connError != nil {
		t.Fatalf("connect: %s", core.Message(connError))
	}
	defer conn.Close()

	conn.Write(server.New(constant.ADMIN_PACKET_ADMIN_RCON).WriteUint32(uint32(1)).WriteString("pause").ToBytes())

	var join = fake.AwaitPacket(constant.ADMIN_PACKET_ADMIN_RCON, 5)
	if join == nil {
		t.Fatalf("expected the RCON packet to be recorded")
	}
	if join.GetWireTypeId() != int(constant.ADMIN_PACKET_ADMIN_RCON) {
		t.Errorf("expected wire id %d, got %d", int(constant.ADMIN_PACKET_ADMIN_RCON), join.GetWireTypeId())
	}

	var payload = join.GetPacket()
	var commandId = payload.ReadUint32()
	var command = payload.ReadString()
	if command != "pause" {
		t.Errorf("expected the RCON command 'pause', got %s", command)
	}
	if commandId != uint32(1) {
		t.Errorf("expected the RCON command id 1, got %d", commandId)
	}

	// The harness also reads back the packets it sends itself, exactly as the
	// Java one does, so the packets it scripts show up in received().
	conn.Write(server.New(constant.ADMIN_PACKET_ADMIN_QUIT).ToBytes())
	fake.AwaitPacket(constant.ADMIN_PACKET_ADMIN_QUIT, 5)
	if len(fake.Packets(constant.ADMIN_PACKET_ADMIN_QUIT)) != 1 {
		t.Errorf("expected one QUIT packet, got %d", len(fake.Packets(constant.ADMIN_PACKET_ADMIN_QUIT)))
	}
}
