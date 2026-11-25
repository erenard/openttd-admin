// NetworkType mirrors com.openttd.network.constant.NetworkType. See network_type.h
// for updates.

package constant

// NetworkVehicleType: vehicle types in the order they are sent in info packets.
// COMPANY_STATS sends five vehicle counts then five station counts in this order.
type NetworkVehicleType int

const (
	NETWORK_VEH_TRAIN NetworkVehicleType = 0
	NETWORK_VEH_LORRY NetworkVehicleType = 1
	NETWORK_VEH_BUS NetworkVehicleType = 2
	NETWORK_VEH_PLANE NetworkVehicleType = 3
	NETWORK_VEH_SHIP NetworkVehicleType = 4
	NETWORK_VEH_END NetworkVehicleType = 5
)

// NetworkVehicleTypeLength is NetworkVehicleType.length(): how many vehicle
// types a stats packet carries, i.e. the index of NETWORK_VEH_END.
const NETWORK_VEHICLE_TYPE_LENGTH = 5

// NetworkVehicleTypeValues is NetworkVehicleType.values(), sentinel included.
func NetworkVehicleTypeValues() []NetworkVehicleType {
	return []NetworkVehicleType{
		NETWORK_VEH_TRAIN,
		NETWORK_VEH_LORRY,
		NETWORK_VEH_BUS,
		NETWORK_VEH_PLANE,
		NETWORK_VEH_SHIP,
		NETWORK_VEH_END,
	}
}


// DestType: who a message or notice is addressed to.
type DestType int

const (
	DESTTYPE_BROADCAST DestType = 0 // Send message/notice to all clients (All)
	DESTTYPE_TEAM DestType = 1 // Send message/notice to everyone playing the same company (Team)
	DESTTYPE_CLIENT DestType = 2 // Send message/notice to only a certain client (Private)
)

// NetworkAction: actions that can be used for NetworkTextMessage.
type NetworkAction int

const (
	NETWORK_ACTION_JOIN NetworkAction = 0
	NETWORK_ACTION_LEAVE NetworkAction = 1
	NETWORK_ACTION_SERVER_MESSAGE NetworkAction = 2
	NETWORK_ACTION_CHAT NetworkAction = 3
	NETWORK_ACTION_CHAT_COMPANY NetworkAction = 4
	NETWORK_ACTION_CHAT_CLIENT NetworkAction = 5
	NETWORK_ACTION_GIVE_MONEY NetworkAction = 6
	NETWORK_ACTION_NAME_CHANGE NetworkAction = 7
	NETWORK_ACTION_COMPANY_SPECTATOR NetworkAction = 8
	NETWORK_ACTION_COMPANY_JOIN NetworkAction = 9
	NETWORK_ACTION_COMPANY_NEW NetworkAction = 10
	NETWORK_ACTION_KICKED NetworkAction = 11
	NETWORK_ACTION_EXTERNAL_CHAT NetworkAction = 12
)

// NetworkErrorCode: the codes the server reports failures with.
type NetworkErrorCode int

const (
	NETWORK_ERROR_GENERAL NetworkErrorCode = 0 // Try to use this one like never

	// Signals from clients
	NETWORK_ERROR_DESYNC NetworkErrorCode = 1
	NETWORK_ERROR_SAVEGAME_FAILED NetworkErrorCode = 2
	NETWORK_ERROR_CONNECTION_LOST NetworkErrorCode = 3
	NETWORK_ERROR_ILLEGAL_PACKET NetworkErrorCode = 4
	NETWORK_ERROR_NEWGRF_MISMATCH NetworkErrorCode = 5

	// Signals from servers
	NETWORK_ERROR_NOT_AUTHORIZED NetworkErrorCode = 6
	NETWORK_ERROR_NOT_EXPECTED NetworkErrorCode = 7
	NETWORK_ERROR_WRONG_REVISION NetworkErrorCode = 8
	NETWORK_ERROR_NAME_IN_USE NetworkErrorCode = 9
	NETWORK_ERROR_WRONG_PASSWORD NetworkErrorCode = 10
	NETWORK_ERROR_COMPANY_MISMATCH NetworkErrorCode = 11 // Happens in CLIENT_COMMAND
	NETWORK_ERROR_KICKED NetworkErrorCode = 12
	NETWORK_ERROR_CHEATER NetworkErrorCode = 13
	NETWORK_ERROR_FULL NetworkErrorCode = 14
	NETWORK_ERROR_TOO_MANY_COMMANDS NetworkErrorCode = 15
	NETWORK_ERROR_TIMEOUT_PASSWORD NetworkErrorCode = 16
	NETWORK_ERROR_TIMEOUT_COMPUTER NetworkErrorCode = 17
	NETWORK_ERROR_TIMEOUT_MAP NetworkErrorCode = 18
	NETWORK_ERROR_TIMEOUT_JOIN NetworkErrorCode = 19
	NETWORK_ERROR_INVALID_CLIENT_NAME NetworkErrorCode = 20
	NETWORK_ERROR_NOT_ON_ALLOW_LIST NetworkErrorCode = 21
	NETWORK_ERROR_NO_AUTHENTICATION_METHOD_AVAILABLE NetworkErrorCode = 22
	NETWORK_ERROR_END NetworkErrorCode = 23
)

const NETWORK_ERROR_CODE_COUNT = 24

// NETWORK_ERROR_NONE stands in for the null Java's valueOf returns for an
// unknown id.
const NETWORK_ERROR_NONE NetworkErrorCode = -1

// NetworkErrorCodeOf is NetworkErrorCode.valueOf: an unknown id is the NONE
// sentinel, which stands in for Java's null.
func NetworkErrorCodeOf(order int) NetworkErrorCode {
	if order < 0 || order >= NETWORK_ERROR_CODE_COUNT {
		return NETWORK_ERROR_NONE
	}
	return NetworkErrorCode(order)
}

// networkErrorCodes is the source text of each NetworkErrorCodeValues() entry.
// Java's String.valueOf(enum) prints it; the Go port has no enum names, so the
// error log lines and NetworkException's message ask for them by index.
var networkErrorCodes = []string{
	"NETWORK_ERROR_GENERAL",
	"NETWORK_ERROR_DESYNC",
	"NETWORK_ERROR_SAVEGAME_FAILED",
	"NETWORK_ERROR_CONNECTION_LOST",
	"NETWORK_ERROR_ILLEGAL_PACKET",
	"NETWORK_ERROR_NEWGRF_MISMATCH",
	"NETWORK_ERROR_NOT_AUTHORIZED",
	"NETWORK_ERROR_NOT_EXPECTED",
	"NETWORK_ERROR_WRONG_REVISION",
	"NETWORK_ERROR_NAME_IN_USE",
	"NETWORK_ERROR_WRONG_PASSWORD",
	"NETWORK_ERROR_COMPANY_MISMATCH",
	"NETWORK_ERROR_KICKED",
	"NETWORK_ERROR_CHEATER",
	"NETWORK_ERROR_FULL",
	"NETWORK_ERROR_TOO_MANY_COMMANDS",
	"NETWORK_ERROR_TIMEOUT_PASSWORD",
	"NETWORK_ERROR_TIMEOUT_COMPUTER",
	"NETWORK_ERROR_TIMEOUT_MAP",
	"NETWORK_ERROR_TIMEOUT_JOIN",
	"NETWORK_ERROR_INVALID_CLIENT_NAME",
	"NETWORK_ERROR_NOT_ON_ALLOW_LIST",
	"NETWORK_ERROR_NO_AUTHENTICATION_METHOD_AVAILABLE",
	"NETWORK_ERROR_END",
}

// NetworkErrorCodeName is the enum constant's name, i.e. what the log lines
// NetworkAdmin and NetworkClient print for an error id. Java prints "null" for
// an id outside the enum (valueOf returns null, then toString() throws); the
// Go port prints the sentinel's name instead.
func NetworkErrorCodeName(e NetworkErrorCode) string {
	if e < 0 || int(e) >= len(networkErrorCodes) {
		return "NETWORK_ERROR_NONE"
	}
	return networkErrorCodes[int(e)]
}

// DestTypeValues is DestType.values().
func DestTypeValues() []DestType {
	return []DestType{DESTTYPE_BROADCAST, DESTTYPE_TEAM, DESTTYPE_CLIENT}
}

// NetworkActionValues is NetworkAction.values().
func NetworkActionValues() []NetworkAction {
	return []NetworkAction{
		NETWORK_ACTION_JOIN,
		NETWORK_ACTION_LEAVE,
		NETWORK_ACTION_SERVER_MESSAGE,
		NETWORK_ACTION_CHAT,
		NETWORK_ACTION_CHAT_COMPANY,
		NETWORK_ACTION_CHAT_CLIENT,
		NETWORK_ACTION_GIVE_MONEY,
		NETWORK_ACTION_NAME_CHANGE,
		NETWORK_ACTION_COMPANY_SPECTATOR,
		NETWORK_ACTION_COMPANY_JOIN,
		NETWORK_ACTION_COMPANY_NEW,
		NETWORK_ACTION_KICKED,
		NETWORK_ACTION_EXTERNAL_CHAT,
	}
}

// DESTTYPE_NONE and NETWORK_ACTION_NONE stand in for the null Java's valueOf
// returns for an unknown id, as they do for the other enums in this port.
const DESTTYPE_NONE DestType = -1
const NETWORK_ACTION_NONE NetworkAction = -1

// destTypes and networkActions are the source text of each values() entry. Java's
// String.valueOf(enum) prints it; the Go port has no enum names, so the chat
// log lines NetworkAdmin prints for ADMIN_PACKET_SERVER_CHAT ask for them by
// index.
var destTypes = []string{
	"DESTTYPE_BROADCAST",
	"DESTTYPE_TEAM",
	"DESTTYPE_CLIENT",
}

var networkActions = []string{
	"NETWORK_ACTION_JOIN",
	"NETWORK_ACTION_LEAVE",
	"NETWORK_ACTION_SERVER_MESSAGE",
	"NETWORK_ACTION_CHAT",
	"NETWORK_ACTION_CHAT_COMPANY",
	"NETWORK_ACTION_CHAT_CLIENT",
	"NETWORK_ACTION_GIVE_MONEY",
	"NETWORK_ACTION_NAME_CHANGE",
	"NETWORK_ACTION_COMPANY_SPECTATOR",
	"NETWORK_ACTION_COMPANY_JOIN",
	"NETWORK_ACTION_COMPANY_NEW",
	"NETWORK_ACTION_KICKED",
	"NETWORK_ACTION_EXTERNAL_CHAT",
}

// DestTypeName is the enum constant's name.
func DestTypeName(t DestType) string {
	if t < 0 || int(t) >= len(destTypes) {
		return "DESTTYPE_NONE"
	}
	return destTypes[int(t)]
}

// NetworkActionName is the enum constant's name.
func NetworkActionName(a NetworkAction) string {
	if a < 0 || int(a) >= len(networkActions) {
		return "NETWORK_ACTION_NONE"
	}
	return networkActions[int(a)]
}

// NetworkErrorCodeValues is NetworkErrorCode.values(), sentinel included.
func NetworkErrorCodeValues() []NetworkErrorCode {
	return []NetworkErrorCode{
		NETWORK_ERROR_GENERAL,
		NETWORK_ERROR_DESYNC,
		NETWORK_ERROR_SAVEGAME_FAILED,
		NETWORK_ERROR_CONNECTION_LOST,
		NETWORK_ERROR_ILLEGAL_PACKET,
		NETWORK_ERROR_NEWGRF_MISMATCH,
		NETWORK_ERROR_NOT_AUTHORIZED,
		NETWORK_ERROR_NOT_EXPECTED,
		NETWORK_ERROR_WRONG_REVISION,
		NETWORK_ERROR_NAME_IN_USE,
		NETWORK_ERROR_WRONG_PASSWORD,
		NETWORK_ERROR_COMPANY_MISMATCH,
		NETWORK_ERROR_KICKED,
		NETWORK_ERROR_CHEATER,
		NETWORK_ERROR_FULL,
		NETWORK_ERROR_TOO_MANY_COMMANDS,
		NETWORK_ERROR_TIMEOUT_PASSWORD,
		NETWORK_ERROR_TIMEOUT_COMPUTER,
		NETWORK_ERROR_TIMEOUT_MAP,
		NETWORK_ERROR_TIMEOUT_JOIN,
		NETWORK_ERROR_INVALID_CLIENT_NAME,
		NETWORK_ERROR_NOT_ON_ALLOW_LIST,
		NETWORK_ERROR_NO_AUTHENTICATION_METHOD_AVAILABLE,
		NETWORK_ERROR_END,
	}
}

