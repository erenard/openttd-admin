// The wire codec, checked against OpenTTD 15.3's src/network/core/packet.cpp:
// little endian integers, NUL terminated strings, and a uint16 header holding
// the total packet length including those three header bytes.
package core_test

import "testing"
import "net"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/server"

// decode is the Java helper of the same name: put the writer's bytes through
// the receive path, then skip the packet type byte.
func decode(w *server.PacketWriter) *core.Packet {
	var packet = core.NewForReceive(w.ToBytes())
	packet.PrepareToRead()
	packet.ReadUint8()
	return packet
}

func TestHeaderLength(t *testing.T) {
	var packet = core.PacketToSend(constant.ADMIN_PACKET_ADMIN_PING)
	packet.WriteUint32(uint32(7))
	packet.PrepareToSend()

	var bytes = packet.GetBytes()
	var length = int(bytes[0]&0xFF) | int(bytes[1]&0xFF) << 8
	if length != 7 {
		t.Errorf("header length: expected 7, got %d", length)
	}
	if int(bytes[2]&0xFF) != int(constant.ADMIN_PACKET_ADMIN_PING) {
		t.Errorf("header type id: expected %d, got %d", int(constant.ADMIN_PACKET_ADMIN_PING), int(bytes[2]&0xFF))
	}
}

func TestServerPacketIdOffset(t *testing.T) {
	var bytes = server.New(constant.ADMIN_PACKET_SERVER_DATE).WriteUint32(uint32(1)).ToBytes()
	if int(bytes[2]&0xFF) != 107 {
		t.Errorf("server packet id offset: expected 107, got %d", int(bytes[2]&0xFF))
	}
}

func TestUint32LittleEndian(t *testing.T) {
	var packet = decode(server.New(constant.ADMIN_PACKET_SERVER_PONG).WriteUint32(uint32(0xDEADBEEF)))
	if packet.ReadUint32() != uint32(0xDEADBEEF) {
		t.Errorf("uint32: expected 0xDEADBEEF, got %d", packet.ReadUint32())
	}
	// The four bytes are least significant first.
	var payload = server.New(constant.ADMIN_PACKET_SERVER_PONG).WriteUint32(uint32(0x04030201)).GetPayload()
	var expected = []int{0x01, 0x02, 0x03, 0x04}
	for i := 0; i < len(expected); i++ {
		if int(payload[i]&0xFF) != expected[i] {
			t.Errorf("uint32 byte %d: expected %d, got %d", i, expected[i], int(payload[i]&0xFF))
		}
	}
}

func TestUint16LittleEndian(t *testing.T) {
	var packet = decode(server.New(constant.ADMIN_PACKET_SERVER_RCON).WriteUint16(0x1234))
	if packet.ReadUint16() != 0x1234 {
		t.Errorf("uint16: expected 0x1234, got %d", packet.ReadUint16())
	}
}

func TestUint8AndBool8(t *testing.T) {
	var packet = decode(server.New(constant.ADMIN_PACKET_SERVER_CMD_NAMES).WriteUint8(255).WriteBool8(true).WriteBool8(false))
	if packet.ReadUint8() != 255 {
		t.Errorf("uint8: expected 255, got %d", packet.ReadUint8())
	}
	if !packet.ReadBool8() {
		t.Errorf("bool8: expected true")
	}
	if packet.ReadBool8() {
		t.Errorf("bool8: expected false")
	}
}

// Send_uint64 / Send_int64 are 8 raw little endian bytes. readUint64 must
// scale the high half by 2^32, not by 2^32 - 1. Java needs a BigInteger for
// the top bit; Go's uint64 covers the whole range.
func TestUint64(t *testing.T) {
	var values = []uint64{
		uint64(0),
		uint64(1),
		uint64(4294967295), // 2^32 - 1
		uint64(4294967296), // 2^32
		uint64(4294967297), // 2^32 + 1
		uint64(9223372036854775807), // 2^63 - 1
		uint64(1) << 63, // 2^63
		^uint64(0), // 2^64 - 1
	}
	for i := 0; i < len(values); i++ {
		var value = values[i]
		var packet = decode(server.New(constant.ADMIN_PACKET_SERVER_COMPANY_ECONOMY).WriteUint64(value))
		if packet.ReadUint64() != value {
			t.Errorf("uint64 %d: got %d", value, packet.ReadUint64())
		}
	}
}

func TestInt64(t *testing.T) {
	// Long.MAX_VALUE and Long.MIN_VALUE, spelled so the compiler does not
	// fold them into an out of range constant.
	var longMax = int64(9223372036854775807)
	var longMin = int64(0) - longMax - int64(1)
	var values = []int64{
		int64(0),
		int64(1),
		int64(-1),
		int64(12345678901234),
		int64(-12345678901234),
		longMax, // Long.MAX_VALUE
		longMin, // Long.MIN_VALUE
	}
	for i := 0; i < len(values); i++ {
		var value = values[i]
		var packet = decode(server.New(constant.ADMIN_PACKET_SERVER_COMPANY_ECONOMY).WriteInt64(value))
		if packet.ReadInt64() != value {
			t.Errorf("int64 %d: got %d", value, packet.ReadInt64())
		}
	}
}

// Send_string sends UTF-8 bytes; writeString encodes UTF-8, so readString
// must decode UTF-8 rather than one byte per char.
func TestUtf8Strings(t *testing.T) {
	var values = []string{"", "hello", "Grüße", "こんにちは", "café ☕", "line\nbreak"}
	for i := 0; i < len(values); i++ {
		var value = values[i]
		var packet = decode(server.New(constant.ADMIN_PACKET_SERVER_CHAT).WriteString(value))
		if packet.ReadString() != value {
			t.Errorf("string %s: got %s", value, packet.ReadString())
		}
	}
}

func TestRemaining(t *testing.T) {
	var variableData = []byte{1, 2, 3, 4, 5}
	var writer = server.New(constant.ADMIN_PACKET_SERVER_CMD_LOGGING)
	writer.WriteUint32(uint32(2))
	writer.WriteUint8(1)
	writer.WriteUint16(37)
	writer.WriteBytes(variableData)
	writer.WriteUint32(uint32(99))
	var packet = decode(writer)
	packet.ReadUint32()
	packet.ReadUint8()
	packet.ReadUint16()
	if packet.Remaining() != len(variableData)+4 {
		t.Errorf("remaining: expected %d, got %d", len(variableData)+4, packet.Remaining())
	}
	packet.ReadBytes(len(variableData))
	if packet.Remaining() != 4 {
		t.Errorf("remaining: expected 4, got %d", packet.Remaining())
	}
	if packet.ReadUint32() != uint32(99) {
		t.Errorf("uint32: expected 99, got %d", packet.ReadUint32())
	}
	if packet.Remaining() != 0 {
		t.Errorf("remaining: expected 0, got %d", packet.Remaining())
	}
}

// COMPAT_MTU in src/network/core/config.h is 1460, and Packet.MTU is the
// buffer both sides allocate.
func TestMtu(t *testing.T) {
	if core.MTU != 1460 {
		t.Errorf("MTU: expected 1460, got %d", core.MTU)
	}
}

// Socket.receive() reassembles packets split across TCP reads, and unpacks
// packets concatenated into one read.
func TestPacketFraming(t *testing.T) {
	var address, addressError = net.ResolveTCPAddr("tcp", "localhost:0")
	if addressError != nil {
		t.Errorf("resolve: %s", core.Message(addressError))
		return
	}
	var listener, listenError = net.ListenTCP("tcp", address)
	if listenError != nil {
		t.Errorf("listen: %s", core.Message(listenError))
		return
	}
	var bound, boundError = net.ResolveTCPAddr("tcp", listener.Addr().String())
	if boundError != nil {
		t.Errorf("bound port: %s", core.Message(boundError))
		return
	}
	var client, clientError = core.NewTcpSocket("localhost", bound.Port)
	if clientError != nil {
		t.Errorf("connect: %s", core.Message(clientError))
		return
	}
	var serverSide, acceptError = listener.Accept()
	if acceptError != nil {
		t.Errorf("accept: %s", core.Message(acceptError))
		return
	}

	var datePacket = server.New(constant.ADMIN_PACKET_SERVER_DATE).WriteUint32(uint32(42))
	var pongPacket = server.New(constant.ADMIN_PACKET_SERVER_PONG).WriteUint32(uint32(7))
	var first = datePacket.ToBytes()
	var second = pongPacket.ToBytes()

	// Half of the first packet only: receive() must not report it yet.
	serverSide.Write(first[0:len(first)/2])
	var partial, partialError = client.Receive()
	if partialError != nil {
		t.Errorf("receive: %s", core.Message(partialError))
		return
	}
	if len(partial) != 0 {
		t.Errorf("partial read: expected 0 packets, got %d", len(partial))
	}

	// The rest of the first packet, plus the whole second one. A single TCP
	// read is not guaranteed to deliver both writes, so keep receiving until
	// the framing has produced both packets.
	var rest = make([]byte, 0)
	rest = append(rest, first[len(first)/2:len(first)]...)
	rest = append(rest, second...)
	serverSide.Write(rest)
	var packets = partial
	var attempts = 0
	for len(packets) < 2 && attempts < 10 {
		var more, moreError = client.Receive()
		if moreError != nil {
			t.Errorf("receive: %s", core.Message(moreError))
			return
		}
		packets = append(packets, more...)
		attempts += 1
	}
	if len(packets) != 2 {
		t.Errorf("framing: expected 2 packets, got %d", len(packets))
		return
	}

	if packets[0].ReadUint8() != 107 {
		t.Errorf("packet 0 type: expected 107, got %d", packets[0].ReadUint8())
	}
	if packets[0].ReadUint32() != uint32(42) {
		t.Errorf("packet 0 date: expected 42, got %d", packets[0].ReadUint32())
	}
	if packets[1].ReadUint8() != 126 {
		t.Errorf("packet 1 type: expected 126, got %d", packets[1].ReadUint8())
	}
	if packets[1].ReadUint32() != uint32(7) {
		t.Errorf("packet 1 pong: expected 7, got %d", packets[1].ReadUint32())
	}

	client.Close()
	serverSide.Close()
	listener.Close()
}
