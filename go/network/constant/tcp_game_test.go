// Wire format parity for the game port, the same rule tcp_admin_wire_format_test
//.go pins for the admin port: the ids come from
// git show 15.3:src/network/core/tcp_game.h, and the Go constants must stay in
// upstream order because the port spells each value out instead of deriving it
// from an ordinal. tcp_game.h adds a rule the admin port does not have: the
// first three pairs must keep their order for clients joining a server of
// another version, so those six ids are the ones a reordering edit would break
// the most quietly.
package constant_test

import "testing"
import "com.openttd/network/constant"

func TestGamePacketIds(t *testing.T) {
	assertWireIds(t, "PacketGameType", gameIds(), []int{
		0, // PACKET_SERVER_FULL
		1, // PACKET_SERVER_BANNED
		2, // PACKET_CLIENT_JOIN
		3, // PACKET_SERVER_ERROR
		4, // PACKET_CLIENT_UNUSED
		5, // PACKET_SERVER_UNUSED
		6, // PACKET_SERVER_GAME_INFO
		7, // PACKET_CLIENT_GAME_INFO
		8, // PACKET_SERVER_CHECK_NEWGRFS
		9, // PACKET_CLIENT_NEWGRFS_CHECKED
		10, // PACKET_SERVER_NEED_GAME_PASSWORD
		11, // PACKET_CLIENT_GAME_PASSWORD
		12, // PACKET_SERVER_NEED_COMPANY_PASSWORD
		13, // PACKET_CLIENT_COMPANY_PASSWORD
		14, // PACKET_SERVER_WELCOME
		15, // PACKET_SERVER_CLIENT_INFO
		16, // PACKET_CLIENT_GETMAP
		17, // PACKET_SERVER_WAIT
		18, // PACKET_SERVER_MAP_BEGIN
		19, // PACKET_SERVER_MAP_SIZE
		20, // PACKET_SERVER_MAP_DATA
		21, // PACKET_SERVER_MAP_DONE
		22, // PACKET_CLIENT_MAP_OK
		23, // PACKET_SERVER_JOIN
		24, // PACKET_SERVER_FRAME
		25, // PACKET_CLIENT_ACK
		26, // PACKET_SERVER_SYNC
		27, // PACKET_CLIENT_COMMAND
		28, // PACKET_SERVER_COMMAND
		29, // PACKET_CLIENT_CHAT
		30, // PACKET_SERVER_CHAT
		31, // PACKET_SERVER_EXTERNAL_CHAT
		32, // PACKET_CLIENT_RCON
		33, // PACKET_SERVER_RCON
		34, // PACKET_CLIENT_MOVE
		35, // PACKET_SERVER_MOVE
		36, // PACKET_CLIENT_SET_PASSWORD
		37, // PACKET_CLIENT_SET_NAME
		38, // PACKET_SERVER_COMPANY_UPDATE
		39, // PACKET_SERVER_CONFIG_UPDATE
		40, // PACKET_SERVER_NEWGAME
		41, // PACKET_SERVER_SHUTDOWN
		42, // PACKET_CLIENT_QUIT
		43, // PACKET_SERVER_QUIT
		44, // PACKET_CLIENT_ERROR
		45, // PACKET_SERVER_ERROR_QUIT
		46, // PACKET_END
	})
	if len(constant.PacketGameTypeValues()) != 47 {
		t.Errorf("PacketGameType.values().length: expected 47, got %d", len(constant.PacketGameTypeValues()))
	}
}

func TestGamePacketRoundTrip(t *testing.T) {
	var all = constant.PacketGameTypeValues()
	for i := 0; i < len(all); i++ {
		if constant.PacketGameTypeOf(all[i].Id()) != all[i] {
			t.Errorf("PacketGameType.valueOf(%d): expected %d, got %d", all[i].Id(), int(all[i]), int(constant.PacketGameTypeOf(all[i].Id())))
		}
	}
	// Java returns null for an unknown id; the port uses the NONE sentinel.
	if constant.PacketGameTypeOf(47) != constant.PACKET_GAME_NONE {
		t.Errorf("PacketGameType.valueOf(47) must be the NONE sentinel")
	}
	if constant.PacketGameTypeOf(-1) != constant.PACKET_GAME_NONE {
		t.Errorf("PacketGameType.valueOf(-1) must be the NONE sentinel")
	}
}

// The game port has no offset, unlike the admin port's server packets, whose
// ids start at SERVER_PACKET_START_VALUE. A client packet and an admin packet
// can therefore share a wire id: the protocol, not the id, tells them apart.
func TestGameProtocolIsTheGameOne(t *testing.T) {
	var all = constant.PacketGameTypeValues()
	for i := 0; i < len(all); i++ {
		if all[i].Protocol() != constant.TCP_GAME {
			t.Errorf("%s protocol: expected TCP_GAME, got %d", constant.PacketGameTypeName(all[i]), int(all[i].Protocol()))
		}
	}
	// PACKET_CLIENT_JOIN (2) and ADMIN_PACKET_ADMIN_QUIT (1) plus the admin
	// offset do not collide, but PACKET_SERVER_ERROR (3) and
	// ADMIN_PACKET_ADMIN_POLL (3) do, which is only safe because they travel
	// on different sockets.
	if constant.PACKET_SERVER_ERROR.Id() != constant.ADMIN_PACKET_ADMIN_POLL.Id() {
		t.Errorf("the game and admin ports are expected to share id 3")
	}
	if constant.PACKET_SERVER_ERROR.Protocol() == constant.ADMIN_PACKET_ADMIN_POLL.Protocol() {
		t.Errorf("PACKET_SERVER_ERROR and ADMIN_PACKET_ADMIN_POLL must differ by protocol")
	}
}

// Both NetworkClient.handlePacket and NetworkClientSender.debugPacket log the
// constant's name, which the port can only do from a table.
func TestGamePacketNames(t *testing.T) {
	var expected = []string{
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
	var all = constant.PacketGameTypeValues()
	if len(expected) != len(all) {
		t.Errorf("expected a name for every PacketGameType, got %d names for %d constants", len(expected), len(all))
		return
	}
	for i := 0; i < len(all); i++ {
		if constant.PacketGameTypeName(all[i]) != expected[i] {
			t.Errorf("PacketGameType %d name: expected %s, got %s", i, expected[i], constant.PacketGameTypeName(all[i]))
		}
	}
	// An unknown id logs the sentinel's name rather than crashing.
	if constant.PacketGameTypeName(constant.PACKET_GAME_NONE) != "PACKET_END" {
		t.Errorf("expected the NONE sentinel to log PACKET_END, got %s", constant.PacketGameTypeName(constant.PACKET_GAME_NONE))
	}
}

func gameIds() []int {
	var ids []int
	var all = constant.PacketGameTypeValues()
	for i := 0; i < len(all); i++ {
		ids = append(ids, all[i].Id())
	}
	return ids
}
