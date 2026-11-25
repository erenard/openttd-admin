// TcpAdmin mirrors com.openttd.network.constant.TcpAdmin. See tcp_admin.h for
// updates.
//
// Java derives the wire value from the enum ordinal: admin packets use the
// ordinal, server packets the ordinal plus 100. Upstream appends new enum items
// at the end and keeps retired ids in place (ADMIN_PACKET_SERVER_CMD_LOGGING_OLD
// is the id ADMIN_PACKET_SERVER_CMD_LOGGING used in admin protocol version 1),
// so inserting or reordering a constant silently changes the protocol. The
// wire ids are pinned by tcp_admin_wire_format_test.go.

package constant

// PacketAdminType: ADMIN_PACKET_ADMIN_* in tcp_admin.h, wire id = ordinal.
type PacketAdminType int

const (
	ADMIN_PACKET_ADMIN_JOIN PacketAdminType = 0 // The admin announces and authenticates itself to the server using an unsecured password.
	ADMIN_PACKET_ADMIN_QUIT PacketAdminType = 1 // The admin tells the server that it is quitting.
	ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY PacketAdminType = 2 // The admin tells the server the update frequency of a particular piece of information.
	ADMIN_PACKET_ADMIN_POLL PacketAdminType = 3 // The admin explicitly polls for a piece of information.
	ADMIN_PACKET_ADMIN_CHAT PacketAdminType = 4 // The admin sends a chat message to be distributed.
	ADMIN_PACKET_ADMIN_RCON PacketAdminType = 5 // The admin sends a remote console command.
	ADMIN_PACKET_ADMIN_GAMESCRIPT PacketAdminType = 6 // The admin sends a JSON string for the GameScript.
	ADMIN_PACKET_ADMIN_PING PacketAdminType = 7 // The admin sends a ping to the server, expecting a ping-reply (PONG) packet.
	ADMIN_PACKET_ADMIN_EXTERNAL_CHAT PacketAdminType = 8 // The admin sends a chat message from external source.
	ADMIN_PACKET_ADMIN_JOIN_SECURE PacketAdminType = 9 // The admin announces and starts a secure authentication handshake. Not implemented, see ADMIN_PACKET_ADMIN_JOIN.
	ADMIN_PACKET_ADMIN_AUTH_RESPONSE PacketAdminType = 10 // The admin responds to the authentication request. Not implemented, see ADMIN_PACKET_ADMIN_JOIN.
)

// ADMIN_PACKET_ADMIN_NONE stands in for the null Java's valueOf returns for an
// unknown id.
const ADMIN_PACKET_ADMIN_NONE PacketAdminType = -1

// ADMIN_PACKET_ADMIN_COUNT is how many admin commands 15.3 has.
const ADMIN_PACKET_ADMIN_COUNT = 11

func (t PacketAdminType) Id() int { return int(t) }

func (t PacketAdminType) Protocol() SubProtocol { return TCP_ADMIN }

// PacketAdminTypeValues is PacketAdminType.values(), in wire order.
func PacketAdminTypeValues() []PacketAdminType {
	return []PacketAdminType{
		ADMIN_PACKET_ADMIN_JOIN,
		ADMIN_PACKET_ADMIN_QUIT,
		ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY,
		ADMIN_PACKET_ADMIN_POLL,
		ADMIN_PACKET_ADMIN_CHAT,
		ADMIN_PACKET_ADMIN_RCON,
		ADMIN_PACKET_ADMIN_GAMESCRIPT,
		ADMIN_PACKET_ADMIN_PING,
		ADMIN_PACKET_ADMIN_EXTERNAL_CHAT,
		ADMIN_PACKET_ADMIN_JOIN_SECURE,
		ADMIN_PACKET_ADMIN_AUTH_RESPONSE,
	}
}

// PacketAdminTypeOf is PacketAdminType.valueOf.
func PacketAdminTypeOf(order int) PacketAdminType {
	if order < 0 || order >= ADMIN_PACKET_ADMIN_COUNT {
		return ADMIN_PACKET_ADMIN_NONE
	}
	return PacketAdminType(order)
}

// packetAdminTypeNames is the source text of each PacketAdminTypeValues() entry.
// Java's String.valueOf(enum) prints it; the Go port has no enum names, so the
// harness asks for them by index.
var packetAdminTypeNames = []string{
	"ADMIN_PACKET_ADMIN_JOIN",
	"ADMIN_PACKET_ADMIN_QUIT",
	"ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY",
	"ADMIN_PACKET_ADMIN_POLL",
	"ADMIN_PACKET_ADMIN_CHAT",
	"ADMIN_PACKET_ADMIN_RCON",
	"ADMIN_PACKET_ADMIN_GAMESCRIPT",
	"ADMIN_PACKET_ADMIN_PING",
	"ADMIN_PACKET_ADMIN_EXTERNAL_CHAT",
	"ADMIN_PACKET_ADMIN_JOIN_SECURE",
	"ADMIN_PACKET_ADMIN_AUTH_RESPONSE",
}

// PacketAdminTypeName is the enum constant's name, i.e. what the harness's
// failure messages and toString() print in Java.
func PacketAdminTypeName(t PacketAdminType) string {
	if t < 0 || int(t) >= len(packetAdminTypeNames) {
		return "ADMIN_PACKET_ADMIN_NONE"
	}
	return packetAdminTypeNames[int(t)]
}

// PacketServerType: ADMIN_PACKET_SERVER_* in tcp_admin.h, wire id = ordinal +
// SERVER_PACKET_START_VALUE.
type PacketServerType int

const (
	ADMIN_PACKET_SERVER_FULL PacketServerType = 0 // The server tells the admin it cannot accept the admin.
	ADMIN_PACKET_SERVER_BANNED PacketServerType = 1 // The server tells the admin it is banned.
	ADMIN_PACKET_SERVER_ERROR PacketServerType = 2 // The server tells the admin an error has occurred.
	ADMIN_PACKET_SERVER_PROTOCOL PacketServerType = 3 // The server tells the admin its protocol version.
	ADMIN_PACKET_SERVER_WELCOME PacketServerType = 4 // The server welcomes the admin to a game.
	ADMIN_PACKET_SERVER_NEWGAME PacketServerType = 5 // The server tells the admin its going to start a new game.
	ADMIN_PACKET_SERVER_SHUTDOWN PacketServerType = 6 // The server tells the admin its shutting down.

	ADMIN_PACKET_SERVER_DATE PacketServerType = 7 // The server tells the admin what the current game date is.
	ADMIN_PACKET_SERVER_CLIENT_JOIN PacketServerType = 8 // The server tells the admin that a client has joined.
	ADMIN_PACKET_SERVER_CLIENT_INFO PacketServerType = 9 // The server gives the admin information about a client.
	ADMIN_PACKET_SERVER_CLIENT_UPDATE PacketServerType = 10 // The server gives the admin an information update on a client.
	ADMIN_PACKET_SERVER_CLIENT_QUIT PacketServerType = 11 // The server tells the admin that a client quit.
	ADMIN_PACKET_SERVER_CLIENT_ERROR PacketServerType = 12 // The server tells the admin that a client caused an error.
	ADMIN_PACKET_SERVER_COMPANY_NEW PacketServerType = 13 // The server tells the admin that a new company has started.
	ADMIN_PACKET_SERVER_COMPANY_INFO PacketServerType = 14 // The server gives the admin information about a company.
	ADMIN_PACKET_SERVER_COMPANY_UPDATE PacketServerType = 15 // The server gives the admin an information update on a company.
	ADMIN_PACKET_SERVER_COMPANY_REMOVE PacketServerType = 16 // The server tells the admin that a company was removed.
	ADMIN_PACKET_SERVER_COMPANY_ECONOMY PacketServerType = 17 // The server gives the admin some economy related company information.
	ADMIN_PACKET_SERVER_COMPANY_STATS PacketServerType = 18 // The server gives the admin some statistics about a company.
	ADMIN_PACKET_SERVER_CHAT PacketServerType = 19 // The server received a chat message and relays it.
	ADMIN_PACKET_SERVER_RCON PacketServerType = 20 // The server's reply to a remove console command.
	ADMIN_PACKET_SERVER_CONSOLE PacketServerType = 21 // The server gives the admin the data that got printed to its console.
	ADMIN_PACKET_SERVER_CMD_NAMES PacketServerType = 22 // The server sends out the names of the DoCommands to the admins.
	// Used to be the type ID of ADMIN_PACKET_SERVER_CMD_LOGGING in
	// NETWORK_GAME_ADMIN_VERSION 1. Kept so legacy layouts stay decodable.
	ADMIN_PACKET_SERVER_CMD_LOGGING_OLD PacketServerType = 23
	ADMIN_PACKET_SERVER_GAMESCRIPT PacketServerType = 24 // The server gives the admin information from the GameScript in JSON.
	ADMIN_PACKET_SERVER_RCON_END PacketServerType = 25 // The server indicates that the remote console command has completed.
	ADMIN_PACKET_SERVER_PONG PacketServerType = 26 // The server replies to a ping request from the admin.
	ADMIN_PACKET_SERVER_CMD_LOGGING PacketServerType = 27 // The server gives the admin copies of incoming command packets.
	ADMIN_PACKET_SERVER_AUTH_REQUEST PacketServerType = 28 // The server gives the admin the used authentication method and required parameters. Not implemented, see ADMIN_PACKET_ADMIN_JOIN.
	ADMIN_PACKET_SERVER_ENABLE_ENCRYPTION PacketServerType = 29 // The server tells that authentication completed and requests encryption with the keys of the last ADMIN_PACKET_ADMIN_AUTH_RESPONSE. Not implemented.

	INVALID_ADMIN_PACKET PacketServerType = 30 // Sentinel for unknown packets.
)

const SERVER_PACKET_START_VALUE = 100

func (t PacketServerType) Id() int { return int(t) + SERVER_PACKET_START_VALUE }

func (t PacketServerType) Protocol() SubProtocol { return TCP_ADMIN }

// PacketServerTypeValues is PacketServerType.values(), in wire order, sentinel
// included.
func PacketServerTypeValues() []PacketServerType {
	return []PacketServerType{
		ADMIN_PACKET_SERVER_FULL,
		ADMIN_PACKET_SERVER_BANNED,
		ADMIN_PACKET_SERVER_ERROR,
		ADMIN_PACKET_SERVER_PROTOCOL,
		ADMIN_PACKET_SERVER_WELCOME,
		ADMIN_PACKET_SERVER_NEWGAME,
		ADMIN_PACKET_SERVER_SHUTDOWN,
		ADMIN_PACKET_SERVER_DATE,
		ADMIN_PACKET_SERVER_CLIENT_JOIN,
		ADMIN_PACKET_SERVER_CLIENT_INFO,
		ADMIN_PACKET_SERVER_CLIENT_UPDATE,
		ADMIN_PACKET_SERVER_CLIENT_QUIT,
		ADMIN_PACKET_SERVER_CLIENT_ERROR,
		ADMIN_PACKET_SERVER_COMPANY_NEW,
		ADMIN_PACKET_SERVER_COMPANY_INFO,
		ADMIN_PACKET_SERVER_COMPANY_UPDATE,
		ADMIN_PACKET_SERVER_COMPANY_REMOVE,
		ADMIN_PACKET_SERVER_COMPANY_ECONOMY,
		ADMIN_PACKET_SERVER_COMPANY_STATS,
		ADMIN_PACKET_SERVER_CHAT,
		ADMIN_PACKET_SERVER_RCON,
		ADMIN_PACKET_SERVER_CONSOLE,
		ADMIN_PACKET_SERVER_CMD_NAMES,
		ADMIN_PACKET_SERVER_CMD_LOGGING_OLD,
		ADMIN_PACKET_SERVER_GAMESCRIPT,
		ADMIN_PACKET_SERVER_RCON_END,
		ADMIN_PACKET_SERVER_PONG,
		ADMIN_PACKET_SERVER_CMD_LOGGING,
		ADMIN_PACKET_SERVER_AUTH_REQUEST,
		ADMIN_PACKET_SERVER_ENABLE_ENCRYPTION,
		INVALID_ADMIN_PACKET,
	}
}

// PacketServerTypeOf is PacketServerType.valueOf: unlike the admin commands, an
// unknown id maps to the sentinel instead of null, because the server keeps
// sending new packet types and the listener must not crash on one.
func PacketServerTypeOf(order int) PacketServerType {
	if order >= SERVER_PACKET_START_VALUE && order <= int(ADMIN_PACKET_SERVER_ENABLE_ENCRYPTION)+SERVER_PACKET_START_VALUE {
		return PacketServerType(order - SERVER_PACKET_START_VALUE)
	}
	return INVALID_ADMIN_PACKET
}

// packetServerTypeNames is the source text of each PacketServerTypeValues() entry.
var packetServerTypeNames = []string{
	"ADMIN_PACKET_SERVER_FULL",
	"ADMIN_PACKET_SERVER_BANNED",
	"ADMIN_PACKET_SERVER_ERROR",
	"ADMIN_PACKET_SERVER_PROTOCOL",
	"ADMIN_PACKET_SERVER_WELCOME",
	"ADMIN_PACKET_SERVER_NEWGAME",
	"ADMIN_PACKET_SERVER_SHUTDOWN",
	"ADMIN_PACKET_SERVER_DATE",
	"ADMIN_PACKET_SERVER_CLIENT_JOIN",
	"ADMIN_PACKET_SERVER_CLIENT_INFO",
	"ADMIN_PACKET_SERVER_CLIENT_UPDATE",
	"ADMIN_PACKET_SERVER_CLIENT_QUIT",
	"ADMIN_PACKET_SERVER_CLIENT_ERROR",
	"ADMIN_PACKET_SERVER_COMPANY_NEW",
	"ADMIN_PACKET_SERVER_COMPANY_INFO",
	"ADMIN_PACKET_SERVER_COMPANY_UPDATE",
	"ADMIN_PACKET_SERVER_COMPANY_REMOVE",
	"ADMIN_PACKET_SERVER_COMPANY_ECONOMY",
	"ADMIN_PACKET_SERVER_COMPANY_STATS",
	"ADMIN_PACKET_SERVER_CHAT",
	"ADMIN_PACKET_SERVER_RCON",
	"ADMIN_PACKET_SERVER_CONSOLE",
	"ADMIN_PACKET_SERVER_CMD_NAMES",
	"ADMIN_PACKET_SERVER_CMD_LOGGING_OLD",
	"ADMIN_PACKET_SERVER_GAMESCRIPT",
	"ADMIN_PACKET_SERVER_RCON_END",
	"ADMIN_PACKET_SERVER_PONG",
	"ADMIN_PACKET_SERVER_CMD_LOGGING",
	"ADMIN_PACKET_SERVER_AUTH_REQUEST",
	"ADMIN_PACKET_SERVER_ENABLE_ENCRYPTION",
	"INVALID_ADMIN_PACKET",
}

// PacketServerTypeName is the enum constant's name.
func PacketServerTypeName(t PacketServerType) string {
	if t < 0 || int(t) >= len(packetServerTypeNames) {
		return "INVALID_ADMIN_PACKET"
	}
	return packetServerTypeNames[int(t)]
}

// AdminUpdateType: update types an admin can register a frequency for.
type AdminUpdateType int

const (
	ADMIN_UPDATE_DATE AdminUpdateType = 0 // Updates about the date of the game.
	ADMIN_UPDATE_CLIENT_INFO AdminUpdateType = 1 // Updates about the information of clients.
	ADMIN_UPDATE_COMPANY_INFO AdminUpdateType = 2 // Updates about the generic information of companies.
	ADMIN_UPDATE_COMPANY_ECONOMY AdminUpdateType = 3 // Updates about the economy of companies.
	ADMIN_UPDATE_COMPANY_STATS AdminUpdateType = 4 // Updates about the statistics of companies.
	ADMIN_UPDATE_CHAT AdminUpdateType = 5 // The admin would like to have chat messages.
	ADMIN_UPDATE_CONSOLE AdminUpdateType = 6 // The admin would like to have console messages.
	ADMIN_UPDATE_CMD_NAMES AdminUpdateType = 7 // The admin would like a list of all DoCommand names.
	ADMIN_UPDATE_CMD_LOGGING AdminUpdateType = 8 // The admin would like to have DoCommand information.
	ADMIN_UPDATE_GAMESCRIPT AdminUpdateType = 9 // The admin would like to have gamescript messages.
	ADMIN_UPDATE_END AdminUpdateType = 10 // Must ALWAYS be on the end of this list!! (period)
)

// AdminUpdateTypeOf is AdminUpdateType.valueOf. The server iterates
// 0..ADMIN_UPDATE_END, so an unknown id maps to the sentinel.
func AdminUpdateTypeOf(order int) AdminUpdateType {
	if order < 0 || order > int(ADMIN_UPDATE_END) {
		return ADMIN_UPDATE_END
	}
	return AdminUpdateType(order)
}

// AdminUpdateTypeValues is AdminUpdateType.values(), in wire order, sentinel
// included. NetworkAdmin iterates this the way the server iterates 0..END.
func AdminUpdateTypeValues() []AdminUpdateType {
	return []AdminUpdateType{
		ADMIN_UPDATE_DATE,
		ADMIN_UPDATE_CLIENT_INFO,
		ADMIN_UPDATE_COMPANY_INFO,
		ADMIN_UPDATE_COMPANY_ECONOMY,
		ADMIN_UPDATE_COMPANY_STATS,
		ADMIN_UPDATE_CHAT,
		ADMIN_UPDATE_CONSOLE,
		ADMIN_UPDATE_CMD_NAMES,
		ADMIN_UPDATE_CMD_LOGGING,
		ADMIN_UPDATE_GAMESCRIPT,
		ADMIN_UPDATE_END,
	}
}

// adminUpdateTypeNames is the source text of each AdminUpdateTypeValues() entry.
var adminUpdateTypeNames = []string{
	"ADMIN_UPDATE_DATE",
	"ADMIN_UPDATE_CLIENT_INFO",
	"ADMIN_UPDATE_COMPANY_INFO",
	"ADMIN_UPDATE_COMPANY_ECONOMY",
	"ADMIN_UPDATE_COMPANY_STATS",
	"ADMIN_UPDATE_CHAT",
	"ADMIN_UPDATE_CONSOLE",
	"ADMIN_UPDATE_CMD_NAMES",
	"ADMIN_UPDATE_CMD_LOGGING",
	"ADMIN_UPDATE_GAMESCRIPT",
	"ADMIN_UPDATE_END",
}

// AdminUpdateTypeName is the enum constant's name.
func AdminUpdateTypeName(t AdminUpdateType) string {
	if t < 0 || int(t) >= len(adminUpdateTypeNames) {
		return "ADMIN_UPDATE_END"
	}
	return adminUpdateTypeNames[int(t)]
}

// AdminUpdateFrequency: update frequencies an admin can register. The value is
// the bit mask tcp_admin.h's EnumBitSet gives each constant.
type AdminUpdateFrequency int

const (
	ADMIN_FREQUENCY_POLL AdminUpdateFrequency = 0x01 // The admin can poll this.
	ADMIN_FREQUENCY_DAILY AdminUpdateFrequency = 0x02 // The admin gets information about this on a daily basis.
	ADMIN_FREQUENCY_WEEKLY AdminUpdateFrequency = 0x04 // The admin gets information about this on a weekly basis.
	ADMIN_FREQUENCY_MONTHLY AdminUpdateFrequency = 0x08 // The admin gets information about this on a monthly basis.
	ADMIN_FREQUENCY_QUARTERLY AdminUpdateFrequency = 0x10 // The admin gets information about this on a quarterly basis.
	ADMIN_FREQUENCY_ANUALLY AdminUpdateFrequency = 0x20 // The admin gets information about this on a yearly basis.
	ADMIN_FREQUENCY_AUTOMATIC AdminUpdateFrequency = 0x40 // The admin gets information about this when it changes.
)

// ADMIN_FREQUENCY_NONE stands in for the null Java's firstSupported returns when
// the server advertises no frequency for an update type. Its mask is 0, so it
// never matches a bit.
const ADMIN_FREQUENCY_NONE AdminUpdateFrequency = 0

// Mask is the bit this frequency contributes to an update type's mask.
func (f AdminUpdateFrequency) Mask() int { return int(f) }

// AdminUpdateFrequencyValues is AdminUpdateFrequency.values().
func AdminUpdateFrequencyValues() []AdminUpdateFrequency {
	return []AdminUpdateFrequency{
		ADMIN_FREQUENCY_POLL,
		ADMIN_FREQUENCY_DAILY,
		ADMIN_FREQUENCY_WEEKLY,
		ADMIN_FREQUENCY_MONTHLY,
		ADMIN_FREQUENCY_QUARTERLY,
		ADMIN_FREQUENCY_ANUALLY,
		ADMIN_FREQUENCY_AUTOMATIC,
	}
}

// AdminUpdateFrequencyName is the enum constant's name. The masks are powers of
// two, so this is a switch rather than an index into the values list.
func AdminUpdateFrequencyName(f AdminUpdateFrequency) string {
	switch f {
	case ADMIN_FREQUENCY_POLL:
		return "ADMIN_FREQUENCY_POLL"
	case ADMIN_FREQUENCY_DAILY:
		return "ADMIN_FREQUENCY_DAILY"
	case ADMIN_FREQUENCY_WEEKLY:
		return "ADMIN_FREQUENCY_WEEKLY"
	case ADMIN_FREQUENCY_MONTHLY:
		return "ADMIN_FREQUENCY_MONTHLY"
	case ADMIN_FREQUENCY_QUARTERLY:
		return "ADMIN_FREQUENCY_QUARTERLY"
	case ADMIN_FREQUENCY_ANUALLY:
		return "ADMIN_FREQUENCY_ANUALLY"
	case ADMIN_FREQUENCY_AUTOMATIC:
		return "ADMIN_FREQUENCY_AUTOMATIC"
	default:
		return "ADMIN_FREQUENCY_UNKNOWN"
	}
}

// AdminCompanyRemoveReason: reasons for removing a company, communicated to
// admins.
type AdminCompanyRemoveReason int

const (
	ADMIN_CRR_MANUAL AdminCompanyRemoveReason = 0 // The company is manually removed.
	ADMIN_CRR_AUTOCLEAN AdminCompanyRemoveReason = 1 // The company is removed due to autoclean.
	ADMIN_CRR_BANKRUPT AdminCompanyRemoveReason = 2 // The company went belly-up.
	ADMIN_CRR_END AdminCompanyRemoveReason = 3 // Sentinel for end.
)

// AdminCompanyRemoveReasonOf is AdminCompanyRemoveReason.valueOf.
func AdminCompanyRemoveReasonOf(order int) AdminCompanyRemoveReason {
	if order < 0 || order > int(ADMIN_CRR_END) {
		return ADMIN_CRR_END
	}
	return AdminCompanyRemoveReason(order)
}

// AdminCompanyRemoveReasonValues is AdminCompanyRemoveReason.values(), sentinel
// included.
func AdminCompanyRemoveReasonValues() []AdminCompanyRemoveReason {
	return []AdminCompanyRemoveReason{
		ADMIN_CRR_MANUAL,
		ADMIN_CRR_AUTOCLEAN,
		ADMIN_CRR_BANKRUPT,
		ADMIN_CRR_END,
	}
}

// adminCompanyRemoveReasonNames is the source text of each
// AdminCompanyRemoveReasonValues() entry.
var adminCompanyRemoveReasonNames = []string{
	"ADMIN_CRR_MANUAL",
	"ADMIN_CRR_AUTOCLEAN",
	"ADMIN_CRR_BANKRUPT",
	"ADMIN_CRR_END",
}

// AdminCompanyRemoveReasonName is the enum constant's name, i.e. what
// NetworkAdminEvent.toString() prints for the reason a company was removed.
func AdminCompanyRemoveReasonName(r AdminCompanyRemoveReason) string {
	if r < 0 || int(r) >= len(adminCompanyRemoveReasonNames) {
		return "ADMIN_CRR_END"
	}
	return adminCompanyRemoveReasonNames[int(r)]
}

