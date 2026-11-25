// Socket mirrors com.openttd.network.core.Socket: one connected TCP or UDP
// channel plus the framing state that turns the byte stream into Packets.
// Framing follows src/network/core/packet.cpp: each packet starts with a uint16
// holding its total length, header included, so one TCP read can deliver several
// packets or only a fraction of one.
package core

import "errors"
import "net"
import "strconv"
import "com.openttd/network/constant"
import "com.openttd/util"

var log = util.GetLogger("Socket")

// Buffer is the framing state Java keeps as a ByteBuffer: an MTU sized slice
// and how many bytes of it are currently held. Held bytes are always the
// beginning of an unfinished packet, so the next read appends after them.
type Buffer struct {
	bytes []byte
	position int
}

type Socket struct {
	channel net.Conn
	data *Buffer
}

// NewTcpSocket is Socket.newTcpSocket: connect now, fail if nothing listens.
func NewTcpSocket(host string, port int) (*Socket, error) {
	var channel, err = net.Dial("tcp", net.JoinHostPort(host, strconv.Itoa(port)))
	if err != nil {
		return nil, err
	}
	return &Socket{channel: channel}, nil
}

// NewUdpSocket is Socket.newUdpSocket: a connected UDP socket, so the kernel
// filters the replies to the one host.
func NewUdpSocket(host string, port int) (*Socket, error) {
	var address, addressError = net.ResolveUDPAddr("udp", net.JoinHostPort(host, strconv.Itoa(port)))
	if addressError != nil {
		return nil, addressError
	}
	var channel, err = net.DialUDP("udp", nil, address)
	if err != nil {
		return nil, err
	}
	return &Socket{channel: channel}, nil
}

// SendEmptyPacket is sendEmptyPacket: a packet with nothing but its header.
func (s *Socket) SendEmptyPacket(packetType constant.PacketType) {
	s.Send(PacketToSend(packetType))
}

// Send writes one whole packet. A failed write is logged, not raised: the
// listener loop notices the dead connection on the next receive(), which is
// what the Java code does with its caught IOException.
func (s *Socket) Send(packet *Packet) {
	packet.PrepareToSend()
	var _, err = s.channel.Write(packet.ToBytes())
	if err != nil {
		log.Error(Message(err))
	}
}

// Receive is receive(): packet-ize whatever the stream has offered so far.
// The packets completed by this read are returned, which can be none when a
// read only delivered part of a packet.
func (s *Socket) Receive() ([]*Packet, error) {
	if s.data == nil {
		s.data = &Buffer{bytes: make([]byte, MTU)}
	}
	var data, packets, err = receiveMore(s.channel, s.data)
	if err != nil {
		return packets, err
	}
	s.data = data
	return packets, nil
}

// receiveMore is receiveMore: one read, then cut as many packets out of the
// held bytes as are complete. Java passes the packet list in as an out
// parameter; here it comes back as a result instead. The leftover, if any, is
// returned as a fresh buffer with the unfinished packet at its start.
func receiveMore(channel net.Conn, data *Buffer) (*Buffer, []*Packet, error) {
	var packets []*Packet
	var tail = data.bytes[data.position:len(data.bytes)]
	var n, err = channel.Read(tail)
	if err != nil {
		// EOF or a closed channel: Java's read yields no bytes, and the
		// dataLength check below turns that into the connection lost error.
		n = 0
	}
	var dataLength = data.position + n
	if dataLength <= 0 {
		return data, packets, errors.New("Connection lost, received data length: "+strconv.Itoa(dataLength))
	}

	// pos walks the held bytes. It starts after the first packet's length,
	// which is read first, exactly as the Java version reads it through a
	// Packet that shares the buffer.
	var pos = 2
	var packetLength = int(data.bytes[1]&0xFF) << 8 | int(data.bytes[0]&0xFF)
	for dataLength >= packetLength && dataLength > 0 {
		var buffer = make([]byte, MTU)
		// Rewind the packet size, then copy the whole packet out of the held
		// bytes into the buffer this Packet will own.
		pos -= 2
		for i := 0; i < packetLength; i++ {
			buffer[i] = data.bytes[pos]
			pos += 1
		}
		var packet = NewForReceive(buffer)
		packet.PrepareToRead()
		packets = append(packets, packet)
		dataLength -= packetLength
		if dataLength > 0 {
			// Read the next packet's size.
			packetLength = int(data.bytes[pos+1]&0xFF) << 8 | int(data.bytes[pos]&0xFF)
			pos += 2
		} else {
			// Useful for the final if: the packet is finished.
			packetLength = 0
		}
	}

	if packetLength > dataLength {
		// The packet is unfinished: keep what we have for the next read.
		var buffer = make([]byte, MTU)
		pos -= 2
		for i := 0; i < dataLength; i++ {
			buffer[i] = data.bytes[pos]
			pos += 1
		}
		return &Buffer{bytes: buffer, position: dataLength}, packets, nil
	}
	return nil, packets, nil
}

// IsOpen is isOpen.
func (s *Socket) IsOpen() bool {
	return s.channel != nil
}

// Close is close: closing an already closed or broken channel is not an error.
func (s *Socket) Close() {
	if s.channel != nil {
		s.channel.Close()
	}
}
