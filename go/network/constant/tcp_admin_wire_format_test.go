// Wire format parity with the OpenTTD binary in use. The expected ids below are
// copied from the 15.3 tag, not from the working checkout (which is 16.0-dev):
//
//   - git show 15.3:src/network/core/tcp_admin.h (AdminPacketAdminType,
//     AdminPacketServerType, AdminUpdateType, AdminUpdateFrequency,
//     AdminCompanyRemoveReason)
//   - git show 15.3:src/network/network_type.h (NetworkVehicleType, DestType,
//     NetworkAction, NetworkErrorCode)
//   - git show 15.3:src/network/core/config.h (NETWORK_GAME_ADMIN_VERSION = 3)
//
// Java derives the wire value from ordinal(): admin packets use the ordinal,
// server packets the ordinal plus 100. The Go port spells the value out in each
// constant, which makes a reordering edit invisible, so this test is the only
// thing that keeps the lists in upstream order. Upstream appends new items at
// the end and keeps retired ids in place (ADMIN_PACKET_SERVER_CMD_LOGGING_OLD
// is the id ADMIN_PACKET_SERVER_CMD_LOGGING used in admin protocol version 1).
package constant_test

import "testing"
import "com.openttd/network/constant"

func TestAdminCommandIds(t *testing.T) {
	assertWireIds(t, "PacketAdminType", adminIds(), []int{
		0,  // ADMIN_PACKET_ADMIN_JOIN
		1,  // ADMIN_PACKET_ADMIN_QUIT
		2,  // ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY
		3,  // ADMIN_PACKET_ADMIN_POLL
		4,  // ADMIN_PACKET_ADMIN_CHAT
		5,  // ADMIN_PACKET_ADMIN_RCON
		6,  // ADMIN_PACKET_ADMIN_GAMESCRIPT
		7,  // ADMIN_PACKET_ADMIN_PING
		8,  // ADMIN_PACKET_ADMIN_EXTERNAL_CHAT
		9,  // ADMIN_PACKET_ADMIN_JOIN_SECURE
		10, // ADMIN_PACKET_ADMIN_AUTH_RESPONSE
	})
	if len(constant.PacketAdminTypeValues()) != 11 {
		t.Errorf("PacketAdminType.values().length: expected 11, got %d", len(constant.PacketAdminTypeValues()))
	}
}

func TestServerPacketIds(t *testing.T) {
	assertWireIds(t, "PacketServerType", serverIds(), []int{
		100, // ADMIN_PACKET_SERVER_FULL
		101, // ADMIN_PACKET_SERVER_BANNED
		102, // ADMIN_PACKET_SERVER_ERROR
		103, // ADMIN_PACKET_SERVER_PROTOCOL
		104, // ADMIN_PACKET_SERVER_WELCOME
		105, // ADMIN_PACKET_SERVER_NEWGAME
		106, // ADMIN_PACKET_SERVER_SHUTDOWN
		107, // ADMIN_PACKET_SERVER_DATE
		108, // ADMIN_PACKET_SERVER_CLIENT_JOIN
		109, // ADMIN_PACKET_SERVER_CLIENT_INFO
		110, // ADMIN_PACKET_SERVER_CLIENT_UPDATE
		111, // ADMIN_PACKET_SERVER_CLIENT_QUIT
		112, // ADMIN_PACKET_SERVER_CLIENT_ERROR
		113, // ADMIN_PACKET_SERVER_COMPANY_NEW
		114, // ADMIN_PACKET_SERVER_COMPANY_INFO
		115, // ADMIN_PACKET_SERVER_COMPANY_UPDATE
		116, // ADMIN_PACKET_SERVER_COMPANY_REMOVE
		117, // ADMIN_PACKET_SERVER_COMPANY_ECONOMY
		118, // ADMIN_PACKET_SERVER_COMPANY_STATS
		119, // ADMIN_PACKET_SERVER_CHAT
		120, // ADMIN_PACKET_SERVER_RCON
		121, // ADMIN_PACKET_SERVER_CONSOLE
		122, // ADMIN_PACKET_SERVER_CMD_NAMES
		123, // ADMIN_PACKET_SERVER_CMD_LOGGING_OLD
		124, // ADMIN_PACKET_SERVER_GAMESCRIPT
		125, // ADMIN_PACKET_SERVER_RCON_END
		126, // ADMIN_PACKET_SERVER_PONG
		127, // ADMIN_PACKET_SERVER_CMD_LOGGING
		128, // ADMIN_PACKET_SERVER_AUTH_REQUEST
		129, // ADMIN_PACKET_SERVER_ENABLE_ENCRYPTION
	})
}

func TestServerPacketRoundTrip(t *testing.T) {
	var constants = serverConstants()
	for i := 0; i < len(constants); i++ {
		if constant.PacketServerTypeOf(constants[i].Id()) != constants[i] {
			t.Errorf("PacketServerType.valueOf(%d): expected %d, got %d", constants[i].Id(), int(constants[i]), int(constant.PacketServerTypeOf(constants[i].Id())))
		}
	}
	// The sentinel is only a fallback, it must not claim a real packet id.
	if constant.PacketServerTypeOf(99) != constant.INVALID_ADMIN_PACKET {
		t.Errorf("PacketServerType.valueOf(99) must be the sentinel")
	}
	if constant.PacketServerTypeOf(130) != constant.INVALID_ADMIN_PACKET {
		t.Errorf("PacketServerType.valueOf(130) must be the sentinel")
	}
	if constant.PacketServerTypeOf(0) != constant.INVALID_ADMIN_PACKET {
		t.Errorf("PacketServerType.valueOf(0) must be the sentinel")
	}
}

func TestAdminCommandRoundTrip(t *testing.T) {
	var constants = constant.PacketAdminTypeValues()
	for i := 0; i < len(constants); i++ {
		if constant.PacketAdminTypeOf(int(constants[i])) != constants[i] {
			t.Errorf("PacketAdminType.valueOf(%d): expected %d, got %d", int(constants[i]), int(constants[i]), int(constant.PacketAdminTypeOf(int(constants[i]))))
		}
	}
	// Java returns null for an unknown id; the port uses the NONE sentinel.
	if constant.PacketAdminTypeOf(11) != constant.ADMIN_PACKET_ADMIN_NONE {
		t.Errorf("PacketAdminType.valueOf(11) must be the NONE sentinel")
	}
	if constant.PacketAdminTypeOf(100) != constant.ADMIN_PACKET_ADMIN_NONE {
		t.Errorf("PacketAdminType.valueOf(100) must be the NONE sentinel")
	}
}

func TestUpdateTypes(t *testing.T) {
	assertWireIds(t, "AdminUpdateType", updateTypeIds(), []int{
		0,  // ADMIN_UPDATE_DATE
		1,  // ADMIN_UPDATE_CLIENT_INFO
		2,  // ADMIN_UPDATE_COMPANY_INFO
		3,  // ADMIN_UPDATE_COMPANY_ECONOMY
		4,  // ADMIN_UPDATE_COMPANY_STATS
		5,  // ADMIN_UPDATE_CHAT
		6,  // ADMIN_UPDATE_CONSOLE
		7,  // ADMIN_UPDATE_CMD_NAMES
		8,  // ADMIN_UPDATE_CMD_LOGGING
		9,  // ADMIN_UPDATE_GAMESCRIPT
		10, // ADMIN_UPDATE_END
	})
	// The server iterates 0..ADMIN_UPDATE_END, an unknown id maps to the sentinel.
	if constant.AdminUpdateTypeOf(11) != constant.ADMIN_UPDATE_END {
		t.Errorf("AdminUpdateType.valueOf(11) must be ADMIN_UPDATE_END")
	}
	if constant.AdminUpdateTypeOf(-1) != constant.ADMIN_UPDATE_END {
		t.Errorf("AdminUpdateType.valueOf(-1) must be ADMIN_UPDATE_END")
	}
}

func TestUpdateFrequencyMasks(t *testing.T) {
	var expected = []int{0x01, 0x02, 0x04, 0x08, 0x10, 0x20, 0x40}
	var names = []string{"ADMIN_FREQUENCY_POLL", "ADMIN_FREQUENCY_DAILY", "ADMIN_FREQUENCY_WEEKLY", "ADMIN_FREQUENCY_MONTHLY", "ADMIN_FREQUENCY_QUARTERLY", "ADMIN_FREQUENCY_ANUALLY", "ADMIN_FREQUENCY_AUTOMATIC"}
	var masks = frequencyMasks()
	if len(masks) != len(expected) {
		t.Errorf("AdminUpdateFrequency.values().length: expected 7, got %d", len(masks))
	}
	for i := 0; i < len(expected) && i < len(masks); i++ {
		if masks[i] != expected[i] {
			t.Errorf("AdminUpdateFrequency.%s mask: expected %d, got %d", names[i], expected[i], masks[i])
		}
	}
}

func TestCompanyRemoveReasons(t *testing.T) {
	assertWireIds(t, "AdminCompanyRemoveReason", removeReasonIds(), []int{0, 1, 2, 3})
	if constant.AdminCompanyRemoveReasonOf(0) != constant.ADMIN_CRR_MANUAL {
		t.Errorf("AdminCompanyRemoveReason.valueOf(0) must be ADMIN_CRR_MANUAL")
	}
	if constant.AdminCompanyRemoveReasonOf(1) != constant.ADMIN_CRR_AUTOCLEAN {
		t.Errorf("AdminCompanyRemoveReason.valueOf(1) must be ADMIN_CRR_AUTOCLEAN")
	}
	if constant.AdminCompanyRemoveReasonOf(2) != constant.ADMIN_CRR_BANKRUPT {
		t.Errorf("AdminCompanyRemoveReason.valueOf(2) must be ADMIN_CRR_BANKRUPT")
	}
	if constant.AdminCompanyRemoveReasonOf(4) != constant.ADMIN_CRR_END {
		t.Errorf("AdminCompanyRemoveReason.valueOf(4) must be ADMIN_CRR_END")
	}
}

func TestNetworkErrorCodes(t *testing.T) {
	assertWireIds(t, "NetworkErrorCode", networkErrorIds(), []int{
		0,  // NETWORK_ERROR_GENERAL
		1,  // NETWORK_ERROR_DESYNC
		2,  // NETWORK_ERROR_SAVEGAME_FAILED
		3,  // NETWORK_ERROR_CONNECTION_LOST
		4,  // NETWORK_ERROR_ILLEGAL_PACKET
		5,  // NETWORK_ERROR_NEWGRF_MISMATCH
		6,  // NETWORK_ERROR_NOT_AUTHORIZED
		7,  // NETWORK_ERROR_NOT_EXPECTED
		8,  // NETWORK_ERROR_WRONG_REVISION
		9,  // NETWORK_ERROR_NAME_IN_USE
		10, // NETWORK_ERROR_WRONG_PASSWORD
		11, // NETWORK_ERROR_COMPANY_MISMATCH
		12, // NETWORK_ERROR_KICKED
		13, // NETWORK_ERROR_CHEATER
		14, // NETWORK_ERROR_FULL
		15, // NETWORK_ERROR_TOO_MANY_COMMANDS
		16, // NETWORK_ERROR_TIMEOUT_PASSWORD
		17, // NETWORK_ERROR_TIMEOUT_COMPUTER
		18, // NETWORK_ERROR_TIMEOUT_MAP
		19, // NETWORK_ERROR_TIMEOUT_JOIN
		20, // NETWORK_ERROR_INVALID_CLIENT_NAME
		21, // NETWORK_ERROR_NOT_ON_ALLOW_LIST
		22, // NETWORK_ERROR_NO_AUTHENTICATION_METHOD_AVAILABLE
		23, // NETWORK_ERROR_END
	})
	if constant.NetworkErrorCodeOf(10) != constant.NETWORK_ERROR_WRONG_PASSWORD {
		t.Errorf("NetworkErrorCode.valueOf(10) must be NETWORK_ERROR_WRONG_PASSWORD")
	}
	if constant.NetworkErrorCodeOf(21) != constant.NETWORK_ERROR_NOT_ON_ALLOW_LIST {
		t.Errorf("NetworkErrorCode.valueOf(21) must be NETWORK_ERROR_NOT_ON_ALLOW_LIST")
	}
	if constant.NetworkErrorCodeOf(22) != constant.NETWORK_ERROR_NO_AUTHENTICATION_METHOD_AVAILABLE {
		t.Errorf("NetworkErrorCode.valueOf(22) must be NETWORK_ERROR_NO_AUTHENTICATION_METHOD_AVAILABLE")
	}
	// Java returns null for an unknown id; the port uses the NONE sentinel.
	if constant.NetworkErrorCodeOf(24) != constant.NETWORK_ERROR_NONE {
		t.Errorf("NetworkErrorCode.valueOf(24) must be the NONE sentinel")
	}
}

func TestChatEnums(t *testing.T) {
	assertWireIds(t, "DestType", destTypeIds(), []int{0, 1, 2})
	assertWireIds(t, "NetworkAction", networkActionIds(), []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12})
	if int(constant.NETWORK_ACTION_SERVER_MESSAGE) != 2 {
		t.Errorf("NETWORK_ACTION_SERVER_MESSAGE ordinal: expected 2, got %d", int(constant.NETWORK_ACTION_SERVER_MESSAGE))
	}
	if int(constant.NETWORK_ACTION_CHAT_CLIENT) != 5 {
		t.Errorf("NETWORK_ACTION_CHAT_CLIENT ordinal: expected 5, got %d", int(constant.NETWORK_ACTION_CHAT_CLIENT))
	}
	if int(constant.NETWORK_ACTION_EXTERNAL_CHAT) != 12 {
		t.Errorf("NETWORK_ACTION_EXTERNAL_CHAT ordinal: expected 12, got %d", int(constant.NETWORK_ACTION_EXTERNAL_CHAT))
	}
	if int(constant.DESTTYPE_BROADCAST) != 0 {
		t.Errorf("DESTTYPE_BROADCAST ordinal: expected 0, got %d", int(constant.DESTTYPE_BROADCAST))
	}
	if int(constant.DESTTYPE_TEAM) != 1 {
		t.Errorf("DESTTYPE_TEAM ordinal: expected 1, got %d", int(constant.DESTTYPE_TEAM))
	}
	if int(constant.DESTTYPE_CLIENT) != 2 {
		t.Errorf("DESTTYPE_CLIENT ordinal: expected 2, got %d", int(constant.DESTTYPE_CLIENT))
	}
}

// COMPANY_STATS sends five vehicle counts then five station counts, in
// NetworkVehicleType order (network_type.h: 0..NETWORK_VEH_END).
func TestVehicleTypeLength(t *testing.T) {
	if constant.NETWORK_VEHICLE_TYPE_LENGTH != 5 {
		t.Errorf("NetworkVehicleType.length(): expected 5, got %d", constant.NETWORK_VEHICLE_TYPE_LENGTH)
	}
	if len(constant.NetworkVehicleTypeValues()) != 6 {
		t.Errorf("NetworkVehicleType.values().length: expected 6, got %d", len(constant.NetworkVehicleTypeValues()))
	}
	var expected = []int{0, 1, 2, 3, 4, 5}
	var constants = constant.NetworkVehicleTypeValues()
	for i := 0; i < len(constants); i++ {
		if int(constants[i]) != expected[i] {
			t.Errorf("NetworkVehicleType constant %d: expected %d, got %d", i, expected[i], int(constants[i]))
		}
	}
}

func TestSubProtocol(t *testing.T) {
	var adminConstants = constant.PacketAdminTypeValues()
	for i := 0; i < len(adminConstants); i++ {
		if adminConstants[i].Protocol() != constant.TCP_ADMIN {
			t.Errorf("PacketAdminType %d sub protocol: expected TCP_ADMIN, got %d", int(adminConstants[i]), int(adminConstants[i].Protocol()))
		}
	}
	var serverConstants = serverConstants()
	for i := 0; i < len(serverConstants); i++ {
		if serverConstants[i].Protocol() != constant.TCP_ADMIN {
			t.Errorf("PacketServerType %d sub protocol: expected TCP_ADMIN, got %d", int(serverConstants[i]), int(serverConstants[i].Protocol()))
		}
	}
}

// Everything except the INVALID_ADMIN_PACKET sentinel.
func serverConstants() []constant.PacketServerType {
	var constants []constant.PacketServerType
	var all = constant.PacketServerTypeValues()
	for i := 0; i < len(all); i++ {
		if all[i] != constant.INVALID_ADMIN_PACKET {
			constants = append(constants, all[i])
		}
	}
	return constants
}

func adminIds() []int {
	var ids []int
	var all = constant.PacketAdminTypeValues()
	for i := 0; i < len(all); i++ {
		ids = append(ids, int(all[i]))
	}
	return ids
}

func serverIds() []int {
	var ids []int
	var all = serverConstants()
	for i := 0; i < len(all); i++ {
		ids = append(ids, all[i].Id())
	}
	return ids
}

func updateTypeIds() []int {
	var ids []int
	var all = constant.AdminUpdateTypeValues()
	for i := 0; i < len(all); i++ {
		ids = append(ids, int(all[i]))
	}
	return ids
}

func removeReasonIds() []int {
	var ids []int
	var all = constant.AdminCompanyRemoveReasonValues()
	for i := 0; i < len(all); i++ {
		ids = append(ids, int(all[i]))
	}
	return ids
}

func networkErrorIds() []int {
	var ids []int
	var all = constant.NetworkErrorCodeValues()
	for i := 0; i < len(all); i++ {
		ids = append(ids, int(all[i]))
	}
	return ids
}

func destTypeIds() []int {
	var ids []int
	var all = constant.DestTypeValues()
	for i := 0; i < len(all); i++ {
		ids = append(ids, int(all[i]))
	}
	return ids
}

func networkActionIds() []int {
	var ids []int
	var all = constant.NetworkActionValues()
	for i := 0; i < len(all); i++ {
		ids = append(ids, int(all[i]))
	}
	return ids
}

func frequencyMasks() []int {
	var masks []int
	var all = constant.AdminUpdateFrequencyValues()
	for i := 0; i < len(all); i++ {
		masks = append(masks, all[i].Mask())
	}
	return masks
}

// assertWireIds is the Java helper of the same name: the same number of
// constants as 15.3 has, then the ids in upstream order.
func assertWireIds(t *testing.T, name string, actual []int, expected []int) {
	if len(actual) != len(expected) {
		t.Errorf("%s has a different number of constants than 15.3: expected %d, got %d", name, len(expected), len(actual))
		return
	}
	for i := 0; i < len(expected); i++ {
		if actual[i] != expected[i] {
			t.Errorf("%s constant %d wire id: expected %d, got %d", name, i, expected[i], actual[i])
		}
	}
}
