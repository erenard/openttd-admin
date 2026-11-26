// FakeUdpServer is not in Java: the only Java code that exercises NetworkUdp is
// its main(), which needs a real server on :3979 and prints what it sees. This
// binds a UDP socket on an ephemeral port and answers the way
// git show 15.3:src/network/network_udp.cpp does, whose
// ServerNetworkUDPSocketHandler::Receive_CLIENT_FIND_SERVER sends an empty
// PACKET_UDP_SERVER_RESPONSE back to client_addr - both packets are header plus
// type byte only, so nothing else has to be scripted.
package server

import "strconv"
import "sync"
import "com.openttd"
import "com.openttd/network/constant"
import "com.openttd/network/core"

type ReceivedUdpPacket struct {
	wireTypeId int
	packet *core.Packet
}

func (p *ReceivedUdpPacket) GetWireTypeId() int { return p.wireTypeId }

// GetType is getType: PacketUdpType.valueOf of the wire id, i.e. the NONE
// sentinel for a packet this protocol never sends.
func (p *ReceivedUdpPacket) GetType() constant.PacketUdpType {
	return constant.PacketUdpTypeOf(p.wireTypeId)
}

// GetPacket is getPacket: positioned after the header, so the payload can be
// read with Packet's read* methods. Like the Java version it re-reads the
// header, so calling it twice starts the payload over.
func (p *ReceivedUdpPacket) GetPacket() *core.Packet {
	p.packet.PrepareToRead()
	p.packet.ReadUint8()
	return p.packet
}

func (p *ReceivedUdpPacket) ToString() string {
	return constant.PacketUdpTypeName(p.GetType())+" ("+strconv.Itoa(p.wireTypeId)+")"
}

type FakeUdpServer struct {
	socket *core.Socket
	port int
	received []*ReceivedUdpPacket
	receivedLock sync.Mutex
	running bool
}

// NewFakeUdpServer is the constructor: bind an ephemeral UDP port, report the
// number the kernel picked, and read datagrams on a thread of its own, as
// udp.cpp's NetworkUDPSocketHandler::ReceivePackets does from its own loop.
func NewFakeUdpServer() (*FakeUdpServer, error) {
	var socket, port, socketError = core.NewUdpListener("localhost", 0)
	if socketError != nil {
		return nil, socketError
	}
	var server = &FakeUdpServer{socket: socket, port: port, running: true, receivedLock: sync.Mutex{}}
	go answerUdp(server)
	return server, nil
}

// GetPort is getPort: the port a NetworkUdp should be pointed at.
func (s *FakeUdpServer) GetPort() int { return s.port }

// answerUdp is the server side of udp.h: read one datagram, record it, and
// answer a find-server query with the response packet, sent to the address that
// asked. A read error means close() closed the socket, so the thread leaves.
func answerUdp(s *FakeUdpServer) {
	for s.running {
		var packet, address, readError = s.socket.ReceiveFrom()
		if readError != nil {
			return
		}
		var wireTypeId = packet.ReadUint8()
		var entry = &ReceivedUdpPacket{wireTypeId: wireTypeId, packet: packet}
		s.receivedLock.Lock()
		s.received = append(s.received, entry)
		s.receivedLock.Unlock()
		if wireTypeId == int(constant.PACKET_UDP_CLIENT_FIND_SERVER) {
			var response = core.NewForSend(int(constant.PACKET_UDP_SERVER_RESPONSE))
			s.socket.SendTo(response, address)
		}
	}
}

// Received is received(): a snapshot of the packets the client has sent.
func (s *FakeUdpServer) Received() []*ReceivedUdpPacket {
	s.receivedLock.Lock()
	var copy []*ReceivedUdpPacket
	for i := 0; i < len(s.received); i++ {
		copy = append(copy, s.received[i])
	}
	s.receivedLock.Unlock()
	return copy
}

// AwaitPacket is awaitPacket: wait for a specific packet from the client.
func (s *FakeUdpServer) AwaitPacket(packetType constant.PacketUdpType, timeoutSeconds int) []*ReceivedUdpPacket {
	var wireTypeId = int(packetType)
	openttd.Await(func() bool { return s.countPackets(wireTypeId) > 0 }, timeoutSeconds,
		constant.PacketUdpTypeName(packetType))
	return s.Packets(packetType)
}

// Packets is packets(): every recorded packet of one type, oldest first.
func (s *FakeUdpServer) Packets(packetType constant.PacketUdpType) []*ReceivedUdpPacket {
	var wireTypeId = int(packetType)
	var found []*ReceivedUdpPacket
	for _, entry := range s.Received() {
		if entry.wireTypeId == wireTypeId {
			found = append(found, entry)
		}
	}
	return found
}

func (s *FakeUdpServer) countPackets(wireTypeId int) int {
	var count = 0
	for _, entry := range s.Received() {
		if entry.wireTypeId == wireTypeId {
			count += 1
		}
	}
	return count
}

// Close is close: stop the loop and release the socket. The blocked ReadFrom
// returns an error, which ends the thread; Java interrupts it instead.
func (s *FakeUdpServer) Close() {
	s.running = false
	s.socket.Close()
}
