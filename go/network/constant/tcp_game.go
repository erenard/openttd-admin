// TcpGame mirrors com.openttd.network.constant.TcpGame. See tcp_game.h for
// updates.
//
// Java derives the wire value from the enum ordinal, with no offset: unlike the
// admin port's server packets, game packets are sent as their ordinal. Upstream
// appends new enum items at the end and keeps retired ids in place, so inserting
// or reordering a constant silently changes the protocol. The first six entries
// are the exception: tcp_game.h warns that those three pairs must stay in this
// order for clients joining a server of another version, which is why
// tcp_game_test.go pins every id.

package constant

// PacketGameType: PACKET_* in tcp_game.h, wire id = ordinal.
type PacketGameType int

const (
	// Packets sent by the socket accepting code without ever constructing a
	// client socket instance.
	PACKET_SERVER_FULL PacketGameType = 0 // The server is full and has no place for you.
	PACKET_SERVER_BANNED PacketGameType = 1 // The server has banned you.

	// Packets used by the client to join and an error message when the revision is wrong.
	PACKET_CLIENT_JOIN PacketGameType = 2 // The client telling the server it wants to join.
	PACKET_SERVER_ERROR PacketGameType = 3 // Server sending an error message to the client.

	// Unused packet types, formerly used for the pre-game lobby.
	PACKET_CLIENT_UNUSED PacketGameType = 4 // Unused.
	PACKET_SERVER_UNUSED PacketGameType = 5 // Unused.

	// Packets used to get the game info.
	PACKET_SERVER_GAME_INFO PacketGameType = 6 // Information about the server.
	PACKET_CLIENT_GAME_INFO PacketGameType = 7 // Request information about the server.

	//
	// Packets after here assume that the client and the server run the same
	// version, so their order no longer matters. The remainder of the packets
	// below are sent as part of authenticating and getting the map and other
	// important data.
	//
	// After the join step, the first is checking NewGRFs.
	PACKET_SERVER_CHECK_NEWGRFS PacketGameType = 8 // Server sends NewGRF IDs and MD5 checksums for the client to check.
	PACKET_CLIENT_NEWGRFS_CHECKED PacketGameType = 9 // Client acknowledges that it has all required NewGRFs.

	// Checking the game, and then company passwords.
	PACKET_SERVER_NEED_GAME_PASSWORD PacketGameType = 10 // Server requests the (hashed) game password.
	PACKET_CLIENT_GAME_PASSWORD PacketGameType = 11 // Clients sends the (hashed) game password.
	PACKET_SERVER_NEED_COMPANY_PASSWORD PacketGameType = 12 // Server requests the (hashed) company password.
	PACKET_CLIENT_COMPANY_PASSWORD PacketGameType = 13 // Client sends the (hashed) company password.

	// The server welcomes the authenticated client and sends information of other clients.
	PACKET_SERVER_WELCOME PacketGameType = 14 // Server welcomes you and gives you your ClientID.
	PACKET_SERVER_CLIENT_INFO PacketGameType = 15 // Server sends you information about a client.

	// Getting the savegame/map.
	PACKET_CLIENT_GETMAP PacketGameType = 16 // Client requests the actual map.
	PACKET_SERVER_WAIT PacketGameType = 17 // Server tells the client there are some people waiting for the map as well.
	PACKET_SERVER_MAP_BEGIN PacketGameType = 18 // Server tells the client that it is beginning to send the map.
	PACKET_SERVER_MAP_SIZE PacketGameType = 19 // Server tells the client what the (compressed) size of the map is.
	PACKET_SERVER_MAP_DATA PacketGameType = 20 // Server sends bits of the map to the client.
	PACKET_SERVER_MAP_DONE PacketGameType = 21 // Server tells it has just sent the last bits of the map to the client.
	PACKET_CLIENT_MAP_OK PacketGameType = 22 // Client tells the server that it received the whole map.

	PACKET_SERVER_JOIN PacketGameType = 23 // Tells clients that a new client has joined.

	//
	// At this moment the client has the map and the client is fully
	// authenticated, so the normal communication starts.
	//
	// Game progress monitoring.
	PACKET_SERVER_FRAME PacketGameType = 24 // Server tells the client what frame it is in, and thus to where the client may progress.
	PACKET_CLIENT_ACK PacketGameType = 25 // The client tells the server which frame it has executed.
	PACKET_SERVER_SYNC PacketGameType = 26 // Server tells the client what the random state should be.

	// Sending commands around.
	PACKET_CLIENT_COMMAND PacketGameType = 27 // Client executed a command and sends it to the server.
	PACKET_SERVER_COMMAND PacketGameType = 28 // Server distributes a command to (all) the clients.

	// Human communication!
	PACKET_CLIENT_CHAT PacketGameType = 29 // Client said something that should be distributed.
	PACKET_SERVER_CHAT PacketGameType = 30 // Server distributing the message of a client (or itself).
	PACKET_SERVER_EXTERNAL_CHAT PacketGameType = 31 // Server distributing the message from external source.

	// Remote console.
	PACKET_CLIENT_RCON PacketGameType = 32 // Client asks the server to execute some command.
	PACKET_SERVER_RCON PacketGameType = 33 // Response of the executed command on the server.

	// Moving a client.
	PACKET_CLIENT_MOVE PacketGameType = 34 // A client would like to be moved to another company.
	PACKET_SERVER_MOVE PacketGameType = 35 // Server tells everyone that someone is moved to another company.

	// Configuration updates.
	PACKET_CLIENT_SET_PASSWORD PacketGameType = 36 // A client (re)sets its company's password.
	PACKET_CLIENT_SET_NAME PacketGameType = 37 // A client changes its name.
	PACKET_SERVER_COMPANY_UPDATE PacketGameType = 38 // Information (password) of a company changed.
	PACKET_SERVER_CONFIG_UPDATE PacketGameType = 39 // Some network configuration important to the client changed.

	// A server quitting this game.
	PACKET_SERVER_NEWGAME PacketGameType = 40 // The server is preparing to start a new game.
	PACKET_SERVER_SHUTDOWN PacketGameType = 41 // The server is shutting down.

	// A client quitting.
	PACKET_CLIENT_QUIT PacketGameType = 42 // A client tells the server it is going to quit.
	PACKET_SERVER_QUIT PacketGameType = 43 // A server tells that a client has quit.
	PACKET_CLIENT_ERROR PacketGameType = 44 // A client reports an error to the server.
	PACKET_SERVER_ERROR_QUIT PacketGameType = 45 // A server tells that a client has hit an error and did quit.

	PACKET_END PacketGameType = 46 // Must ALWAYS be on the end of this list!! (period)
)

// PACKET_GAME_NONE stands in for the null Java's valueOf returns for an
// unknown id.
const PACKET_GAME_NONE PacketGameType = -1

// PACKET_GAME_COUNT is how many game packet types 15.3 has, sentinel included.
const PACKET_GAME_COUNT = 47

func (t PacketGameType) Id() int { return int(t) }

func (t PacketGameType) Protocol() SubProtocol { return TCP_GAME }

// PacketGameTypeValues is PacketGameType.values(), in wire order, sentinel
// included.
func PacketGameTypeValues() []PacketGameType {
	return []PacketGameType{
		PACKET_SERVER_FULL,
		PACKET_SERVER_BANNED,
		PACKET_CLIENT_JOIN,
		PACKET_SERVER_ERROR,
		PACKET_CLIENT_UNUSED,
		PACKET_SERVER_UNUSED,
		PACKET_SERVER_GAME_INFO,
		PACKET_CLIENT_GAME_INFO,
		PACKET_SERVER_CHECK_NEWGRFS,
		PACKET_CLIENT_NEWGRFS_CHECKED,
		PACKET_SERVER_NEED_GAME_PASSWORD,
		PACKET_CLIENT_GAME_PASSWORD,
		PACKET_SERVER_NEED_COMPANY_PASSWORD,
		PACKET_CLIENT_COMPANY_PASSWORD,
		PACKET_SERVER_WELCOME,
		PACKET_SERVER_CLIENT_INFO,
		PACKET_CLIENT_GETMAP,
		PACKET_SERVER_WAIT,
		PACKET_SERVER_MAP_BEGIN,
		PACKET_SERVER_MAP_SIZE,
		PACKET_SERVER_MAP_DATA,
		PACKET_SERVER_MAP_DONE,
		PACKET_CLIENT_MAP_OK,
		PACKET_SERVER_JOIN,
		PACKET_SERVER_FRAME,
		PACKET_CLIENT_ACK,
		PACKET_SERVER_SYNC,
		PACKET_CLIENT_COMMAND,
		PACKET_SERVER_COMMAND,
		PACKET_CLIENT_CHAT,
		PACKET_SERVER_CHAT,
		PACKET_SERVER_EXTERNAL_CHAT,
		PACKET_CLIENT_RCON,
		PACKET_SERVER_RCON,
		PACKET_CLIENT_MOVE,
		PACKET_SERVER_MOVE,
		PACKET_CLIENT_SET_PASSWORD,
		PACKET_CLIENT_SET_NAME,
		PACKET_SERVER_COMPANY_UPDATE,
		PACKET_SERVER_CONFIG_UPDATE,
		PACKET_SERVER_NEWGAME,
		PACKET_SERVER_SHUTDOWN,
		PACKET_CLIENT_QUIT,
		PACKET_SERVER_QUIT,
		PACKET_CLIENT_ERROR,
		PACKET_SERVER_ERROR_QUIT,
		PACKET_END,
	}
}

// PacketGameTypeOf is PacketGameType.valueOf: the constant with that ordinal,
// or PACKET_GAME_NONE when the id is unknown, which stands in for Java's null.
func PacketGameTypeOf(order int) PacketGameType {
	if order < 0 || order >= PACKET_GAME_COUNT {
		return PACKET_GAME_NONE
	}
	return PacketGameType(order)
}

// packetGameTypeNames is the source text of each PacketGameTypeValues() entry.
// The port has no enum names, and both NetworkClient.handlePacket and
// NetworkClientSender.debugPacket log the constant, so the table lives here.
var packetGameTypeNames = []string{
	"PACKET_SERVER_FULL",
	"PACKET_SERVER_BANNED",
	"PACKET_CLIENT_JOIN",
	"PACKET_SERVER_ERROR",
	"PACKET_CLIENT_UNUSED",
	"PACKET_SERVER_UNUSED",
	"PACKET_SERVER_GAME_INFO",
	"PACKET_CLIENT_GAME_INFO",
	"PACKET_SERVER_CHECK_NEWGRFS",
	"PACKET_CLIENT_NEWGRFS_CHECKED",
	"PACKET_SERVER_NEED_GAME_PASSWORD",
	"PACKET_CLIENT_GAME_PASSWORD",
	"PACKET_SERVER_NEED_COMPANY_PASSWORD",
	"PACKET_CLIENT_COMPANY_PASSWORD",
	"PACKET_SERVER_WELCOME",
	"PACKET_SERVER_CLIENT_INFO",
	"PACKET_CLIENT_GETMAP",
	"PACKET_SERVER_WAIT",
	"PACKET_SERVER_MAP_BEGIN",
	"PACKET_SERVER_MAP_SIZE",
	"PACKET_SERVER_MAP_DATA",
	"PACKET_SERVER_MAP_DONE",
	"PACKET_CLIENT_MAP_OK",
	"PACKET_SERVER_JOIN",
	"PACKET_SERVER_FRAME",
	"PACKET_CLIENT_ACK",
	"PACKET_SERVER_SYNC",
	"PACKET_CLIENT_COMMAND",
	"PACKET_SERVER_COMMAND",
	"PACKET_CLIENT_CHAT",
	"PACKET_SERVER_CHAT",
	"PACKET_SERVER_EXTERNAL_CHAT",
	"PACKET_CLIENT_RCON",
	"PACKET_SERVER_RCON",
	"PACKET_CLIENT_MOVE",
	"PACKET_SERVER_MOVE",
	"PACKET_CLIENT_SET_PASSWORD",
	"PACKET_CLIENT_SET_NAME",
	"PACKET_SERVER_COMPANY_UPDATE",
	"PACKET_SERVER_CONFIG_UPDATE",
	"PACKET_SERVER_NEWGAME",
	"PACKET_SERVER_SHUTDOWN",
	"PACKET_CLIENT_QUIT",
	"PACKET_SERVER_QUIT",
	"PACKET_CLIENT_ERROR",
	"PACKET_SERVER_ERROR_QUIT",
	"PACKET_END",
}

// PacketGameTypeName is the enum constant's name, i.e. what the "Rcv ..." and
// "Snd ..." log lines print.
func PacketGameTypeName(t PacketGameType) string {
	if int(t) < 0 || int(t) >= len(packetGameTypeNames) {
		return "PACKET_END"
	}
	return packetGameTypeNames[int(t)]
}
