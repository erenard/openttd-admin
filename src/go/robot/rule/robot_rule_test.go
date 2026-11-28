// The specification for the robot rule layer: ExternalUsers, Administration,
// ServerAnnouncer, CompanyLifeCycle, CompanyPasswordRemover and ExposeGameInfo,
// driven through the same fake admin server and event recorder the admin layer
// tests use, so a rule is exercised the way the game director drives it: a packet
// arrives, the admin listener turns it into an event, the rule answers with rcon
// or chat packets.
//
// There are no Java tests for any class in com.openttd.robot, so unlike the
// earlier stages of this port nothing here is a port of an existing test: these
// cases are new coverage written against the Go sources as the specification,
// derived from the Java sources' behaviour. Where the port deviates from Java
// (a missing service is logged instead of raising a NullPointerException, a
// command with no argument sends its usage line instead of raising
// ArrayIndexOutOfBoundsException, a company with no owner prints an empty CEO
// column) the test pins the Go behaviour and says so.

package rule_test

import "strconv"
import "strings"
import "testing"
import "time"
import "com.openttd"
import "com.openttd/network/constant"
import "com.openttd/robot"
import netadmin "com.openttd/network/admin"
import rule "com.openttd/robot/rule"
import robotmodel "com.openttd/robot/model"
import "com.openttd/server"

// PASSWORD and ADMIN_NAME are AdminHarness' defaults, which the rules echo back
// as the caller's own name in the chat lines they send.
const PASSWORD = "test-password"
const ADMIN_NAME = "openttd-admin-test"

// CLIENT_ID and COMPANY_ID are the client and company the fixtures talk about.
const CLIENT_ID = 7
const COMPANY_ID = 6

// SETTLE_MILLIS is how long a test waits before concluding that a rule answered
// nothing at all.
const SETTLE_MILLIS = 200

type FakeUserService struct {
	identified []*robotmodel.ExternalUser
	logged     []string
}

func (s *FakeUserService) IdentifyUser(token string) *robotmodel.ExternalUser {
	s.logged = append(s.logged, token)
	if len(s.identified) == 0 {
		return nil
	}
	return s.identified[0]
}

func NewUserService(user *robotmodel.ExternalUser) *FakeUserService {
	if user == nil {
		return &FakeUserService{identified: []*robotmodel.ExternalUser{}, logged: []string{}}
	}
	return &FakeUserService{identified: []*robotmodel.ExternalUser{user}, logged: []string{}}
}

type FakeGameService struct {
	exposed        []netadmin.GameInfo
	saved          []robotmodel.GamePlayer
	scenarioId     int64
	endGameCalled  bool
	saveGameCalled bool
}

func (s *FakeGameService) GetGame(scenarioId int64) *netadmin.GameInfo {
	return nil
}

func (s *FakeGameService) ExposeGame(scenarioId int64, gameState *netadmin.GameInfo) {
	s.scenarioId = scenarioId
	s.exposed = append(s.exposed, *gameState)
}

func (s *FakeGameService) SaveGame(scenarioId int64, gamePlayers []robotmodel.GamePlayer) {
	s.saveGameCalled = true
	for _, player := range gamePlayers {
		s.saved = append(s.saved, player)
	}
}

func (s *FakeGameService) EndGame(scenarioId int64) {
	s.endGameCalled = true
}

// NewGameService injects a service for one scenario id, the way the embedding
// application does at startup.
func NewGameService(scenarioId int64) *FakeGameService {
	return &FakeGameService{exposed: []netadmin.GameInfo{}, saved: []robotmodel.GamePlayer{}, scenarioId: scenarioId,
		endGameCalled: false, saveGameCalled: false}
}

// NewAdminHarness starts the fake admin server, connects the admin and hands
// back the harness plus its recorder, so a test can both drive packets and read
// the events the rules react to.
func NewAdminHarness(t *testing.T) *server.AdminHarness {
	var harness, err = server.NewAdminHarness()
	if err != nil {
		t.Fatal(err.Error())
	}
	openttd.Await(func() bool { return harness.GetServer().GetConnectionCount() == 1 }, 10,
		"the admin to have connected")
	return harness
}

// ChatPacket is sendChat(NETWORK_ACTION_CHAT, DESTTYPE_BROADCAST, clientId,
// message): what the game director sends for a public chat line, and therefore
// what the rules see. The uint32 of a chat packet is the client the line comes
// from, which is what ChatEvent.getClientId() reports.
func ChatPacket(clientId int, message string) *server.PacketWriter {
	return server.New(constant.ADMIN_PACKET_SERVER_CHAT).
		WriteUint8(int(constant.NETWORK_ACTION_CHAT)).
		WriteUint8(int(constant.DESTTYPE_BROADCAST)).
		WriteUint32(uint32(clientId)).
		WriteString(message).
		WriteUint64(uint64(0))
}

// SendPacket writes one packet as the fake server. The admin reads it and
// dispatches the event on its own threads, so a test that expects an answer
// waits for that answer with AwaitReply, and a test that expects none waits a
// fixed SETTLE_MILLIS with Settle.
func SendPacket(harness *server.AdminHarness, packet *server.PacketWriter) {
	harness.GetServer().Send(packet)
}

// SendChat sends one chat line from CLIENT_ID.
func SendChat(harness *server.AdminHarness, message string) {
	SendPacket(harness, ChatPacket(CLIENT_ID, message))
}

// SendChatFrom sends one chat line from an arbitrary client, which the rules
// that police several players need.
func SendChatFrom(harness *server.AdminHarness, clientId int, message string) {
	SendPacket(harness, ChatPacket(clientId, message))
}

// SendDate sends a date packet for one day number.
func SendDate(harness *server.AdminHarness, day int) {
	SendPacket(harness, server.New(constant.ADMIN_PACKET_SERVER_DATE).
		WriteUint32(uint32(day)))
}

// SendCompanyEconomy reports one company's economy to the admin: the objectives
// read lastPerformance and previousPerformance, the other fields of the packet
// are filled with zeros because no rule of this port looks at them.
func SendCompanyEconomy(harness *server.AdminHarness, companyId int, lastPerformance int, previousPerformance int) {
	SendPacket(harness, server.New(constant.ADMIN_PACKET_SERVER_COMPANY_ECONOMY).
		WriteUint8(companyId).
		WriteInt64(0).WriteUint64(uint64(0)).WriteInt64(0).WriteUint16(0).
		WriteUint64(uint64(0)).WriteUint16(lastPerformance).WriteUint16(0).
		WriteUint64(uint64(0)).WriteUint16(previousPerformance).WriteUint16(0))
}

// DAY_1951 is the day number of 8 January 1951, the date the game clock shows
// when a test wants a realistic one: ISO week 1 of that year, the fourth trimester
// of it, and a game year later than the inauguration year SendCompanyInfo writes,
// which is what CompanyEconomyObjective's duration arithmetic needs.
const DAY_1951 = 712595

// DAY_APRIL is 23 April 1951, a week and a trimester after DAY_1951.
const DAY_APRIL = 712700

// SendClientJoin sends a client join packet for one client id.
func SendClientJoin(harness *server.AdminHarness, clientId int) {
	SendPacket(harness, server.New(constant.ADMIN_PACKET_SERVER_CLIENT_JOIN).
		WriteUint32(uint32(clientId)))
}

// SendClientInfo puts a client in the admin's model: ip, name, join date and
// company, the way the server reports a client.
func SendClientInfo(harness *server.AdminHarness, clientId int, ip string, name string, companyId int) {
	SendPacket(harness, server.New(constant.ADMIN_PACKET_SERVER_CLIENT_INFO).
		WriteUint32(uint32(clientId)).WriteString(ip).WriteString(name).
		WriteUint8(0).WriteUint32(uint32(1)).WriteUint8(companyId))
}

// SendCompanyNew sends a company new packet for one company id.
func SendCompanyNew(harness *server.AdminHarness, companyId int) {
	SendPacket(harness, server.New(constant.ADMIN_PACKET_SERVER_COMPANY_NEW).
		WriteUint8(companyId))
}

// SendCompanyRemove tells the admin a company was removed, with the reason the
// server gives (ADMIN_CRR_*).
func SendCompanyRemove(harness *server.AdminHarness, companyId int, reason constant.AdminCompanyRemoveReason) {
	SendPacket(harness, server.New(constant.ADMIN_PACKET_SERVER_COMPANY_REMOVE).
		WriteUint8(companyId).WriteUint8(int(reason)))
}

// SendCompanyInfo puts a company in the admin's model, with or without a
// password set, which is what CompanyPasswordRemainder looks at.
func SendCompanyInfo(harness *server.AdminHarness, companyId int, name string, usePassword bool) {
	SendPacket(harness, server.New(constant.ADMIN_PACKET_SERVER_COMPANY_INFO).
		WriteUint8(companyId).WriteString(name).WriteString("Manager").
		WriteUint8(1).WriteBool8(usePassword).WriteUint32(uint32(1950)).
		WriteBool8(false).WriteUint8(0).WriteUint8(255).WriteUint8(255).
		WriteUint8(255).WriteUint8(255))
}

// AwaitChat waits until the admin has sent count chat packets. Chat and rcon
// are counted separately on purpose: a total would let one rule's answer satisfy
// the wait for another's.
func AwaitChat(harness *server.AdminHarness, count int) {
	openttd.AwaitSize(func() int { return len(ChatMessages(harness)) }, count, 10,
		strconv.Itoa(count)+" chat replies")
}

// AwaitRcon waits until the admin has sent count rcon packets.
func AwaitRcon(harness *server.AdminHarness, count int) {
	openttd.AwaitSize(func() int { return len(RconCommands(harness)) }, count, 10,
		strconv.Itoa(count)+" rcon replies")
}

// Settle gives the event dispatcher time to run a rule that is expected to
// answer nothing, the way the mapping test's "none" rows do.
func Settle() {
	time.Sleep(SETTLE_MILLIS * time.Millisecond)
}

// RconCommands is the rcon commands the admin has sent, oldest first, decoded
// from the wire.
func RconCommands(harness *server.AdminHarness) []string {
	var commands = []string{}
	for _, p := range harness.GetServer().Received() {
		if p.GetWireTypeId() == int(constant.ADMIN_PACKET_ADMIN_RCON) {
			commands = append(commands, p.GetPacket().ReadString())
		}
	}
	return commands
}

// ChatMessages is the chat lines the admin has sent, oldest first. The admin
// decoder skips the action and destination type, so the reader has to skip them
// too before it gets to the message.
func ChatMessages(harness *server.AdminHarness) []string {
	var messages = []string{}
	for _, p := range harness.GetServer().Received() {
		if p.GetWireTypeId() == int(constant.ADMIN_PACKET_ADMIN_CHAT) {
			var packet = p.GetPacket()
			packet.ReadUint8()  // action
			packet.ReadUint8()  // destination type
			packet.ReadUint32() // destination id
			messages = append(messages, packet.ReadString())
		}
	}
	return messages
}

// HasRcon reports whether command is one of the rcon commands sent.
func HasRcon(harness *server.AdminHarness, command string) bool {
	return strings.Contains(strings.Join(RconCommands(harness), "\n"), command)
}

// HasChat reports whether message is one of the chat lines sent.
func HasChat(harness *server.AdminHarness, message string) bool {
	return strings.Contains(strings.Join(ChatMessages(harness), "\n"), message)
}

// CountRcon is how many rcon commands were sent.
func CountRcon(harness *server.AdminHarness) int {
	return len(harness.GetServer().Packets(constant.ADMIN_PACKET_ADMIN_RCON))
}

// CountChat is how many chat lines were sent.
func CountChat(harness *server.AdminHarness) int {
	return len(harness.GetServer().Packets(constant.ADMIN_PACKET_ADMIN_CHAT))
}

// NewUser builds the record the application's ExternalUserService hands back.
func NewUser(id int64, name string, isAdmin bool) *robotmodel.ExternalUser {
	var user = robotmodel.NewExternalUser()
	user.SetId(id)
	user.SetName(name)
	user.SetAdmin(isAdmin)
	return user
}

// InjectUserService puts a service that answers every token with user into the
// process-wide ExternalServices, the way the embedding application does at
// startup. ExternalUsers reads the singleton when it is constructed, so this
// has to happen first.
func InjectUserService(user *robotmodel.ExternalUser) {
	robot.GetInstance().SetExternalUserService(NewUserService(user))
}

// TestLoginRegistersTheClient is the "!login token" rule: the application says
// who the player is, so the rule remembers it, greets them, and renames the
// client to the name the application gave them.
func TestLoginRegistersTheClient(t *testing.T) {
	var harness = NewAdminHarness(t)
	defer harness.Close()
	InjectUserService(NewUser(1, "Alice", false))
	var users = rule.NewExternalUsers(harness.GetAdmin())
	if users == nil {
		t.Errorf("ExternalUsers: nil")
		return
	}
	SendChat(harness, "!login 0f8a3c")
	AwaitChat(harness, 2)
	AwaitRcon(harness, 1)
	if !HasChat(harness, "Login succeed.") {
		t.Errorf("login: %s", strings.Join(ChatMessages(harness), " | "))
	}
	if !HasRcon(harness, "client_name 7 Alice") {
		t.Errorf("rename: %s", strings.Join(RconCommands(harness), " | "))
	}
	var user = users.GetExternalUser(CLIENT_ID)
	if user == nil {
		t.Errorf("logged user: none")
	} else if user.GetName() != "Alice" {
		t.Errorf("logged user: %s", user.ToString())
	}
}

// TestLoginWithoutServiceReportsFailure is the port's deviation from Java: a
// missing ExternalUserService is logged and reported as a failed login, where
// Java raises a NullPointerException on the dispatcher thread.
func TestLoginWithoutServiceReportsFailure(t *testing.T) {
	var harness = NewAdminHarness(t)
	defer harness.Close()
	robot.GetInstance().SetExternalUserService(nil)
	var users = rule.NewExternalUsers(harness.GetAdmin())
	SendChat(harness, "!login 0f8a3c")
	AwaitChat(harness, 2)
	if !HasChat(harness, "Login failed.") {
		t.Errorf("failed login: %s", strings.Join(ChatMessages(harness), " | "))
	}
	if !HasChat(harness, "How to login ***") {
		t.Errorf("howto after a failed login: %s", strings.Join(ChatMessages(harness), " | "))
	}
	Settle()
	if users.GetExternalUser(CLIENT_ID) != nil {
		t.Errorf("a failed login must not register a user")
	}
	if len(RconCommands(harness)) != 0 {
		t.Errorf("a failed login renamed the client: %s", strings.Join(RconCommands(harness), " | "))
	}
}

// TestLoginTwiceReportsAlreadyLogged is the second half of the login rule: the
// same token again is refused, and the client is not renamed a second time.
func TestLoginTwiceReportsAlreadyLogged(t *testing.T) {
	var harness = NewAdminHarness(t)
	defer harness.Close()
	InjectUserService(NewUser(1, "Alice", false))
	var users = rule.NewExternalUsers(harness.GetAdmin())
	SendChat(harness, "!login 0f8a3c")
	AwaitChat(harness, 2)
	SendChat(harness, "!login 0f8a3c")
	AwaitChat(harness, 3)
	if !HasChat(harness, "You are already logged.") {
		t.Errorf("second login: %s", strings.Join(ChatMessages(harness), " | "))
	}
	if len(RconCommands(harness)) != 1 {
		t.Errorf("second login renamed again: %s", strings.Join(RconCommands(harness), " | "))
	}
	if users.GetExternalUser(CLIENT_ID) == nil {
		t.Errorf("second login lost the registration")
	}
}

// TestRenameRejectsPlayerNames is the "!rename" rule: a name the application
// would confuse with the server's own placeholder players is refused, spaces
// become underscores, and an empty name gets the usage line.
func TestRenameRejectsPlayerNames(t *testing.T) {
	var harness = NewAdminHarness(t)
	defer harness.Close()
	robot.GetInstance().SetExternalUserService(nil)
	rule.NewExternalUsers(harness.GetAdmin())
	SendChat(harness, "!rename Player 5")
	AwaitChat(harness, 1)
	if !HasChat(harness, "Player... names are not allowed, try again...") {
		t.Errorf("Player name: %s", strings.Join(ChatMessages(harness), " | "))
	}
	SendChat(harness, "!rename New Name")
	AwaitRcon(harness, 1)
	if !HasRcon(harness, "client_name 7 New_Name") {
		t.Errorf("spaced name: %s", strings.Join(RconCommands(harness), " | "))
	}
	SendChat(harness, "!rename x")
	AwaitChat(harness, 2)
	if !HasChat(harness, "Try : !rename NewName") {
		t.Errorf("short name: %s", strings.Join(ChatMessages(harness), " | "))
	}
}

// TestHowtoSendsTheFourLines is the "!howto" rule: the four lines that tell a
// player where to get a login token.
func TestHowtoSendsTheFourLines(t *testing.T) {
	var harness = NewAdminHarness(t)
	defer harness.Close()
	robot.GetInstance().SetExternalUserService(nil)
	rule.NewExternalUsers(harness.GetAdmin())
	SendChat(harness, "!howto")
	AwaitChat(harness, 4)
	if !HasChat(harness, "3. Copy and paste the whole !login command line into the chat window of the game.") {
		t.Errorf("howto: %s", strings.Join(ChatMessages(harness), " | "))
	}
	if CountRcon(harness) != 0 {
		t.Errorf("howto sent rcon: %s", strings.Join(RconCommands(harness), " | "))
	}
}

// ---------------------------------------------------------------------------
// Administration
// ---------------------------------------------------------------------------

// ExternalUsersFixture is the ExternalUsers rule plus the service that answers
// its login tokens, kept together because the Administration tests need both.
type ExternalUsersFixture struct {
	users   *rule.ExternalUsers
	service *FakeUserService
}

// NewExternalUsersFixture injects a service that answers every token with user,
// then builds the rule, in that order: ExternalUsers reads the ExternalServices
// singleton when it is constructed, so injecting afterwards would leave it with
// no service at all.
func NewExternalUsersFixture(harness *server.AdminHarness, user *robotmodel.ExternalUser) *ExternalUsersFixture {
	var service = NewUserService(user)
	robot.GetInstance().SetExternalUserService(service)
	var users = rule.NewExternalUsers(harness.GetAdmin())
	if users == nil {
		return nil
	}
	return &ExternalUsersFixture{users: users, service: service}
}

// LogInAdmin logs one player in as an administrator, the only status the
// Administration rule accepts, and hands back the fixture its rules are built
// on.
func LogInAdmin(harness *server.AdminHarness) *ExternalUsersFixture {
	var fixture = NewExternalUsersFixture(harness, NewUser(1, "Alice", true))
	if fixture == nil {
		return nil
	}
	SendChat(harness, "!login 0f8a3c")
	AwaitChat(harness, 2)
	return fixture
}

// TestNonAdminSendsNothing is the Administration rule's gate: a player who is
// not an administrator gets no answer at all, not even for $clients.
func TestNonAdminSendsNothing(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = NewExternalUsersFixture(harness, NewUser(1, "Alice", false))
	if fixture == nil {
		return
	}
	rule.NewAdministration(harness.GetAdmin(), fixture.users)
	SendChat(harness, "$clients")
	Settle()
	if len(ChatMessages(harness)) != 0 || len(RconCommands(harness)) != 0 {
		t.Errorf("a non admin answered: %s / %s", strings.Join(ChatMessages(harness), " | "),
			strings.Join(RconCommands(harness), " | "))
	}
}

// TestClientsListsTheModel is "$clients": the header line, then one row per
// client in the admin's model, filled by CLIENT_INFO packets from the fake
// server.
func TestClientsListsTheModel(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = LogInAdmin(harness)
	if fixture == nil {
		return
	}
	rule.NewAdministration(harness.GetAdmin(), fixture.users)
	SendClientInfo(harness, CLIENT_ID, "1.2.3.4.5.6.7.8", "Alice", COMPANY_ID)
	SendClientInfo(harness, 9, "9.8.7.6.5.4.3.2", "Bob", 7)
	var before = len(ChatMessages(harness))
	SendChat(harness, "$clients")
	AwaitChat(harness, before+3)
	var messages = ChatMessages(harness)
	if messages[before] != "#ClientId, Name, IpAddress, CompanyId" {
		t.Errorf("clients header = %q", messages[before])
	}
	if messages[before+1] != "#7, Alice, 1.2.3.4.5.6.7.8, 6" {
		t.Errorf("clients row 1 = %q", messages[before+1])
	}
	if messages[before+2] != "#9, Bob, 9.8.7.6.5.4.3.2, 7" {
		t.Errorf("clients row 2 = %q", messages[before+2])
	}
}

// TestKickAndBanUseRcon checks the moderation commands are turned into rcon
// commands, never into chat messages.
func TestKickAndBanUseRcon(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = LogInAdmin(harness)
	if fixture == nil {
		return
	}
	rule.NewAdministration(harness.GetAdmin(), fixture.users)
	var chat = len(ChatMessages(harness))
	var rcon = len(RconCommands(harness))
	SendChat(harness, "$kick 3")
	SendChat(harness, "$b 4")
	SendChat(harness, "$p")
	SendChat(harness, "$u")
	AwaitRcon(harness, rcon+4)
	var commands = RconCommands(harness)
	if commands[rcon] != "kick 3" || commands[rcon+1] != "ban 4" ||
		commands[rcon+2] != "pause" || commands[rcon+3] != "unpause" {
		t.Errorf("rcon commands = %s", strings.Join(commands[rcon:], " | "))
	}
	if len(ChatMessages(harness)) != chat {
		t.Errorf("moderation answered in chat: %s",
			strings.Join(ChatMessages(harness)[chat:], " | "))
	}
}

// TestKickWithoutArgumentSendsUsage is Java's array bounds check: without a
// clientId the rule sends its usage line instead of throwing.
func TestKickWithoutArgumentSendsUsage(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = LogInAdmin(harness)
	if fixture == nil {
		return
	}
	rule.NewAdministration(harness.GetAdmin(), fixture.users)
	var chat = len(ChatMessages(harness))
	var rcon = len(RconCommands(harness))
	SendChat(harness, "$kick")
	AwaitChat(harness, chat+1)
	var messages = ChatMessages(harness)
	if messages[chat] != "Usage: $kick clientId (try $clients to find clientIds)" {
		t.Errorf("kick usage = %q", messages[chat])
	}
	if len(RconCommands(harness)) != rcon {
		t.Errorf("kick without argument sent rcon: %s",
			strings.Join(RconCommands(harness)[rcon:], " | "))
	}
}

// TestWarnEscalatesToKickThenBan walks the warning ladder: the count read
// decides what this warning does, then is incremented, so the third warning
// kicks and the fourth bans.
func TestWarnEscalatesToKickThenBan(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = LogInAdmin(harness)
	if fixture == nil {
		return
	}
	rule.NewAdministration(harness.GetAdmin(), fixture.users)
	var chat = len(ChatMessages(harness))
	var rcon = len(RconCommands(harness))
	SendChat(harness, "$warn 3")
	SendChat(harness, "$warn 3 slow down")
	SendChat(harness, "$warn 3")
	SendChat(harness, "$warn 3")
	AwaitChat(harness, chat+2)
	AwaitRcon(harness, rcon+2)
	var messages = ChatMessages(harness)
	if messages[chat] != "Be warned! No rule breaking!" {
		t.Errorf("first warning = %q", messages[chat])
	}
	if messages[chat+1] != "Last warning! No rule breaking! (slow down)" {
		t.Errorf("second warning = %q", messages[chat+1])
	}
	var commands = RconCommands(harness)
	if commands[rcon] != "kick 3" || commands[rcon+1] != "ban 3" {
		t.Errorf("escalation = %s", strings.Join(commands[rcon:], " | "))
	}
}

// TestHelpListsTheCommands is "$help": the long and the short command list.
func TestHelpListsTheCommands(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = LogInAdmin(harness)
	if fixture == nil {
		return
	}
	rule.NewAdministration(harness.GetAdmin(), fixture.users)
	var chat = len(ChatMessages(harness))
	SendChat(harness, "$help")
	AwaitChat(harness, chat+2)
	var messages = ChatMessages(harness)
	if messages[chat] != "Valid commands are $clients, $companies, $warn, $kick, $ban, $reset, $pause, $unpause" {
		t.Errorf("help line 1 = %q", messages[chat])
	}
	if messages[chat+1] != "Short commands are $cls, $cps, $w, $k, $b, $r, $p, $u" {
		t.Errorf("help line 2 = %q", messages[chat+1])
	}
}

// ---------------------------------------------------------------------------
// ServerAnnouncer
// ---------------------------------------------------------------------------

// TestWelcomeGreetsNewClients is the ClientEvent CREATE case: the two welcome
// lines are sent to the client that just joined.
func TestWelcomeGreetsNewClients(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	rule.NewServerAnnouncer(harness.GetAdmin())
	var before = CountChat(harness)
	SendClientJoin(harness, 9)
	AwaitChat(harness, before+2)
	var messages = ChatMessages(harness)
	if messages[before] != "Welcome to www.strategyboard.net goal server." {
		t.Errorf("welcome line 1 = %q", messages[before])
	}
	if messages[before+1] != "Type: !howto or !help." {
		t.Errorf("welcome line 2 = %q", messages[before+1])
	}
}

// TestRulesListsTheElevenRules is "!rules": the preamble, the eleven numbered
// rules and the link, thirteen lines in all.
func TestRulesListsTheElevenRules(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	rule.NewServerAnnouncer(harness.GetAdmin())
	var before = CountChat(harness)
	SendChat(harness, "!rules")
	AwaitChat(harness, before+13)
	var messages = ChatMessages(harness)
	if messages[before] != "Any violation of these rules will make you banned." {
		t.Errorf("rules preamble = %q", messages[before])
	}
	if messages[before+12] != "http://www.strategyboard.net/ for more info" {
		t.Errorf("rules footer = %q", messages[before+12])
	}
	if messages[before+7] != "# 6: No one way trains!" {
		t.Errorf("rules line 7 = %q", messages[before+7])
	}
	if CountChat(harness) != before+13 {
		t.Errorf("rules sent %d lines, expected 13", CountChat(harness)-before)
	}
}

// TestHelpListsThePlayerCommands is "!help": the two player command lists.
func TestHelpListsThePlayerCommands(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	rule.NewServerAnnouncer(harness.GetAdmin())
	var before = CountChat(harness)
	SendChat(harness, "!help")
	AwaitChat(harness, before+2)
	var messages = ChatMessages(harness)
	if messages[before] != "Valid commands are !goal, !rules, !login, !howto, !score, !rename and !resetme." {
		t.Errorf("help line 1 = %q", messages[before])
	}
	if messages[before+1] != "Short commands are !g, !r, !cv or !cp (equ. to !score)" {
		t.Errorf("help line 2 = %q", messages[before+1])
	}
}

// InjectGameService puts a recording service for one scenario into the
// process-wide ExternalServices. ExposeGameInfo reads the singleton when a
// date arrives, so the order does not matter for it, but the objectives capture
// it at construction.
func InjectGameService(scenarioId int64) *FakeGameService {
	var service = NewGameService(scenarioId)
	robot.GetInstance().SetExternalGameService(service)
	return service
}

// ---------------------------------------------------------------------------
// CompanyPasswordRemainder
// ---------------------------------------------------------------------------

// TestWarnsCompaniesWithoutPassword is the DateEvent rule: once the month key
// moves forward, every company that has a password set is left alone and every
// company without one is warned.
func TestWarnsCompaniesWithoutPassword(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	rule.NewCompanyPasswordRemainder(harness.GetAdmin())
	SendCompanyInfo(harness, COMPANY_ID, "NoPass", false)
	SendCompanyInfo(harness, 3, "HasPass", true)
	var before = CountChat(harness)
	// Day 31 is the second month of the game clock: the month key starts at 0,
	// so the first date packet that moves it forward is the one that warns.
	SendDate(harness, 31)
	AwaitChat(harness, before+1)
	if !HasChat(harness, "Warning: your company has NO PASSWORD.") {
		t.Errorf("no warning for the passwordless company: %s",
			strings.Join(ChatMessages(harness)[before:], " | "))
	}
	if CountChat(harness) != before+1 {
		t.Errorf("the company with a password was warned too: %s",
			strings.Join(ChatMessages(harness)[before:], " | "))
	}
	// The same date again: the month key did not move, so nothing is sent.
	SendDate(harness, 31)
	Settle()
	if CountChat(harness) != before+1 {
		t.Errorf("the same month warned again: %s",
			strings.Join(ChatMessages(harness)[before:], " | "))
	}
	// A month later warns once more.
	SendDate(harness, 70)
	AwaitChat(harness, before+2)
}

// ---------------------------------------------------------------------------
// ExposeGameInfo
// ---------------------------------------------------------------------------

// TestExposeGameInfoPublishesOncePerWeek is the DateEvent rule that pushes the
// game's info to the application: once per week, never twice in the same one.
func TestExposeGameInfoPublishesOncePerWeek(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var service = InjectGameService(42)
	rule.NewExposeGameInfo(harness.GetAdmin(), 42)
	SendDate(harness, 0)
	openttd.Await(func() bool { return len(service.exposed) == 1 }, 10,
		"the game info to be exposed")
	if service.scenarioId != 42 {
		t.Errorf("exposed for scenario %d", service.scenarioId)
	}
	// The next day is still the same ISO week.
	SendDate(harness, 1)
	Settle()
	if len(service.exposed) != 1 {
		t.Errorf("exposed %d times inside one week", len(service.exposed))
	}
	// The following week exposes it again.
	SendDate(harness, 8)
	openttd.Await(func() bool { return len(service.exposed) == 2 }, 10,
		"the game info to be exposed in the next week")
}

// ---------------------------------------------------------------------------
// CompanyLifeCycle
// ---------------------------------------------------------------------------

// TestCompanyCreationNeedsALoggedPlayer is Rule #1: a player the application
// does not know may not start a company, so the rule explains, tells them how
// to log in, and hands the company back to the game with the rcon pair.
func TestCompanyCreationNeedsALoggedPlayer(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	robot.GetInstance().SetExternalUserService(nil)
	rule.NewCompanyLifeCycle(harness.GetAdmin(), rule.NewExternalUsers(harness.GetAdmin()))
	SendClientInfo(harness, CLIENT_ID, "1.2.3.4.5.6.7.8", "Alice", COMPANY_ID)
	var chat = CountChat(harness)
	var rcon = CountRcon(harness)
	SendCompanyNew(harness, COMPANY_ID)
	AwaitChat(harness, chat+5)
	var messages = ChatMessages(harness)
	if messages[chat] != "Company 6 creation rejected: Please loggin before start a new company." {
		t.Errorf("rejection = %q", messages[chat])
	}
	if messages[chat+1] != "How to login ***" {
		t.Errorf("howto after a rejected company: %q", messages[chat+1])
	}
	AwaitRcon(harness, rcon+2)
	var commands = RconCommands(harness)
	if commands[rcon] != "move 7 255" || commands[rcon+1] != "reset_company 7" {
		t.Errorf("delete rcon = %s", strings.Join(commands[rcon:], " | "))
	}
}

// TestCompanyCreationRegistersTheOwner is the accepted half of the CREATE case:
// a logged player keeps the company, and the rule records them as its owner.
func TestCompanyCreationRegistersTheOwner(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = NewExternalUsersFixture(harness, NewUser(2, "Carol", false))
	if fixture == nil {
		return
	}
	var life = rule.NewCompanyLifeCycle(harness.GetAdmin(), fixture.users)
	if life == nil {
		return
	}
	SendChat(harness, "!login 7ba9c2")
	AwaitChat(harness, 2)
	SendClientInfo(harness, CLIENT_ID, "1.2.3.4.5.6.7.8", "Alice", COMPANY_ID)
	var chat = CountChat(harness)
	var rcon = CountRcon(harness)
	SendCompanyNew(harness, COMPANY_ID)
	Settle()
	if fixture.users.GetOwnerOf(COMPANY_ID) == nil {
		t.Errorf("the creator was not recorded as the owner of company 6")
	}
	if CountChat(harness) != chat || CountRcon(harness) != rcon {
		t.Errorf("an accepted company was answered: %s / %s",
			strings.Join(ChatMessages(harness)[chat:], " | "),
			strings.Join(RconCommands(harness)[rcon:], " | "))
	}
}

// TestSecondCompanyBySameIpIsRejected is Rule #2: one player may not own two
// companies, so the second one is deleted again with the rcon pair.
func TestSecondCompanyBySameIpIsRejected(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = NewExternalUsersFixture(harness, NewUser(3, "Carol", false))
	if fixture == nil {
		return
	}
	rule.NewCompanyLifeCycle(harness.GetAdmin(), fixture.users)
	// The first company: client 7, logged in, owns company 6.
	SendClientInfo(harness, CLIENT_ID, "1.2.3.4.5.6.7.8", "Alice", COMPANY_ID)
	SendCompanyInfo(harness, COMPANY_ID, "First", true)
	SendChat(harness, "!login 7ba9c2")
	AwaitChat(harness, 2)
	SendCompanyNew(harness, COMPANY_ID)
	Settle()
	// The second one comes from another client on the same ip address.
	SendClientInfo(harness, 9, "1.2.3.4.5.6.7.8", "Bob", 8)
	SendChatFrom(harness, 9, "!login 8db1d3")
	AwaitChat(harness, 4)
	var chat = CountChat(harness)
	var rcon = CountRcon(harness)
	SendCompanyNew(harness, 8)
	AwaitChat(harness, chat+1)
	var messages = ChatMessages(harness)
	if messages[chat] != "Company '' creation rejected: You (1.2.3.4.5.6.7.8) already own 'First'." {
		t.Errorf("ip rejection = %q", messages[chat])
	}
	AwaitRcon(harness, rcon+2)
	var commands = RconCommands(harness)
	if commands[rcon] != "move 9 255" || commands[rcon+1] != "reset_company 9" {
		t.Errorf("delete rcon = %s", strings.Join(commands[rcon:], " | "))
	}
}

// TestResetmeEmptiesThePlayersCompany is Rule #3: "!resetme" hands the player's
// own company back to the game, which is the same rcon pair the rejection uses.
func TestResetmeEmptiesThePlayersCompany(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	rule.NewCompanyLifeCycle(harness.GetAdmin(), rule.NewExternalUsers(harness.GetAdmin()))
	SendClientInfo(harness, CLIENT_ID, "1.2.3.4.5.6.7.8", "Alice", COMPANY_ID)
	var rcon = CountRcon(harness)
	SendChat(harness, "!resetme")
	AwaitRcon(harness, rcon+2)
	var commands = RconCommands(harness)
	if commands[rcon] != "move 7 255" || commands[rcon+1] != "reset_company 7" {
		t.Errorf("resetme rcon = %s", strings.Join(commands[rcon:], " | "))
	}
	if CountChat(harness) != 0 {
		t.Errorf("resetme answered in chat: %s", strings.Join(ChatMessages(harness), " | "))
	}
}

// TestCompaniesListsTheOwners is Rule #4: the admin's "$companies" dump, one
// padded line per company, headed by the column list.
func TestCompaniesListsTheOwners(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = NewExternalUsersFixture(harness, NewUser(4, "Dave", true))
	if fixture == nil {
		return
	}
	if rule.NewCompanyLifeCycle(harness.GetAdmin(), fixture.users) == nil {
		return
	}
	SendClientInfo(harness, CLIENT_ID, "1.2.3.4.5.6.7.8", "Alice", COMPANY_ID)
	SendCompanyInfo(harness, COMPANY_ID, "First", true)
	SendChat(harness, "!login 9c4e11")
	AwaitChat(harness, 2)
	SendCompanyNew(harness, COMPANY_ID)
	Settle()
	var chat = CountChat(harness)
	SendChat(harness, "$companies")
	AwaitChat(harness, chat+2)
	var messages = ChatMessages(harness)
	if messages[chat] != "#Id, Name, CEO, Inauguration." {
		t.Errorf("companies header = %q", messages[chat])
	}
	// Java's line is a "#" StringBuilder plus a Formatter, so it starts with
	// two hashes, and every column is padded to its width.
	var line = messages[chat+1]
	if !strings.HasPrefix(line, "##   6, First") {
		t.Errorf("companies row = %q", line[0:20])
	}
	if !strings.Contains(line, "Dave") || !strings.HasSuffix(line, "1950") {
		t.Errorf("companies row misses the CEO or the year: %q", line[255:])
	}
}

// TestResetDeletesTheCompanyIs Rule #6: an admin names a company and the rule
// hands it back to the game.
func TestResetDeletesTheCompany(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = NewExternalUsersFixture(harness, NewUser(5, "Dave", true))
	if fixture == nil {
		return
	}
	if rule.NewCompanyLifeCycle(harness.GetAdmin(), fixture.users) == nil {
		return
	}
	SendClientInfo(harness, CLIENT_ID, "1.2.3.4.5.6.7.8", "Alice", COMPANY_ID)
	SendChat(harness, "!login 2f6b8d")
	AwaitChat(harness, 2)
	var rcon = CountRcon(harness)
	SendChat(harness, "$reset 6")
	AwaitRcon(harness, rcon+2)
	var commands = RconCommands(harness)
	if commands[rcon] != "move 7 255" || commands[rcon+1] != "reset_company 7" {
		t.Errorf("reset rcon = %s", strings.Join(commands[rcon:], " | "))
	}
}

// TestResetWithoutArgumentSendsUsage pins the deviation: Java raises
// ArrayIndexOutOfBoundsException on a command with no argument, the port sends
// the usage line the NumberFormatException branch would have sent.
func TestResetWithoutArgumentSendsUsage(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = NewExternalUsersFixture(harness, NewUser(6, "Dave", true))
	if fixture == nil {
		return
	}
	if rule.NewCompanyLifeCycle(harness.GetAdmin(), fixture.users) == nil {
		return
	}
	SendChat(harness, "!login 4a1c90")
	AwaitChat(harness, 2)
	var chat = CountChat(harness)
	SendChat(harness, "$reset")
	AwaitChat(harness, chat+1)
	if ChatMessages(harness)[chat] != "Usage: $reset companyId (try $companies to find companyIds)" {
		t.Errorf("usage = %q", ChatMessages(harness)[chat])
	}
}

// TestInfoDescribesThePlayer is Rule #5: "!info" tells the player who they are
// and who owns the company they play as.
func TestInfoDescribesThePlayer(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = NewExternalUsersFixture(harness, NewUser(7, "Dave", true))
	if fixture == nil {
		return
	}
	if rule.NewCompanyLifeCycle(harness.GetAdmin(), fixture.users) == nil {
		return
	}
	SendClientInfo(harness, CLIENT_ID, "1.2.3.4.5.6.7.8", "Alice", COMPANY_ID)
	SendCompanyInfo(harness, COMPANY_ID, "First", true)
	SendChat(harness, "!login 3e7d22")
	AwaitChat(harness, 2)
	SendCompanyNew(harness, COMPANY_ID)
	Settle()
	var chat = CountChat(harness)
	SendChat(harness, "!info")
	AwaitChat(harness, chat+2)
	var messages = ChatMessages(harness)
	if messages[chat] != "User: Dave, play as (Alice)." {
		t.Errorf("info line = %q", messages[chat])
	}
	if messages[chat+1] != "Company: First, owned by Dave." {
		t.Errorf("info company = %q", messages[chat+1])
	}
}

// TestDeleteClearsTheOwnerAndTheIp is the DELETE case: the owner record and the
// ip address entry both go away, so the same player may start a company again.
func TestDeleteClearsTheOwnerAndTheIp(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = NewExternalUsersFixture(harness, NewUser(8, "Carol", false))
	if fixture == nil {
		return
	}
	if rule.NewCompanyLifeCycle(harness.GetAdmin(), fixture.users) == nil {
		return
	}
	SendClientInfo(harness, CLIENT_ID, "1.2.3.4.5.6.7.8", "Alice", COMPANY_ID)
	SendChat(harness, "!login 1b5f77")
	AwaitChat(harness, 2)
	SendCompanyNew(harness, COMPANY_ID)
	Settle()
	if fixture.users.GetOwnerOf(COMPANY_ID) == nil {
		t.Fatalf("the creator was never recorded as the owner")
	}
	// The company disappears from the game.
	SendCompanyRemove(harness, COMPANY_ID, constant.ADMIN_CRR_BANKRUPT)
	Settle()
	if fixture.users.GetOwnerOf(COMPANY_ID) != nil {
		t.Errorf("the owner of a deleted company is still recorded")
	}
	// The ip address is free again, so a new company from the same player is
	// accepted: no rejection, no rcon pair.
	SendClientInfo(harness, 9, "1.2.3.4.5.6.7.8", "Bob", 8)
	SendChatFrom(harness, 9, "!login 6c2a44")
	AwaitChat(harness, 4)
	var chat = CountChat(harness)
	var rcon = CountRcon(harness)
	SendCompanyNew(harness, 8)
	Settle()
	if CountChat(harness) != chat || CountRcon(harness) != rcon {
		t.Errorf("a company on a freed ip was rejected: %s / %s",
			strings.Join(ChatMessages(harness)[chat:], " | "),
			strings.Join(RconCommands(harness)[rcon:], " | "))
	}
}
