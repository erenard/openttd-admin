// Packet mirrors com.openttd.network.core.Packet: the wire codec, checked
// against OpenTTD 15.3's src/network/core/packet.cpp. Little endian integers,
// NUL terminated strings, and a uint16 header holding the total packet length
// including those three header bytes.
//
// Java models a packet as a ByteBuffer it writes into and reads back; the
// buffer is MTU sized and zero padded past the packet, which is why reading
// past the end yields zeros instead of throwing. That is kept: a Packet owns a
// byte slice plus a cursor, and a read that would run past the slice marks the
// packet broken (Broken) and yields zeros.

package core

import "fmt"
import "com.openttd/network/constant"

// MTU is COMPAT_MTU in src/network/core/config.h, the buffer both sides allocate.
const MTU = 1460

// Java's Throwable.getMessage(): the text of an error value.
func Message(e error) string {
	return fmt.Sprint(e)
}


type Packet struct {
	typeId int
	length int
	bytes []byte
	pos int
	broken bool
}

// NewForSend starts a packet to write: two placeholder bytes for the length
// header, then the packet type id, exactly as Packet::packetToSend does.
func NewForSend(typeId int) *Packet {
	var packet = &Packet{typeId: typeId, bytes: make([]byte, MTU), pos: 3}
	packet.bytes[2] = byte(typeId)
	return packet
}

// PacketToSend is Packet.packetToSend(PacketType): the wire id comes from the
// id enum itself, since PacketType has no ordinal in Java either.
func PacketToSend(packetType constant.PacketType) *Packet {
	return NewForSend(packetType.Id())
}

// NewForReceive wraps bytes as read from the network, before the header has
// been read.
func NewForReceive(bytes []byte) *Packet {
	return &Packet{bytes: bytes, pos: 0}
}

// TypeId is getPacketTypeId: the raw id from the header, i.e. already offset
// by 100 for server packets.
func (p *Packet) TypeId() int { return p.typeId }

// GetBytes is getBuffer: the whole packet, header included.
func (p *Packet) GetBytes() []byte { return p.bytes }

// GetLength is getLength: the total wire length, header included.
func (p *Packet) GetLength() int { return p.length }

// PrepareToSend writes the length header and rewinds for reading. Java also
// sets the buffer's limit; ToBytes applies that limit instead.
func (p *Packet) PrepareToSend() int {
	p.length = p.pos
	p.bytes[0] = byte(p.length)
	p.bytes[1] = byte(p.length >> 8)
	return p.length
}

// ToBytes is the packet as it goes on the wire: the header plus the payload,
// without the MTU padding.
func (p *Packet) ToBytes() []byte {
	if p.length == 0 {
		return p.bytes[0:p.pos]
	}
	return p.bytes[0:p.length]
}

// PrepareToRead reads the length header and leaves the cursor on the packet
// type byte.
func (p *Packet) PrepareToRead() int {
	p.pos = 0
	p.length = p.ReadUint16()
	return p.length
}

// Remaining is the bytes left in this packet, i.e. its wire length minus what
// has already been read. Only meaningful after PrepareToRead.
func (p *Packet) Remaining() int { return p.length - p.pos }

// Broken reports whether a read ran past the end of the buffer, where the Java
// version would have thrown.
func (p *Packet) Broken() bool { return p.broken }

func (p *Packet) readBytes(count int) []byte {
	if count < 0 || p.pos + count > len(p.bytes) {
		p.broken = true
		return make([]byte, 0)
	}
	var returned = p.bytes[p.pos:p.pos+count]
	p.pos += count
	return returned
}

// ReadBytes is readBytes: raw bytes, e.g. a fixed size payload.
func (p *Packet) ReadBytes(count int) []byte { return p.readBytes(count) }

// ReadInt64 is readInt64: eight little endian bytes, two's complement.
func (p *Packet) ReadInt64() int64 {
	var buf = p.readBytes(8)
	if p.broken {
		return int64(0)
	}
	var value int64 = 0
	for i := 7; i >= 0; i-- {
		value |= int64(buf[i] & 0xFF) << (8*i)
	}
	return value
}

// ReadUint64 is readUint64. Java needs a BigInteger for the top bit; Go's
// uint64 covers the whole range.
func (p *Packet) ReadUint64() uint64 {
	var buf = p.readBytes(8)
	if p.broken {
		return uint64(0)
	}
	var value uint64 = 0
	for i := 7; i >= 0; i-- {
		value = value << 8 | uint64(buf[i] & 0xFF)
	}
	return value
}

// ReadUint32 is readUint32: four little endian bytes, unsigned.
func (p *Packet) ReadUint32() uint32 {
	var buf = p.readBytes(4)
	if p.broken {
		return uint32(0)
	}
	return uint32(buf[3] & 0xFF) << 24 | uint32(buf[2] & 0xFF) << 16 |
		uint32(buf[1] & 0xFF) << 8 | uint32(buf[0] & 0xFF)
}

// ReadUint16 is readUint16.
func (p *Packet) ReadUint16() int {
	var buf = p.readBytes(2)
	if p.broken {
		return 0
	}
	return int(buf[1] & 0xFF) << 8 | int(buf[0] & 0xFF)
}

// ReadUint8 is readUint8.
func (p *Packet) ReadUint8() int {
	var buf = p.readBytes(1)
	if p.broken {
	 return 0
	}
	return int(buf[0] & 0xFF)
}

// ReadString is readString: bytes up to the next NUL, decoded as UTF-8.
func (p *Packet) ReadString() string {
	if p.pos >= len(p.bytes) {
		p.broken = true
		return ""
	}
	var collected = make([]byte, 0)
	for p.pos < len(p.bytes) {
		var c = p.bytes[p.pos]
		p.pos += 1
		if c == 0 {
			break
		}
		collected = append(collected, c)
	}
	return string(collected)
}

// ReadBool8 is readBool8: a uint8 equal to 1.
func (p *Packet) ReadBool8() bool { return p.ReadUint8() == 1 }

// WriteUint32 is writeUint32.
func (p *Packet) WriteUint32(v uint32) {
	p.writeBytes([]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)})
}

// WriteUint16 is writeUint16.
func (p *Packet) WriteUint16(v int) {
	p.writeBytes([]byte{byte(v), byte(v >> 8)})
}

// WriteUint8 is writeUint8.
func (p *Packet) WriteUint8(v int) {
	p.writeBytes([]byte{byte(v)})
}

// WriteInt64 is writeInt64: eight raw little endian bytes.
func (p *Packet) WriteInt64(v int64) {
	var buf = make([]byte, 8)
	for i := 0; i < 8; i++ {
		buf[i] = byte(v >> (8*i))
	}
	p.writeBytes(buf)
}

// WriteUint64 is writeUint64: eight raw little endian bytes.
func (p *Packet) WriteUint64(v uint64) {
	var buf = make([]byte, 8)
	for i := 0; i < 8; i++ {
		buf[i] = byte(v >> (8*i))
	}
	p.writeBytes(buf)
}

// WriteString is writeString: UTF-8 bytes plus the terminating NUL.
func (p *Packet) WriteString(s string) {
	p.writeBytes(append([]byte(s), byte(0)))
}

// WriteBool8 is writeBool8.
func (p *Packet) WriteBool8(b bool) {
	if b {
		p.writeBytes([]byte{1})
	} else {
		p.writeBytes([]byte{0})
	}
}

// WriteBytes appends raw bytes.
func (p *Packet) WriteBytes(bytes []byte) { p.writeBytes(bytes) }

func (p *Packet) writeBytes(bytes []byte) {
	if p.pos + len(bytes) > len(p.bytes) {
		p.broken = true
		return
	}
	for i := 0; i < len(bytes); i++ {
		p.bytes[p.pos+i] = bytes[i]
	}
	p.pos += len(bytes)
}
