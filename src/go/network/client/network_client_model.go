// NetworkClientModel mirrors com.openttd.network.client.NetworkClientModel: the
// state NetworkClient keeps while it is joined to a game server. Java's fields
// are public and its own handlePacket is the only writer, so they are exported
// here the same way Configuration's are.
package client

type NetworkClientModel struct {
	// FrameCounterServer is frame_counter_server: the frame the server has
	// reached, which is the frame the client may execute up to.
	FrameCounterServer int64
	// Token is token: the server's per-client token, sent back in the ack.
	Token int
	// LastAckFrame is last_ack_frame: the frame the client last acknowledged
	// up to, plus DAY_TICKS, which is what keeps one ack per game day.
	LastAckFrame int64
	// PasswordServerId is passwordServerId: the salt for the company password
	// hash, and the length that salt is padded to.
	PasswordServerId string
	// PasswordGameSeed is passwordGameSeed: the other half of that salt.
	PasswordGameSeed int64
	// ClientId is clientId: the id the server assigned this client in its
	// WELCOME packet.
	ClientId int64
}

// NewNetworkClientModel is NetworkClient's "new NetworkClientModel()": every
// field starts at zero, which is what Java's default constructor leaves it at.
func NewNetworkClientModel() *NetworkClientModel {
	return &NetworkClientModel{}
}

// DAY_TICKS is DAY_TICKS: the number of game ticks (frames) in a day, so the
// client sends one ack per day instead of one per frame.
const DAY_TICKS = 74
