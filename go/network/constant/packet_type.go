// PacketType mirrors com.openttd.network.constant.PacketType: the sub protocol
// a packet id belongs to, and what the codec needs from any packet id.

package constant

type SubProtocol int

const (
	UDP SubProtocol = 0
	TCP_ADMIN SubProtocol = 1
	TCP_GAME SubProtocol = 2 // TODO Rename TCP_CLIENT
)

// PacketType is implemented by every packet id enum (PacketAdminType,
// PacketServerType, PacketUdpType, PacketGameType). Java derives the wire id
// from the concrete enum's ordinal, so the id enum itself has to supply it.
type PacketType interface {
	// Id is the value written in the packet header.
	Id() int

	// Protocol is the sub protocol this packet is sent over.
	Protocol() SubProtocol
}
