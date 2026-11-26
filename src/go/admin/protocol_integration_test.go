// The admin protocol (port 3977) against a real OpenTTD server, ported from
// test/java/com/openttd/admin/OpenttdAdminProtocolIntegrationTest.java.
//
// The offline suite (server.AdminHarness) checks what the listener decodes and what
// the sender encodes; this suite checks that the real server accepts the encoding and
// that the packets it sends back end up in the game model and as functional events.
// Expected behaviour is taken from `git show 15.3:src/network/network_admin.cpp`,
// `network_server.cpp` and `console_cmds.cpp`.
//
// The dedicated server has no clients and no companies (no game-port client joins
// here), so client and company assertions stay modest: the server's own client (#1)
// is always there, and company polls must simply be accepted.
//
// Like integration_test.go, every test skips when nothing listens on 3977; the Go
// ladder never spawns the server itself. Java's nextPingId field becomes the package
// var below: `go test` runs the tests of a package sequentially, so the ids stay
// distinct the way they are in JUnit, where one instance is built per test.
//
// The dedicated server has no Java test either, so this file is new coverage for the
// Go port rather than a port of a test the Java tree already had for these classes.

package admin_test

import "strconv"
import "strings"
import "testing"
import "time"
import "com.openttd"
import "com.openttd/gamescript"
import "com.openttd/admin"
import "com.openttd/admin/event"
import netadmin "com.openttd/network/admin"
import "com.openttd/network/constant"
import rule "com.openttd/robot/rule"
import "com.openttd/server"

// Colour constants from git show 15.3:src/console_type.h.
const (
	CC_DEFAULT = 0x01 // TC_SILVER
	CC_ERROR   = 0x03 // TC_RED
	CC_INFO    = 0x08 // TC_YELLOW, also a colour an external chat may use
	TC_INVALID = 0xFF // TC_INVALID, rejected by IsValidConsoleColour
)

// protocolPingId is Java's nextPingId: assertConnectionUsable() takes one id per call.
var protocolPingId int64

/* The tests of OpenttdAdminProtocolIntegrationTest.java */

// TestJoinAndWelcome is testJoinAndWelcome(): ADMIN_PACKET_ADMIN_JOIN answers with
// PROTOCOL and WELCOME, decoded into GameInfo.
func TestJoinAndWelcome(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	var gameInfo = fixture.admin.GetGame().GetGameInfo()

	if !fixture.admin.IsConnected() {
		t.Error("the admin connection is open")
		return
	}
	if gameInfo.GetServerName() == "" {
		t.Error("server name is not empty")
		return
	}
	if !stringsContains(gameInfo.GetServerRevision(), constant.OPENTTD) {
		t.Errorf("revision %s must match the advertised version", gameInfo.GetServerRevision())
	}
	if !gameInfo.IsServerDedicated() {
		t.Error("the test server is dedicated")
	}
	if gameInfo.GetVersion() != server.GAME_ADMIN_VERSION {
		t.Errorf("protocol version: want %d, got %d", server.GAME_ADMIN_VERSION, gameInfo.GetVersion())
	}

	// openttd/openttd.cfg: map_x = map_y = 8, i.e. 64x64 tiles.
	if gameInfo.GetMapWidth() != gameInfo.GetMapHeight() {
		t.Errorf("square map: %d x %d", gameInfo.GetMapWidth(), gameInfo.GetMapHeight())
	}
	if gameInfo.GetMapWidth() <= 0 {
		t.Error("map size")
	}
	if gameInfo.GetMapSeed() <= 0 {
		t.Error("generation seed")
	}

	// [game_creation] starting_year = 1950.
	var startYear = server.YearOf(int(gameInfo.GetStartDate()))
	if startYear != 1950 {
		t.Errorf("start year: want 1950, got %d", startYear)
	}

	// WELCOME carries only the map's start date; the current date comes with the
	// ADMIN_PACKET_SERVER_DATE the WELCOME handler polls for. Game.UpdateAdmin swaps
	// the network model on every packet, so the game info has to be read again rather
	// than kept in a reference.
	awaitTrue(t, func() bool {
		var latest = fixture.admin.GetGame().GetGameInfo()
		return latest.GetCurrentDate() >= latest.GetStartDate()
	}, 10, "the current game date")
}

// TestWrongPasswordIsNeverWelcomed is testWrongPassword(): Receive_ADMIN_JOIN compares
// the password and answers NETWORK_ERROR_WRONG_PASSWORD.
func TestWrongPasswordIsNeverWelcomed(t *testing.T) {
	var rejected = integrationAdmin(t, "not-the-admin-password", "openttd-admin wrong password test")
	if rejected == nil {
		return
	}

	// Ping through the reconnect window: every connection is rejected, so neither a
	// WELCOME nor a PONG can ever arrive.
	for attempt := 0; attempt < 6; attempt++ {
		var send = rejected.admin.GetSend()
		if send != nil {
			send.Ping(9001)
		}
		openttd.Wait(1)
	}
	if len(rejected.recorder.GetPongEvents()) != 0 {
		t.Errorf("no PONG on a rejected connection: %s", describePongs(rejected.recorder.GetPongEvents()))
	}
	if len(rejected.recorder.GetGameScriptEvents()) != 0 {
		t.Errorf("no gamescript traffic: %s", describeGameScripts(rejected.recorder.GetGameScriptEvents()))
	}
	if welcomeReceived(rejected.admin) {
		t.Error("no WELCOME on a rejected connection")
	}
}

// TestUpdateFrequencyCombinations is testUpdateFrequencyCombinations():
// Receive_ADMIN_UPDATE_FREQUENCY closes the connection on a frequency the server does
// not advertise, so every combination sent must be one the PROTOCOL packet promised,
// including the ones Game.UpdateAdmin registers.
func TestUpdateFrequencyCombinations(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	var send = fixture.admin.GetSend()
	if send == nil {
		t.Fatal("a sender once connected")
		return
	}

	var types = server.FrequencyTypes()
	var masks = server.Frequencies()
	for i := 0; i < len(types); i++ {
		var advertised = masks[i]
		for _, frequency := range constant.AdminUpdateFrequencyValues() {
			if advertised&frequency.Mask() == 0 {
				continue
			}
			send.UpdateFrequency(types[i], frequency)
		}
	}
	assertConnectionUsable(t, fixture)
}

// TestUpdateFrequencyFallback is testUpdateFrequencyFallback(): ADMIN_UPDATE_COMPANY_INFO
// only advertises Poll|Automatic, so the sender falls back to an advertised frequency
// instead of letting the server drop the connection - what keeps strategyboard's
// objectives running.
func TestUpdateFrequencyFallback(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	var send = fixture.admin.GetSend()
	if send == nil {
		t.Fatal("a sender once connected")
		return
	}
	send.UpdateFrequency(constant.ADMIN_UPDATE_COMPANY_INFO, constant.ADMIN_FREQUENCY_ANUALLY)
	assertConnectionUsable(t, fixture)
}

// TestPolls is testPolls(): Receive_ADMIN_POLL answers each pollable type.
func TestPolls(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	var send = fixture.admin.GetSend()
	if send == nil {
		t.Fatal("a sender once connected")
		return
	}
	fixture.recorder.Clear()

	send.PollDate()
	if !awaitSizeTrue(t, func() int { return len(fixture.recorder.GetDateEvents()) }, 1, 10,
		"a DateEvent from PollDate()") {
		return
	}

	send.PollClientInfo(-1) // UINT32_MAX: every client, including the server's own
	// git show 15.3:src/network/network_type.h: CLIENT_ID_SERVER = 1, so a dedicated
	// server has exactly one client info, and id 0 is INVALID_CLIENT_ID.
	if !awaitTrue(t, func() bool {
		var client = fixture.admin.GetGame().GetClient(1)
		return client != nil && client.GetName() != ""
	}, 10, "the server client #1") {
		return
	}
	var client = fixture.admin.GetGame().GetClient(1)
	if client.GetId() != 1 {
		t.Errorf("client #1: got id %d", client.GetId())
	}
	if client.GetName() == "" {
		t.Error("client name")
	}
	if client.GetCompany() != 255 {
		t.Errorf("the server plays as spectator: got company %d", client.GetCompany())
	}

	send.PollClientInfo(1)
	if !awaitSizeTrue(t, func() int { return len(fixture.recorder.GetClientEvents()) }, 2, 10,
		"a ClientEvent for client #1") {
		return
	}

	// No companies on a dedicated server with no clients: the polls are accepted, and
	// nothing is added to the model.
	send.PollCompanyInfo(-1)
	send.PollCompanyEconomy(-1)
	send.PollCompanyStats(-1)
	assertConnectionUsable(t, fixture)
	if len(fixture.admin.GetGame().GetCompanies()) != 0 {
		t.Errorf("no companies on the dedicated server: %s", describeCompanies(fixture.admin.GetGame().GetCompanies()))
	}

	send.PollCmdNames()
	if !awaitTrue(t, func() bool { return len(fixture.admin.GetGame().GetCommandNames()) > 0 }, 10,
		"command names") {
		return
	}
	if len(fixture.admin.GetGame().GetCommandNames()) <= 10 {
		t.Errorf("a command name list: got %d names", len(fixture.admin.GetGame().GetCommandNames()))
	}
}

// TestChatBroadcastAndClient is testChatBroadcastAndClient(): Receive_ADMIN_CHAT calls
// NetworkServerSendChat(..., from_admin = true), and NetworkAdminChat returns immediately
// for admin messages, so the admin gets no SERVER_CHAT echo of its own message - only the
// console relay.
func TestChatBroadcastAndClient(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	var send = fixture.admin.GetSend()
	if send == nil {
		t.Fatal("a sender once connected")
		return
	}
	var broadcast = "integration broadcast " + strconv.FormatInt(int64TestNow(), 10)
	var toServerClient = "integration client chat " + strconv.FormatInt(int64TestNow(), 10)

	send.ChatBroadcast(broadcast)
	awaitConsoleMessage(t, fixture, broadcast)

	// DESTTYPE_CLIENT with dest == CLIENT_ID_SERVER (1) prints the message on the
	// server's console; dest 0 is INVALID_CLIENT_ID and prints nothing.
	send.ChatClient(1, toServerClient)
	awaitConsoleMessage(t, fixture, toServerClient)

	if len(fixture.recorder.GetChatEvents()) != 0 {
		t.Errorf("an admin's own chat must not come back as a ChatEvent: %s",
			describeChats(fixture.recorder.GetChatEvents()))
	}

	send.ChatClient(0, "chat to an invalid client id "+strconv.FormatInt(int64TestNow(), 10))
	assertConnectionUsable(t, fixture)
}

// TestChatCompanyIsAccepted is testChatCompany(): team chat with no members has nothing
// to print, but the packet is legal.
func TestChatCompanyIsAccepted(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	var send = fixture.admin.GetSend()
	if send == nil {
		t.Fatal("a sender once connected")
		return
	}
	var team = "integration team chat " + strconv.FormatInt(int64TestNow(), 10)

	send.ChatCompany(0, team)
	assertConnectionUsable(t, fixture)
	if len(fixture.recorder.GetChatEvents()) != 0 {
		t.Errorf("no ChatEvent for admin chat: %s", describeChats(fixture.recorder.GetChatEvents()))
	}
	if consoleContains(fixture, team) {
		t.Errorf("no client in company 0, so nothing is printed: %s", team)
	}
}

// TestChatUtf8Is testChatUtf8(): chat text is relayed verbatim, including non-ASCII bytes.
func TestChatUtf8(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	var message = "unicode: héllo ütf8 — 100 €"
	fixture.admin.GetSend().ChatBroadcast(message)
	awaitConsoleMessage(t, fixture, message)
}

// TestExternalChatValidColour is testExternalChatValidColour(): Receive_ADMIN_EXTERNAL_CHAT
// requires IsValidConsoleColour.
func TestExternalChatValidColour(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	var message = "external chat " + strconv.FormatInt(int64TestNow(), 10)

	fixture.admin.GetSend().ExternalChat("strategyboard", CC_INFO, "bot", message)
	awaitConsoleMessage(t, fixture, message)
	assertConnectionUsable(t, fixture)
}

// TestExternalChatInvalidColour is testExternalChatInvalidColour(): an invalid colour makes
// the server send NETWORK_ERROR_ILLEGAL_PACKET and close.
func TestExternalChatInvalidColour(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	var message = "invalid colour chat " + strconv.FormatInt(int64TestNow(), 10)

	fixture.admin.GetSend().ExternalChat("strategyboard", TC_INVALID, "bot", message)
	if consoleContains(fixture, message) {
		t.Errorf("an illegal packet is not printed: %s", message)
	}
	// The listener reconnects every 5 s, so the admin becomes usable again.
	assertConnectionUsable(t, fixture)
}

// int64TestNow is System.nanoTime(): a unique suffix for chat text.
func int64TestNow() int64 {
	return time.Now().UnixNano()
}

// stringsContains is Java's String.contains().
func stringsContains(text string, needle string) bool {
	return strings.Contains(text, needle)
}

/* Helpers */

// assertConnectionUsable is assertConnectionUsable().
func assertConnectionUsable(t *testing.T, fixture *IntegrationFixture) {
	protocolPingId = protocolPingId + 1
	assertUsable(t, fixture.admin, fixture.recorder, protocolPingId, 20)
}

// assertUsable is assertUsable(): ping until a PongEvent with pingId arrives. The sender is
// recreated on every reconnect, so a ping sent while the connection is down is dropped, hence
// pinging inside the awaited condition.
func assertUsable(t *testing.T, adminInstance *admin.OpenttdAdmin, recorder *server.EventRecorder,
	pingId int64, timeoutSeconds int) {
	awaitTrue(t, func() bool {
		var send = adminInstance.GetSend()
		if send == nil {
			return false
		}
		send.Ping(pingId)
		for _, pongEvent := range recorder.GetPongEvents() {
			if pongEvent.GetPingId() == pingId {
				return true
			}
		}
		return false
	}, timeoutSeconds, "a PONG for ping "+strconv.FormatInt(pingId, 10))
}

// assertRCon is assertRCon(): the command's output must carry the expected text and colour.
func assertRCon(t *testing.T, fixture *IntegrationFixture, command string, expectedText string,
	expectedColour int) {
	var send = fixture.admin.GetSend()
	if send == nil {
		t.Fatal("a sender once connected")
		return
	}
	fixture.recorder.Clear()

	send.Rcon(command)
	if !awaitTrue(t, func() bool { return len(fixture.recorder.GetRConEndEvents()) > 0 }, 10,
		"an RConEndEvent for "+command) {
		return
	}
	var ends = fixture.recorder.GetRConEndEvents()
	if ends[0].GetCommand() != command {
		t.Errorf("command echo: want %s, got %s", command, ends[0].GetCommand())
	}

	var replies = fixture.recorder.GetRConEvents()
	var hasText = false
	var hasColour = false
	for _, rConEvent := range replies {
		if stringsContains(rConEvent.GetMessage(), expectedText) {
			hasText = true
		}
		if rConEvent.GetColor() == expectedColour {
			hasColour = true
		}
	}
	if !hasText {
		t.Errorf("%s must reply '%s', got %s", command, expectedText, describeRCons(replies))
	}
	if !hasColour {
		t.Errorf("%s must reply in colour %d, got %s", command, expectedColour, describeRCons(replies))
	}
	assertConnectionUsable(t, fixture)
}

func describeRCons(replies []*event.RConEvent) string {
	var text = ""
	for _, rConEvent := range replies {
		text = text + "colour " + strconv.Itoa(rConEvent.GetColor()) + " '" + rConEvent.GetMessage() + "'\n"
	}
	return text
}

// The event lists Java interpolates through toString() have no String method in Go,
// so each one is spelled out through its own ToString().
func describePongs(pongs []*event.PongEvent) string {
	var text = ""
	for _, pongEvent := range pongs {
		text = text + pongEvent.ToString() + "\n"
	}
	return text
}

func describeGameScripts(gameScripts []*event.GameScriptEvent) string {
	var text = ""
	for _, gameScriptEvent := range gameScripts {
		text = text + gameScriptEvent.ToString() + "\n"
	}
	return text
}

func describeChats(chats []*event.ChatEvent) string {
	var text = ""
	for _, chatEvent := range chats {
		text = text + chatEvent.ToString() + "\n"
	}
	return text
}

func describeCompanies(companies []*netadmin.Company) string {
	var text = ""
	for _, company := range companies {
		text = text + company.ToString() + "\n"
	}
	return text
}

// awaitConsoleMessage is awaitConsoleMessage().
func awaitConsoleMessage(t *testing.T, fixture *IntegrationFixture, text string) {
	awaitTrue(t, func() bool { return consoleContains(fixture, text) }, 10,
		"a ConsoleEvent containing '"+text+"'")
}

// consoleContains is consoleContains().
func consoleContains(fixture *IntegrationFixture, text string) bool {
	for _, consoleEvent := range fixture.recorder.GetConsoleEvents() {
		if stringsContains(consoleEvent.GetMessage(), text) {
			return true
		}
	}
	return false
}

// awaitGameScriptData is awaitGameScriptData(): the reply of a gamescript command issued
// after the recorder was cleared. It returns the reply's "data" field, and fails the test
// where Java throws an AssertionError.
func awaitGameScriptData(t *testing.T, fixture *IntegrationFixture, expectedReplyNumber int) string {
	if !awaitSizeTrue(t, func() int { return len(fixture.recorder.GetGameScriptEvents()) }, expectedReplyNumber,
		15, "gamescript reply #"+strconv.Itoa(expectedReplyNumber)) {
		return ""
	}
	var reply = gamescript.Parse(fixture.recorder.GetGameScriptEvents()[expectedReplyNumber-1].GetMessage())
	if reply == nil {
		t.Fatalf("no data or exception in reply #%d", expectedReplyNumber)
		return ""
	}
	if reply.HasData {
		return reply.Data
	}
	if reply.HasException {
		t.Fatalf("gamescript replied an exception: %s", reply.Exception)
		return ""
	}
	t.Fatalf("no data or exception in reply #%d: %s", expectedReplyNumber,
		fixture.recorder.GetGameScriptEvents()[expectedReplyNumber-1].GetMessage())
	return ""
}

// TestRConCommands is testRConCommands(): Receive_ADMIN_RCON redirects the console to
// ADMIN_PACKET_SERVER_RCON, so replies carry colour and text and RCON_END echoes the command.
func TestRConCommands(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	assertRCon(t, fixture, "getdate", "Date: 19", CC_DEFAULT)
	assertRCon(t, fixture, "getseed", "Generation Seed:", CC_DEFAULT)
	// NetworkPrintClients prints the server's own client #1, IP "server".
	assertRCon(t, fixture, "clients", "Client #1", CC_INFO)
	assertRCon(t, fixture, "clients", "IP: server", CC_INFO)
	// list_cmds prints one line per command; getdate is one of them.
	assertRCon(t, fixture, "list_cmds", "getdate", CC_DEFAULT)
}

// TestRConWithoutOutput is testRConWithoutOutput(): `companies` lists nothing when there are
// no companies, but still ends.
func TestRConWithoutOutput(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	var send = fixture.admin.GetSend()
	if send == nil {
		t.Fatal("a sender once connected")
		return
	}
	fixture.recorder.Clear()

	send.Rcon("companies")
	if !awaitTrue(t, func() bool { return len(fixture.recorder.GetRConEndEvents()) > 0 }, 10,
		"an RConEndEvent") {
		return
	}
	if fixture.recorder.GetRConEndEvents()[0].GetCommand() != "companies" {
		t.Errorf("command echo: got %s", fixture.recorder.GetRConEndEvents()[0].GetCommand())
	}
	if len(fixture.recorder.GetRConEvents()) != 0 {
		t.Errorf("no companies, so no output lines: %s", describeRCons(fixture.recorder.GetRConEvents()))
	}
}

// TestRConUnknownCommand is testRConUnknownCommand(): IConsoleCmdExec ends with
// "Command '{}' not found." in CC_ERROR.
func TestRConUnknownCommand(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	assertRCon(t, fixture, "not-a-command", "Command 'not-a-command' not found.", CC_ERROR)
}

// TestRConIsNotRelayedAsConsole is testRConIsNotRelayedAsConsole(): RCON_END must not fall
// through into a console event (see Game.UpdateAdmin).
func TestRConIsNotRelayedAsConsole(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	fixture.recorder.Clear()
	fixture.admin.GetSend().Rcon("getdate")

	if !awaitTrue(t, func() bool {
		return len(fixture.recorder.GetRConEndEvents()) > 0 && len(fixture.recorder.GetRConEvents()) > 0
	}, 10, "an RConEvent and its RConEndEvent") {
		return
	}

	// The redirected console goes as ADMIN_PACKET_SERVER_RCON. ADMIN_PACKET_SERVER_RCON_END
	// dispatches an RConEndEvent only; unrelated Debug(net) lines still arrive as console
	// events, so assert on the command's output, not on emptiness.
	var reply = fixture.recorder.GetRConEvents()[0].GetMessage()
	if consoleContains(fixture, reply) {
		t.Errorf("rcon output stays out of the console events: %s", reply)
	}
	for _, consoleEvent := range fixture.recorder.GetConsoleEvents() {
		if stringsContains(consoleEvent.GetMessage(), "Date: 19") {
			t.Errorf("no rcon output in console events: %s", consoleEvent.GetMessage())
			return
		}
	}
}

// TestConsoleRelayFromGamescript is testConsoleRelayFromGamescript():
// ADMIN_UPDATE_CONSOLE Automatic makes DebugReconsiderSendRemoteMessages queue Debug
// messages, and DebugSendRemoteMessages passes them to ADMIN_PACKET_SERVER_CONSOLE.
// AdminCmd's GSLog.Info is such a message.
func TestConsoleRelayFromGamescript(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	fixture.recorder.Clear()

	var request = gamescript.NewGSJsonRequest(constant.GS_COMMAND_COUNT_INDUSTRIE, []string{})
	fixture.admin.GetGSExecutor().Send(request)
	if !awaitTrue(t, request.HasResponse, 15, "the gamescript reply") {
		return
	}

	// main.nut: GSLog.Info("Command: " + id + " " + cmd + " " + args). Squirrel prints an
	// array as "(array : 0x...)", not "[]". debug.cpp passes the debug category as the
	// console origin, so the origin is "script".
	var commandLog = "Command: \\d+ " + strconv.Itoa(int(constant.GS_COMMAND_COUNT_INDUSTRIE)) + " \\(array"
	if !awaitTrue(t, func() bool {
		for _, consoleEvent := range fixture.recorder.GetConsoleEvents() {
			if consoleEvent.GetOrigin() == "script" && matchesPattern(commandLog, consoleEvent.GetMessage()) {
				return true
			}
		}
		return false
	}, 10, "the gamescript's console log of the command") {
		return
	}
}

// TestGameScriptCountIndustry is testGameScriptCountIndustry(): GSExecutor correlates the
// reply by request id, against the real gamescript.
func TestGameScriptCountIndustry(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	var request = gamescript.NewGSJsonRequest(constant.GS_COMMAND_COUNT_INDUSTRIE, []string{})
	fixture.admin.GetGSExecutor().Send(request)
	if !awaitTrue(t, request.HasResponse, 15, "the countIndustry reply") {
		return
	}
	var result, exception = request.GetResult()
	if exception != nil {
		t.Errorf("countIndustry failed: %s", exception.Error())
		return
	}
	if !matchesPattern("\\d+", result) {
		t.Errorf("industry count '%s'", result)
	}
}

// TestGameScriptNews is testGameScriptNews(): GSNewsPaper to all companies and to one
// company, which does not exist on a dedicated server.
func TestGameScriptNews(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	var executor = fixture.admin.GetGSExecutor()

	var broadcast = gamescript.NewGSNewsPaper(gamescript.NT_GENERAL, "Protocol broadcast")
	executor.Send(broadcast)
	if !awaitTrue(t, broadcast.HasResponse, 15, "the broadcast news reply") {
		return
	}
	var created, broadcastError = broadcast.GetResult()
	if broadcastError != nil {
		t.Errorf("broadcast news failed: %s", broadcastError.Error())
		return
	}
	if !created {
		t.Errorf("GSNewsPaper.New() returns true, got %v", created)
	}

	var forCompany = gamescript.NewGSNewsPaperForCompany(0, gamescript.NT_GENERAL, "Protocol company news")
	executor.Send(forCompany)
	if !awaitTrue(t, forCompany.HasResponse, 15, "the company news reply") {
		return
	}
	// No company 0 on a dedicated server, so ScriptNews refuses it; the reply still
	// arrives, which is what GSExecutor correlates.
	var refused, companyError = forCompany.GetResult()
	if companyError != nil {
		t.Errorf("company news failed: %s", companyError.Error())
		return
	}
	if refused {
		t.Errorf("company 0 does not exist, got %v", refused)
	}
	assertConnectionUsable(t, fixture)
}

// TestGameScriptGoals is testGameScriptGoals(): the goal helpers go through GSExecutor, so
// AdminCmd's reply is correlated. removeGoal used to pass the whole arguments array to
// GSGoal.Remove, which made the gamescript fail; the reply must be a removal, not an
// exception.
func TestGameScriptGoals(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	var send = fixture.admin.GetSend()
	if send == nil {
		t.Fatal("a sender once connected")
		return
	}
	fixture.recorder.Clear()

	send.AddGlobalGoal("protocol goal", constant.GOAL_NONE, 0)
	var added = awaitGameScriptData(t, fixture, 1)
	if !matchesPattern("\\d+", added) {
		t.Errorf("goal id '%s'", added)
		return
	}

	var goalId, ok = rule.ParseInt(added)
	if !ok {
		t.Errorf("goal id '%s' is not a number", added)
		return
	}
	send.RemoveGoal(goalId)
	var removed = awaitGameScriptData(t, fixture, 2)
	if removed != "true" {
		t.Errorf("removing a valid goal returns true, got '%s'", removed)
		return
	}

	send.RemoveAllGoal()
	// Nothing left to remove, so AdminCmd answers with false rather than an exception.
	awaitGameScriptData(t, fixture, 3)
	assertConnectionUsable(t, fixture)
}

// TestGameScriptTownCargoGoal is testGameScriptTownCargoGoal(): AdminCmd validates town and
// cargo effect itself, so a bad id is not an exception.
func TestGameScriptTownCargoGoal(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	fixture.recorder.Clear()
	fixture.admin.GetSend().SetTownCargoGoal(10000, constant.TOWN_EFFECT_MAIL, 100)
	if awaitGameScriptData(t, fixture, 1) != "false" {
		return
	}
	assertConnectionUsable(t, fixture)
}

// TestGameScriptUnknownCommand is testGameScriptUnknownCommand(): a command AdminCmd does
// not know is answered with {id, exception}; the raw JSON path is what robot rules use.
func TestGameScriptUnknownCommand(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	fixture.recorder.Clear()
	if !fixture.admin.GetSend().GameScript("{\"i\":\"4242\",\"c\":99,\"a\":[]}") {
		t.Error("raw json is sent")
		return
	}

	if !awaitTrue(t, func() bool {
		for _, gameScriptEvent := range fixture.recorder.GetGameScriptEvents() {
			var reply = gamescript.Parse(gameScriptEvent.GetMessage())
			if reply != nil && reply.HasException && reply.Exception == "Unknown command" {
				return true
			}
		}
		return false
	}, 15, "the 'Unknown command' reply") {
		return
	}
	assertConnectionUsable(t, fixture)
}

// TestGameScriptReplyWithoutRequest is testGameScriptReplyWithoutRequest(): a reply for a
// request that was never sent must not disturb the executor.
func TestGameScriptReplyWithoutRequest(t *testing.T) {
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if fixture == nil {
		return
	}
	fixture.recorder.Clear()
	if !fixture.admin.GetSend().GameScript("{\"i\":\"99999\",\"c\":0,\"a\":[]}") {
		t.Error("request sent")
		return
	}

	// GSExecutor has no request with id 99999; it logs and ignores the reply, so the
	// listener thread must survive and the connection stay open.
	if !awaitTrue(t, func() bool {
		for _, gameScriptEvent := range fixture.recorder.GetGameScriptEvents() {
			var reply = gamescript.Parse(gameScriptEvent.GetMessage())
			if reply != nil && reply.HasId && reply.Id == "99999" {
				return true
			}
		}
		return false
	}, 15, "the reply with id 99999") {
		return
	}
	assertConnectionUsable(t, fixture)
}

// TestLifecycleIs testLifecycle(): startup() shuts the previous client down first, and
// shutdown() works before startup().
func TestLifecycle(t *testing.T) {
	var fixture = integrationAdmin(t, secretsOrDefault(t), "openttd-admin lifecycle test")
	if fixture == nil {
		return
	}
	// The setUp fixture is the "still connected" admin Java asserts against at the end.
	var main = integrationFixture(t, secretsOrDefault(t), "openttd-admin protocol test")
	if main == nil {
		return
	}

	fixture.admin.Shutdown() // never started
	fixture.admin.Startup()
	fixture.admin.Startup() // the second startup() must not leave a dangling thread

	assertUsable(t, fixture.admin, fixture.recorder, 4242, 20)
	if main.admin.GetGame().GetGameInfo().GetServerName() == "" {
		t.Error("still connected")
	}

	fixture.admin.Shutdown() // Java's finally block
	if fixture.admin.IsConnected() {
		t.Error("shutdown() disconnects")
	}
}
