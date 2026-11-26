// Every ADMIN_PACKET_ADMIN_* command, byte for byte, as the sender puts it on
// the wire. See test/java/com/openttd/network/admin/NetworkAdminSenderTest.java:
// payload layouts are copied from git show 15.3:src/network/core/tcp_admin.h
// ("Receive_ADMIN_*" doc comments) and the field length limits from
// git show 15.3:src/network/core/config.h.
//
// Two deliberate deviations from the Java test, both about a sender that has
// not received the PROTOCOL packet yet: Java's poll() dereferences the protocol
// field and throws NullPointerException, and updateFrequency() does the same.
// The port logs "The server does not support ..." and sends nothing, so those
// two cases are testable here instead of being unreachable, and they are asserted
// at the end of the file.
//
// Like the Java test, this drives everything through AdminHarness: the facade
// builds the sender, and NetworkAdmin decodes the PROTOCOL packet, so the sender
// under test has the protocol a real connection would have.

package admin_test

import "strconv"
import "testing"
import "com.openttd"
import "com.openttd/gamescript"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/server"

// newHarness is setUp(): a fresh connection per test, as Java's @BeforeEach
// builds one. nil means the harness could not connect, which no test can
// recover from.
func newHarness(t *testing.T) *server.AdminHarness {
	var harness, harnessError = server.NewAdminHarness()
	if harnessError != nil {
		t.Errorf("AdminHarness: %s", core.Message(harnessError))
		return nil
	}
	return harness
}

// ADMIN_PACKET_ADMIN_JOIN: string password, string name, string version.
func TestJoin(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	var join = srv.AwaitPacket(constant.ADMIN_PACKET_ADMIN_JOIN, 5)
	if join == nil {
		t.Errorf("no ADMIN_PACKET_ADMIN_JOIN within 5s")
		return
	}
	var payload = join.GetPacket()
	if payload.ReadString() != server.PASSWORD {
		t.Errorf("join: expected the harness password %s", server.PASSWORD)
	}
	if payload.ReadString() != server.ADMIN_NAME {
		t.Errorf("join: expected the harness name %s", server.ADMIN_NAME)
	}
	if payload.ReadString() != constant.OPENTTD {
		t.Errorf("join: expected the client version %s", constant.OPENTTD)
	}
	// NETWORK_PASSWORD_LENGTH / NETWORK_CLIENT_NAME_LENGTH / NETWORK_REVISION_LENGTH
	// are 33, 25 and 33 bytes including the terminator.
	if join.GetPacket().GetLength() > 3+33+25+33 {
		t.Errorf("join packet fits the field limits, was %d bytes", join.GetPacket().GetLength())
	}
}

// ADMIN_PACKET_ADMIN_QUIT: no payload.
func TestQuit(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	harness.GetSend().Quit()
	var quit = srv.AwaitPacket(constant.ADMIN_PACKET_ADMIN_QUIT, 5)
	if quit == nil {
		t.Errorf("no ADMIN_PACKET_ADMIN_QUIT within 5s")
		return
	}
	if quit.GetPacket().GetLength() != 3 {
		t.Errorf("quit packet is header only, was %d bytes", quit.GetPacket().GetLength())
	}
}

// assertSentFrequency waits for an ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY carrying
// this update type and this frequency mask.
func assertSentFrequency(t *testing.T, srv *server.FakeAdminServer, updateType constant.AdminUpdateType, expected constant.AdminUpdateFrequency) {
	openttd.Await(func() bool {
		var packets = srv.Packets(constant.ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY)
		for i := 0; i < len(packets); i++ {
			var payload = packets[i].GetPacket()
			if payload.ReadUint16() == int(updateType) && payload.ReadUint16() == expected.Mask() {
				return true
			}
		}
		return false
	}, 5, constant.AdminUpdateTypeName(updateType)+" sent with "+constant.AdminUpdateFrequencyName(expected))
}

// ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY: uint16 update type, uint16 frequency mask.
func TestUpdateFrequency(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	// COMPANY_ECONOMY advertises Poll|Weekly|..., so Weekly is legal here.
	harness.GetSend().UpdateFrequency(constant.ADMIN_UPDATE_COMPANY_ECONOMY, constant.ADMIN_FREQUENCY_WEEKLY)
	assertSentFrequency(t, srv, constant.ADMIN_UPDATE_COMPANY_ECONOMY, constant.ADMIN_FREQUENCY_WEEKLY)

	var packet = srv.AwaitPacket(constant.ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY, 5)
	if packet == nil {
		t.Errorf("no ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY within 5s")
		return
	}
	if packet.GetWireTypeId() != int(constant.ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY) {
		t.Errorf("command id: expected %d, got %d", int(constant.ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY), packet.GetWireTypeId())
	}
	if packet.GetPacket().GetLength() != 7 {
		t.Errorf("expected header + uint16 type + uint16 frequency = 7 bytes, got %d", packet.GetPacket().GetLength())
	}
}

// The server closes the connection on a frequency it does not advertise, so an
// unsupported request must fall back to an advertised one. POLL is not advertised
// for ADMIN_UPDATE_CHAT, which used to be the fallback.
func TestUpdateFrequencyFallback(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()
	var send = harness.GetSend()

	send.UpdateFrequency(constant.ADMIN_UPDATE_CHAT, constant.ADMIN_FREQUENCY_DAILY)
	assertSentFrequency(t, srv, constant.ADMIN_UPDATE_CHAT, constant.ADMIN_FREQUENCY_AUTOMATIC)

	send.UpdateFrequency(constant.ADMIN_UPDATE_GAMESCRIPT, constant.ADMIN_FREQUENCY_ANUALLY)
	assertSentFrequency(t, srv, constant.ADMIN_UPDATE_GAMESCRIPT, constant.ADMIN_FREQUENCY_AUTOMATIC)

	send.UpdateFrequency(constant.ADMIN_UPDATE_DATE, constant.ADMIN_FREQUENCY_AUTOMATIC)
	assertSentFrequency(t, srv, constant.ADMIN_UPDATE_DATE, constant.ADMIN_FREQUENCY_POLL)

	send.UpdateFrequency(constant.ADMIN_UPDATE_CMD_NAMES, constant.ADMIN_FREQUENCY_AUTOMATIC)
	assertSentFrequency(t, srv, constant.ADMIN_UPDATE_CMD_NAMES, constant.ADMIN_FREQUENCY_POLL)
}

// assertSentPoll waits for an ADMIN_PACKET_ADMIN_POLL carrying this update type
// and this id.
func assertSentPoll(t *testing.T, srv *server.FakeAdminServer, updateType constant.AdminUpdateType, data int64) {
	openttd.Await(func() bool {
		var packets = srv.Packets(constant.ADMIN_PACKET_ADMIN_POLL)
		for i := 0; i < len(packets); i++ {
			var payload = packets[i].GetPacket()
			if payload.ReadUint8() == int(updateType) && int64(payload.ReadUint32()) == data {
				return true
			}
		}
		return false
	}, 5, "poll "+constant.AdminUpdateTypeName(updateType)+" "+strconv.FormatInt(data, 10))
}

// ADMIN_PACKET_ADMIN_POLL: uint8 update type, uint32 id.
func TestPollDate(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()

	harness.GetSend().PollDate()
	assertSentPoll(t, harness.GetServer(), constant.ADMIN_UPDATE_DATE, int64(0))
}

// UINT32_MAX asks for every client.
func TestPollAllClients(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()

	harness.GetSend().PollClientInfo(-1)
	assertSentPoll(t, harness.GetServer(), constant.ADMIN_UPDATE_CLIENT_INFO, 4294967295)
}

func TestPollOneClient(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()

	harness.GetSend().PollClientInfo(2)
	assertSentPoll(t, harness.GetServer(), constant.ADMIN_UPDATE_CLIENT_INFO, 2)
}

// UINT32_MAX asks for every company.
func TestPollAllCompanies(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()

	harness.GetSend().PollCompanyInfo(-1)
	assertSentPoll(t, harness.GetServer(), constant.ADMIN_UPDATE_COMPANY_INFO, 4294967295)
}

func TestPollCompanyEconomy(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()

	harness.GetSend().PollCompanyEconomy(1)
	assertSentPoll(t, harness.GetServer(), constant.ADMIN_UPDATE_COMPANY_ECONOMY, 1)
}

func TestPollCompanyStats(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()

	harness.GetSend().PollCompanyStats(0)
	assertSentPoll(t, harness.GetServer(), constant.ADMIN_UPDATE_COMPANY_STATS, 0)
}

func TestPollCmdNames(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()

	harness.GetSend().PollCmdNames()
	assertSentPoll(t, harness.GetServer(), constant.ADMIN_UPDATE_CMD_NAMES, 0)
}

// CHAT, CONSOLE, CMD_LOGGING and GAMESCRIPT have no Poll bit: polling them makes
// the server answer NETWORK_ERROR_ILLEGAL_PACKET and close the connection, so
// nothing must be sent.
func TestPollUnsupportedType(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()
	var send = harness.GetSend()

	var types = []constant.AdminUpdateType{
		constant.ADMIN_UPDATE_CHAT,
		constant.ADMIN_UPDATE_CONSOLE,
		constant.ADMIN_UPDATE_CMD_LOGGING,
		constant.ADMIN_UPDATE_GAMESCRIPT,
	}
	for i := 0; i < len(types); i++ {
		send.Poll(types[i], int64(0))
	}
	openttd.Wait(1)

	var packets = srv.Packets(constant.ADMIN_PACKET_ADMIN_POLL)
	for i := 0; i < len(packets); i++ {
		var sentType = packets[i].GetPacket().ReadUint8()
		if sentType >= int(constant.ADMIN_UPDATE_END) {
			t.Errorf("poll type %d is out of range", sentType)
		}
		for j := 0; j < len(types); j++ {
			if sentType == int(types[j]) {
				t.Errorf("%s is not pollable, but was polled", constant.AdminUpdateTypeName(types[j]))
			}
		}
	}
}

// assertSentChat waits for an ADMIN_PACKET_ADMIN_CHAT with this action,
// destination type, destination id and message.
func assertSentChat(t *testing.T, srv *server.FakeAdminServer, action constant.NetworkAction, destType constant.DestType, dest int64, message string) {
	openttd.Await(func() bool {
		var packets = srv.Packets(constant.ADMIN_PACKET_ADMIN_CHAT)
		for i := 0; i < len(packets); i++ {
			var payload = packets[i].GetPacket()
			if payload.ReadUint8() == int(action) {
				if payload.ReadUint8() == int(destType) {
					if int64(payload.ReadUint32()) == dest && payload.ReadString() == message {
						return true
					}
				}
			}
		}
		return false
	}, 5, "chat "+constant.NetworkActionName(action)+" "+constant.DestTypeName(destType)+" "+strconv.FormatInt(dest, 10)+" \""+message+"\"")
}

// ADMIN_PACKET_ADMIN_CHAT: uint8 action, uint8 desttype, uint32 dest, string.
func TestChatBroadcast(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()

	harness.GetSend().ChatBroadcast("hello all")
	assertSentChat(t, harness.GetServer(), constant.NETWORK_ACTION_SERVER_MESSAGE, constant.DESTTYPE_BROADCAST, 0, "hello all")
}

func TestChatCompany(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()

	harness.GetSend().ChatCompany(3, "team message")
	assertSentChat(t, harness.GetServer(), constant.NETWORK_ACTION_SERVER_MESSAGE, constant.DESTTYPE_TEAM, 3, "team message")
}

func TestChatClient(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()

	harness.GetSend().ChatClient(5, "private message")
	assertSentChat(t, harness.GetServer(), constant.NETWORK_ACTION_SERVER_MESSAGE, constant.DESTTYPE_CLIENT, 5, "private message")
}

// A negative client id falls back to a broadcast.
func TestChatClientNegativeFallsBackToBroadcast(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()

	harness.GetSend().ChatClient(-1, "to everyone")
	assertSentChat(t, harness.GetServer(), constant.NETWORK_ACTION_SERVER_MESSAGE, constant.DESTTYPE_BROADCAST, 0, "to everyone")
}

// Chat with UTF-8, matching NETWORK_CHAT_LENGTH (900 bytes including '\0').
func TestChatUtf8(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()

	var message = "こんにちは ☕"
	harness.GetSend().ChatBroadcast(message)
	assertSentChat(t, harness.GetServer(), constant.NETWORK_ACTION_SERVER_MESSAGE, constant.DESTTYPE_BROADCAST, 0, message)
}

// ADMIN_PACKET_ADMIN_EXTERNAL_CHAT: string source, uint16 colour, string user,
// string message.
func TestExternalChat(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	harness.GetSend().ExternalChat("strategyboard", 2, "Alice", "a message from outside")
	var packet = srv.AwaitPacket(constant.ADMIN_PACKET_ADMIN_EXTERNAL_CHAT, 5)
	if packet == nil {
		t.Errorf("no ADMIN_PACKET_ADMIN_EXTERNAL_CHAT within 5s")
		return
	}
	var payload = packet.GetPacket()
	if payload.ReadString() != "strategyboard" {
		t.Errorf("external chat source: %s", payload.ReadString())
	}
	if payload.ReadUint16() != 2 {
		t.Errorf("external chat colour: expected 2, got %d", payload.ReadUint16())
	}
	if payload.ReadString() != "Alice" {
		t.Errorf("external chat user: %s", payload.ReadString())
	}
	if payload.ReadString() != "a message from outside" {
		t.Errorf("external chat message: %s", payload.ReadString())
	}
}

// ADMIN_PACKET_ADMIN_RCON: string command.
func TestRcon(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	harness.GetSend().Rcon("getdate")
	var packet = srv.AwaitPacket(constant.ADMIN_PACKET_ADMIN_RCON, 5)
	if packet == nil {
		t.Errorf("no ADMIN_PACKET_ADMIN_RCON within 5s")
		return
	}
	if packet.GetPacket().ReadString() != "getdate" {
		t.Errorf("rcon command: %s", packet.GetPacket().ReadString())
	}
}

// NETWORK_RCONCOMMAND_LENGTH is 500 bytes including '\0': a longer command makes
// the server reject the packet, so callers must keep commands short; the sender
// does not cap them.
func TestRconCommandLength(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	var longCommand = "say "
	for (len(longCommand) < 500) {
		longCommand = longCommand + "x"
	}
	harness.GetSend().Rcon(longCommand)
	var packet = srv.AwaitPacket(constant.ADMIN_PACKET_ADMIN_RCON, 5)
	if packet == nil {
		t.Errorf("no ADMIN_PACKET_ADMIN_RCON within 5s")
		return
	}
	if packet.GetPacket().ReadString() != longCommand {
		t.Errorf("rcon command of %d bytes did not round trip", len(longCommand))
	}
	if packet.GetPacket().GetLength() <= 3+500 {
		t.Errorf("the sender does not cap commands at NETWORK_RCONCOMMAND_LENGTH, got %d bytes", packet.GetPacket().GetLength())
	}
}

// ADMIN_PACKET_ADMIN_GAMESCRIPT: string json.
func TestGameScript(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()
	var json = "{\"c\":1,\"a\":[14,\"hello\",\"-1\"],\"i\":\"0\"}"

	if !harness.GetSend().GameScript(json) {
		t.Errorf("gameScript refused a payload that fits the packet")
	}
	var packet = srv.AwaitPacket(constant.ADMIN_PACKET_ADMIN_GAMESCRIPT, 5)
	if packet == nil {
		t.Errorf("no ADMIN_PACKET_ADMIN_GAMESCRIPT within 5s")
		return
	}
	if packet.GetPacket().ReadString() != json {
		t.Errorf("gamescript json: %s", packet.GetPacket().ReadString())
	}
}

// gamescript json over COMPAT_MTU is refused and not sent. Java passes null to
// the first branch; the empty string is Go's stand-in for it.
func TestGameScriptTooLong(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()
	var send = harness.GetSend()

	if send.GameScript("") {
		t.Errorf("gameScript(\"\") must be refused")
	}
	var json = "{\"a\":\""
	for (len(json) <= core.MTU-3) {
		json = json + "x"
	}
	json = json + "\"}"
	if send.GameScript(json) {
		t.Errorf("gameScript accepted %d bytes of json", len(json))
	}
	openttd.Wait(1)
	if len(srv.Packets(constant.ADMIN_PACKET_ADMIN_GAMESCRIPT)) != 0 {
		t.Errorf("nothing must reach the server, %d gamescript packets arrived",
			len(srv.Packets(constant.ADMIN_PACKET_ADMIN_GAMESCRIPT)))
	}
}

// ADMIN_PACKET_ADMIN_PING: uint32 identifier.
func TestPing(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	harness.GetSend().Ping(4294967295)
	var packet = srv.AwaitPacket(constant.ADMIN_PACKET_ADMIN_PING, 5)
	if packet == nil {
		t.Errorf("no ADMIN_PACKET_ADMIN_PING within 5s")
		return
	}
	if int64(packet.GetPacket().ReadUint32()) != 4294967295 {
		t.Errorf("ping id: expected %d, got %d", 4294967295, int64(packet.GetPacket().ReadUint32()))
	}
}

// findGameScriptJson is awaitGameScriptJson: the payload of the first
// ADMIN_PACKET_ADMIN_GAMESCRIPT whose "c" is this command, read with RawFields
// so every value keeps the text the gamescript wrote (gson's JsonNumber does the
// same, which is why a big id must not come back as 5e10).
func findGameScriptJson(srv *server.FakeAdminServer, command constant.GSCommand) map[string]string {
	var packets = srv.Packets(constant.ADMIN_PACKET_ADMIN_GAMESCRIPT)
	for i := 0; i < len(packets); i++ {
		var json = gamescript.RawFields(packets[i].GetPacket().ReadString())
		if json != nil {
			var cmd, hasCmd = json[constant.GS_CMD]
			if hasCmd && cmd == strconv.Itoa(int(command)) {
				return json
			}
		}
	}
	return nil
}

// awaitGameScriptJson waits for the command, then returns its fields.
func awaitGameScriptJson(t *testing.T, srv *server.FakeAdminServer, command constant.GSCommand) map[string]string {
	openttd.Await(func() bool { return findGameScriptJson(srv, command) != nil }, 5,
		"gamescript command "+constant.GSCommandName(command))
	return findGameScriptJson(srv, command)
}

// Goal commands go through GSExecutor, so they carry a request id: AdminCmd only
// runs requests with "i", "c" and "a".
func TestAddCompanyGoal(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	harness.GetSend().AddCompanyGoal(1, "deliver cargo", constant.GOAL_TOWN, 4)
	var json = awaitGameScriptJson(t, srv, constant.GS_COMMAND_ADD_GOAL)
	var cmd, hasCmd = json[constant.GS_CMD]
	if !hasCmd || cmd != strconv.Itoa(int(constant.GS_COMMAND_ADD_GOAL)) {
		t.Errorf("expected \"c\" = %d", int(constant.GS_COMMAND_ADD_GOAL))
	}
	var id, hasId = json[constant.GS_ID]
	if !hasId {
		t.Errorf("AdminCmd requires \"i\"")
	}
	// The harness's executor.clear() restarts the numbering, so this is the
	// first request of the test.
	if id != "0" {
		t.Errorf("request id: expected 0, got %s", id)
	}
	var args, hasArgs = json[constant.GS_ARGS]
	if !hasArgs {
		t.Errorf("expected an \"a\" array")
		return
	}
	var items = gamescript.JsonListItems(args)
	if len(items) != 4 {
		t.Errorf("expected 4 arguments, got %d", len(items))
		return
	}
	if items[0] != "1" {
		t.Errorf("argument 0: expected company 1, got %s", items[0])
	}
	if items[1] != "deliver cargo" {
		t.Errorf("argument 1: expected the goal text, got %s", items[1])
	}
	if items[2] != strconv.Itoa(int(constant.GOAL_TOWN)) {
		t.Errorf("argument 2: expected goal type %d, got %s", int(constant.GOAL_TOWN), items[2])
	}
	if items[3] != "4" {
		t.Errorf("argument 3: expected destination 4, got %s", items[3])
	}
}

func TestAddGlobalGoal(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	harness.GetSend().AddGlobalGoal("global goal", constant.GOAL_INDUSTRY, 7)
	var json = awaitGameScriptJson(t, srv, constant.GS_COMMAND_ADD_GOAL)
	var args, hasArgs = json[constant.GS_ARGS]
	if !hasArgs {
		t.Errorf("expected an \"a\" array")
		return
	}
	var items = gamescript.JsonListItems(args)
	if len(items) != 4 {
		t.Errorf("expected 4 arguments, got %d", len(items))
		return
	}
	if items[0] != "-1" {
		t.Errorf("a global goal is company -1, got %s", items[0])
	}
	if items[2] != strconv.Itoa(int(constant.GOAL_INDUSTRY)) {
		t.Errorf("argument 2: expected goal type %d, got %s", int(constant.GOAL_INDUSTRY), items[2])
	}
}

func TestRemoveGoal(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	harness.GetSend().RemoveGoal(12)
	var json = awaitGameScriptJson(t, srv, constant.GS_COMMAND_REMOVE_GOAL)
	var args, hasArgs = json[constant.GS_ARGS]
	if !hasArgs {
		t.Errorf("expected an \"a\" array")
		return
	}
	var items = gamescript.JsonListItems(args)
	if len(items) != 1 || items[0] != "12" {
		t.Errorf("expected the goal id 12, got %d arguments", len(items))
	}
	var id, hasId = json[constant.GS_ID]
	if !hasId {
		t.Errorf("AdminCmd requires \"i\", got %s", id)
	}
}

func TestRemoveAllGoal(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	harness.GetSend().RemoveAllGoal()
	var json = awaitGameScriptJson(t, srv, constant.GS_COMMAND_REMOVE_ALL_GOAL)
	var args, hasArgs = json[constant.GS_ARGS]
	if !hasArgs {
		t.Errorf("expected an empty \"a\" array")
		return
	}
	if len(gamescript.JsonListItems(args)) != 0 {
		t.Errorf("removeAllGoal takes no arguments")
	}
	var id, hasId = json[constant.GS_ID]
	if !hasId {
		t.Errorf("AdminCmd requires \"i\", got %s", id)
	}
}

func TestSetTownCargoGoal(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	harness.GetSend().SetTownCargoGoal(3, constant.TOWN_EFFECT_MAIL, 500)
	var json = awaitGameScriptJson(t, srv, constant.GS_COMMAND_SET_TOWN_CARGO_GOAL)
	var args, hasArgs = json[constant.GS_ARGS]
	if !hasArgs {
		t.Errorf("expected an \"a\" array")
		return
	}
	var items = gamescript.JsonListItems(args)
	if len(items) != 3 {
		t.Errorf("expected 3 arguments, got %d", len(items))
		return
	}
	if items[0] != "3" {
		t.Errorf("argument 0: expected town 3, got %s", items[0])
	}
	if items[1] != strconv.Itoa(int(constant.TOWN_EFFECT_MAIL)) {
		t.Errorf("argument 1: expected town effect %d, got %s", int(constant.TOWN_EFFECT_MAIL), items[1])
	}
	if items[2] != "500" {
		t.Errorf("argument 2: expected goal 500, got %s", items[2])
	}
	var id, hasId = json[constant.GS_ID]
	if !hasId {
		t.Errorf("AdminCmd requires \"i\", got %s", id)
	}
}

// settle waits for the sender's queue to drain after the harness's WELCOME
// handler queued its own UPDATE_FREQUENCY and POLL packets, so the two tests
// below can measure a delta instead of racing that batch.
func settle(harness *server.AdminHarness) {
	var send = harness.GetSend()
	openttd.Await(func() bool { return send.GetQueue().Size() == 0 }, 10, "the WELCOME packets to be sent")
	openttd.Wait(1)
}

// Deviation: Java's updateFrequency() throws NullPointerException when the
// protocol is unknown; the port logs the unsupported-frequency error and
// sends nothing, which the server would reject anyway. The harness has already
// answered WELCOME, which queues its own UPDATE_FREQUENCY packets, so the
// assertion is a delta, not a zero.
func TestUpdateFrequencyWithoutProtocol(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()
	var send = harness.GetSend()

	settle(harness)
	var before = len(srv.Packets(constant.ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY))
	send.SetProtocol(nil)
	send.UpdateFrequency(constant.ADMIN_UPDATE_DATE, constant.ADMIN_FREQUENCY_POLL)
	openttd.Wait(1)
	if len(srv.Packets(constant.ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY)) != before {
		t.Errorf("an unsupported frequency must send nothing")
	}
}

// Deviation: same for poll(), where Java's caller would have seen a NullPointerException
// on its own thread. Again a delta: WELCOME's PollDate/PollClientInfo/... calls
// already put POLL packets on the wire.
func TestPollWithoutProtocol(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()
	var send = harness.GetSend()

	settle(harness)
	var before = len(srv.Packets(constant.ADMIN_PACKET_ADMIN_POLL))
	send.SetProtocol(nil)
	send.Poll(constant.ADMIN_UPDATE_DATE, int64(0))
	openttd.Wait(1)
	if len(srv.Packets(constant.ADMIN_PACKET_ADMIN_POLL)) != before {
		t.Errorf("polling an unsupported update must send nothing")
	}
}
