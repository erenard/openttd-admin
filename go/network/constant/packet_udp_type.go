// PacketUdpType mirrors com.openttd.network.constant.PacketUdpType. See udp.h
// for changes.

package constant

type PacketUdpType int

const (
	PACKET_UDP_CLIENT_FIND_SERVER PacketUdpType = 0 // Queries a game server for game information
	PACKET_UDP_SERVER_RESPONSE PacketUdpType = 1 // Reply of the game server with game information
	PACKET_UDP_END PacketUdpType = 2 // Must ALWAYS be on the end of this list!! (period)
)

// PACKET_UDP_NONE stands in for the null Java's valueOf returns for an unknown
// id.
const PACKET_UDP_NONE PacketUdpType = -1

func (t PacketUdpType) Id() int { return int(t) }

func (t PacketUdpType) Protocol() SubProtocol { return UDP }

// PacketUdpTypeOf is PacketUdpType.valueOf; NONE stands in for Java's null.
func PacketUdpTypeOf(order int) PacketUdpType {
	if order < 0 || order > int(PACKET_UDP_END) {
		return PACKET_UDP_NONE
	}
	return PacketUdpType(order)
}

// PacketUdpTypeValues is PacketUdpType.values(), sentinel included.
func PacketUdpTypeValues() []PacketUdpType {
	return []PacketUdpType{
		PACKET_UDP_CLIENT_FIND_SERVER,
		PACKET_UDP_SERVER_RESPONSE,
		PACKET_UDP_END,
	}
}

