// PacketWriter is the test harness's packet builder, mirroring
// com.openttd.server.PacketWriter. It writes the way the C++ server does, i.e.
// exactly as Packet::Send_uint8 / Send_uint16 / Send_uint32 / Send_int64 /
// Send_uint64 / Send_string / Send_bool do in OpenTTD's src/network/core/packet
// cpp: little endian, strings terminated by a NUL byte, and a header holding
// the total length of the packet (header included).
//
// FakeAdminServer uses it to script server->admin packets, and the sender tests
// use it to compare the bytes NetworkAdminSender produces.
package server

import "strconv"
import "com.openttd/network/constant"

type PacketWriter struct {
	packetType constant.PacketType
	payload []byte
}

func New(packetType constant.PacketType) *PacketWriter {
	return &PacketWriter{packetType: packetType, payload: make([]byte, 0)}
}

// GetWireTypeId is getWireTypeId: the id written in the packet header. Admin
// packets use their ordinal, server packets the ordinal plus 100, which is
// what each id enum's Id() already returns.
func (w *PacketWriter) GetWireTypeId() int {
	return w.packetType.Id()
}

func (w *PacketWriter) WriteUint8(value int) *PacketWriter {
	w.payload = append(w.payload, byte(value&0xFF))
	return w
}

func (w *PacketWriter) WriteBool8(value bool) *PacketWriter {
	if value {
		w.payload = append(w.payload, byte(1))
	} else {
		w.payload = append(w.payload, byte(0))
	}
	return w
}

func (w *PacketWriter) WriteUint16(value int) *PacketWriter {
	w.payload = append(w.payload, byte(value&0xFF))
	w.payload = append(w.payload, byte((value >> 8)&0xFF))
	return w
}

func (w *PacketWriter) WriteUint32(value uint32) *PacketWriter {
	for shift := 0; shift < 32; shift += 8 {
		w.payload = append(w.payload, byte(value >> shift))
	}
	return w
}

func (w *PacketWriter) WriteInt64(value int64) *PacketWriter {
	for shift := 0; shift < 64; shift += 8 {
		w.payload = append(w.payload, byte(value >> shift))
	}
	return w
}

func (w *PacketWriter) WriteUint64(value uint64) *PacketWriter {
	for shift := 0; shift < 64; shift += 8 {
		w.payload = append(w.payload, byte(value >> shift))
	}
	return w
}

func (w *PacketWriter) WriteString(value string) *PacketWriter {
	w.payload = append(w.payload, []byte(value)...)
	w.payload = append(w.payload, byte(0))
	return w
}

func (w *PacketWriter) WriteBytes(value []byte) *PacketWriter {
	w.payload = append(w.payload, value...)
	return w
}

// GetPayload is getPayload: the body, without the header.
func (w *PacketWriter) GetPayload() []byte {
	return w.payload
}

// ToBytes is toBytes: the whole packet, header included.
func (w *PacketWriter) ToBytes() []byte {
	var body = w.GetPayload()
	var length = 3 + len(body)
	var packet = make([]byte, length)
	packet[0] = byte(length & 0xFF)
	packet[1] = byte((length >> 8) & 0xFF)
	packet[2] = byte(w.GetWireTypeId())
	for i := 0; i < len(body); i++ {
		packet[3+i] = body[i]
	}
	return packet
}

// ToString is toString, for the log lines the harness produces. Java interpolates
// the enum's name; the Go port's constants carry no name, so the wire id stands
// in for it.
func (w *PacketWriter) ToString() string {
	return "packet "+strconv.Itoa(w.GetWireTypeId())+" ("+strconv.Itoa(len(w.payload))+" bytes of payload)"
}
