// The specification for the demo package: CLIUtil's command line, TestUtil's
// fake services, the two rules that greet the players on Mondays, the console
// that logs one line per event, and AdminPing's single ping.
//
// There is no test/java/com/openttd/demo directory: the Java tree never tested
// these samples, so every test below is new coverage, written against the Go
// sources as the specification. Where a Java sample cannot be driven at all -
// its main() sleeps for minutes, or connects to a server that is not there -
// the tests cover the pieces around it and leave Main alone.

package demo_test

import "os"
import "strings"
import "strconv"
import "time"
import "testing"
import "com.openttd"
import "com.openttd/demo"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/server"
import "com.openttd/util"
import "com.openttd/admin/event"
import "com.openttd/robot"

// PASSWORD is the admin password the samples put into the Configuration.
const PASSWORD = "lepasswd"

// MONDAY_DAY and SATURDAY_DAY are the first week of the game's own date
// counter: day 0 is a Saturday, so day 2 is the first Monday. util_test pins
// both, the demos only use them.
const MONDAY_DAY = 2
const SATURDAY_DAY = 0

// SETTLE_MILLIS is how long to wait before concluding a rule sent nothing, as
// the rule tests do: the sender is a thread, so "nothing yet" needs a moment
// before it can be told apart from "nothing at all".
const SETTLE_MILLIS = 200

// LogCapture points the log shim at a file, so a test can read what a demo
// logged - the port's stand-in for the System.out and System.err the Java
// samples write to. Same idiom as the launcher test's capture.
type LogCapture struct {
	path              string
	file              *os.File
	previous          *os.File
	previousThreshold util.Level
}

func NewLogCapture() *LogCapture {
	var file, fileError = os.CreateTemp("", "openttd-admin-demo-log")
	if fileError != nil {
		panic("cannot open the demo log: " + fileError.Error())
	}
	var capture = &LogCapture{
		path:              file.Name(),
		file:              file,
		previous:          util.GetOutput(),
		previousThreshold: util.GetThreshold()}
	util.SetOutput(file)
	return capture
}

// Close puts the log shim back the way it was and drops the file.
func (c *LogCapture) Close() {
	c.file.Close()
	util.SetOutput(c.previous)
	util.SetLevel(c.previousThreshold)
	os.Remove(c.path)
}

// Text is everything logged so far.
func (c *LogCapture) Text() string {
	var bytes, readError = os.ReadFile(c.path)
	if readError != nil {
		return ""
	}
	return string(bytes)
}

// HasLine is "did the demo log a line containing this text".
func (c *LogCapture) HasLine(line string) bool {
	return strings.Contains(c.Text(), line)
}

// NewFakeAdmin starts a fake admin server and hands back the Configuration that
// points at it, so a demo can be built the way its Main would build it.
func NewFakeAdmin(t *testing.T) (*server.FakeAdminServer, *core.Configuration) {
	var fake, fakeError = server.NewFakeAdminServer()
	if fakeError != nil {
		t.Fatal(fakeError.Error())
	}
	var configuration = core.NewConfiguration()
	configuration.Host = "localhost"
	configuration.AdminPort = fake.GetPort()
	configuration.Password = PASSWORD
	configuration.Name = "openttd-admin-test"
	return fake, configuration
}

// SendDate is sendDate(int day): ADMIN_PACKET_SERVER_DATE, uint32 day. The
// demos' rules read the date off the game model the event carries, so this is
// the packet that decides whether a rule greets its players.
func SendDate(fake *server.FakeAdminServer, day int) {
	fake.Send(server.New(constant.ADMIN_PACKET_SERVER_DATE).
		WriteUint32(uint32(day)))
}

// ChatPacket is sendChat(NETWORK_ACTION_CHAT, DESTTYPE_BROADCAST, clientId,
// message): what the game director sends for a public chat line. The uint32 of
// a chat packet is the client the line comes from.
func ChatPacket(clientId int, message string) *server.PacketWriter {
	return server.New(constant.ADMIN_PACKET_SERVER_CHAT).
		WriteUint8(int(constant.NETWORK_ACTION_CHAT)).
		WriteUint8(int(constant.DESTTYPE_BROADCAST)).
		WriteUint32(uint32(clientId)).
		WriteString(message).
		WriteUint64(uint64(0))
}

// ChatMessages is the text of every chat packet the fake server received, oldest
// first: the lines a demo said to its players.
func ChatMessages(fake *server.FakeAdminServer) []string {
	var messages []string
	for _, received := range fake.Received() {
		if received.GetWireTypeId() == int(constant.ADMIN_PACKET_ADMIN_CHAT) {
			var packet = received.GetPacket()
			packet.ReadUint8()  // action
			packet.ReadUint8()  // destination type
			packet.ReadUint32() // destination id
			messages = append(messages, packet.ReadString())
		}
	}
	return messages
}

// AwaitChat waits until the fake server has seen at least this many chat lines.
func AwaitChat(fake *server.FakeAdminServer, count int) {
	openttd.AwaitSize(func() int { return len(ChatMessages(fake)) }, count, 10,
		strconv.Itoa(count)+" chat lines")
}

// TestParseArgumentsReadsTheCommandLine: the two arguments every sample takes
// become the host, the admin port and the password of the Configuration.
func TestParseArgumentsReadsTheCommandLine(t *testing.T) {
	var configuration = core.NewConfiguration()
	if !demo.ParseArguments([]string{"localhost:3977", PASSWORD}, configuration) {
		t.Errorf("a good command line was rejected")
	}
	if configuration.Host != "localhost" {
		t.Errorf("host = %s", configuration.Host)
	}
	if configuration.AdminPort != 3977 {
		t.Errorf("admin port = %d", configuration.AdminPort)
	}
	if configuration.Password != PASSWORD {
		t.Errorf("password = %s", configuration.Password)
	}
}

// TestParseArgumentsReportsTheUsage: Java only catches a bad number, and lets
// the other three cases die in an exception; the port reports the usage for all
// of them and returns false, so the caller stops instead of the process dying.
func TestParseArgumentsReportsTheUsage(t *testing.T) {
	var capture = NewLogCapture()
	var bad = [][]string{
		[]string{},
		[]string{PASSWORD},
		[]string{"localhost", PASSWORD},
		[]string{"localhost:not-a-number", PASSWORD},
		[]string{"localhost:99999999999", PASSWORD}}
	for _, args := range bad {
		var configuration = core.NewConfiguration()
		if demo.ParseArguments(args, configuration) {
			t.Errorf("accepted a command line of %d arguments", len(args))
		}
	}
	if !capture.HasLine("Usage: java -jar openttd-robot.jar localhost:3977 admin_password") {
		t.Errorf("no usage line in:\n%s", capture.Text())
	}
	if !capture.HasLine("See openttd.cfg to set your server's admin_password first !") {
		t.Errorf("no second usage line in:\n%s", capture.Text())
	}
	capture.Close()
}

// TestFakeUserServiceLogsInAnyToken: the fake has no users of its own, so the
// token becomes the name, and a token starting with "admin" is the whole of its
// idea of a moderator.
func TestFakeUserServiceLogsInAnyToken(t *testing.T) {
	var service = demo.NewFakeUserService()
	var player = service.IdentifyUser("Dave")
	if player == nil {
		t.Errorf("a token of Dave logged no user")
	}
	if player.GetName() != "Dave" {
		t.Errorf("name = %s", player.GetName())
	}
	if player.IsAdmin() {
		t.Errorf("Dave was made a moderator")
	}
	var moderator = service.IdentifyUser("admin-eric")
	if !moderator.IsAdmin() {
		t.Errorf("admin-eric was not made a moderator")
	}
	// The fake never assigns an id, which Java's lambda does not either - and
	// since ExternalUser.equals() is keyed on the id alone, every user this fake
	// produces is equal to every other. The rules key their own maps by client id
	// instead, so the demos still work.
	if player.GetId() != int64(0) {
		t.Errorf("the fake assigned id %d", player.GetId())
	}
}

// TestFakeGameServiceRecordsTheEndOfTheGame: the fake game service has nothing
// to read and nothing to write, so it answers null, prints what it is handed,
// and - where Java leaves the process - records that the game is over.
func TestFakeGameServiceRecordsTheEndOfTheGame(t *testing.T) {
	var service = demo.NewFakeGameService()
	if service.GetGame(int64(1)) != nil {
		t.Errorf("the fake game service returned a game")
	}
	if service.IsEnded() {
		t.Errorf("the fake game service ended on its own")
	}
	service.EndGame(int64(1))
	if !service.IsEnded() {
		t.Errorf("EndGame did not record the end of the game")
	}
}

// TestFakeExternalUserServiceFillsTheSingleton: the demos call this before they
// build their rules, because the rules capture the services in their field
// initializers, exactly as Java's do.
func TestFakeExternalUserServiceFillsTheSingleton(t *testing.T) {
	demo.FakeExternalUserService()
	var services = robot.GetInstance()
	var userService = services.GetExternalUserService()
	var gameService = services.GetExternalGameService()
	if userService == nil {
		t.Errorf("no ExternalUserService injected")
	}
	if gameService == nil {
		t.Errorf("no ExternalGameService injected")
	}
	var logged = userService.IdentifyUser("Alice")
	if logged == nil {
		t.Errorf("the injected service identified nobody")
	}
	if logged.GetName() != "Alice" {
		t.Errorf("the injected service named the user %s", logged.GetName())
	}
	if gameService.GetGame(int64(1)) != nil {
		t.Errorf("the injected game service returned a game")
	}
}

// SendClientJoin is sendClientJoin(int clientId): ADMIN_PACKET_SERVER_CLIENT_JOIN.
func SendClientJoin(fake *server.FakeAdminServer, clientId int) {
	fake.Send(server.New(constant.ADMIN_PACKET_SERVER_CLIENT_JOIN).
		WriteUint32(uint32(clientId)))
}

// SendCompanyNew is sendCompanyNew(short companyId): ADMIN_PACKET_SERVER_COMPANY_NEW.
func SendCompanyNew(fake *server.FakeAdminServer, companyId int) {
	fake.Send(server.New(constant.ADMIN_PACKET_SERVER_COMPANY_NEW).
		WriteUint8(companyId))
}

// TestHelloWorldGreetsOnMondays: the sample rule is a DateEventListener that
// says hello once a week, so a Monday date packet must produce exactly one chat
// line and any other date none.
func TestHelloWorldGreetsOnMondays(t *testing.T) {
	var fake, configuration = NewFakeAdmin(t)
	var robotDemo = demo.NewHelloWorldDemo(configuration)
	robotDemo.Startup()
	fake.Handshake()
	// The WELCOME packet carries a date of its own, so the tests below count
	// the chat lines the fake server received from then on.
	var before = len(ChatMessages(fake))
	SendDate(fake, MONDAY_DAY)
	AwaitChat(fake, before+1)
	var messages = ChatMessages(fake)
	if messages[before] != "Hello world !!!" {
		t.Errorf("the Monday greeting was %s", messages[before])
	}
	// A Saturday must greet nobody: the sender is a thread, so "nothing" needs
	// a moment before it can be told apart from "nothing yet".
	SendDate(fake, SATURDAY_DAY)
	time.Sleep(time.Duration(SETTLE_MILLIS) * time.Millisecond)
	if len(ChatMessages(fake)) != before+1 {
		t.Errorf("a Saturday greeted %d times", len(ChatMessages(fake))-before)
	}
	robotDemo.Shutdown()
	fake.Close()
}

// TestIngameInfoBriefsThePlayersOnMondays: the login demo's own rule tells the
// players what the server can do for them, once a week, in this order.
func TestIngameInfoBriefsThePlayersOnMondays(t *testing.T) {
	demo.FakeExternalUserService()
	var fake, configuration = NewFakeAdmin(t)
	var robotDemo = demo.NewAdministrationDemo(configuration)
	robotDemo.Startup()
	fake.Handshake()
	var before = len(ChatMessages(fake))
	SendDate(fake, MONDAY_DAY)
	AwaitChat(fake, before+4)
	var messages = ChatMessages(fake)
	var expected = []string{
		"Welcome to this login demo.",
		"Type '!login username' to log on (any name will work).",
		"If your username starts with 'admin', you will have moderation powers.",
		"Once logged, type '$help' to know more about moderation commands."}
	for i := 0; i < len(expected); i++ {
		if messages[before+i] != expected[i] {
			t.Errorf("line %d was %s", i, messages[before+i])
		}
	}
	SendDate(fake, SATURDAY_DAY)
	time.Sleep(time.Duration(SETTLE_MILLIS) * time.Millisecond)
	if len(ChatMessages(fake)) != before+4 {
		t.Errorf("a Saturday briefed %d lines", len(ChatMessages(fake))-before)
	}
	robotDemo.Shutdown()
	fake.Close()
}

// TestConsoleLogsOneLinePerEvent: the console sample is the facade with nothing
// added but one log line per event kind, which is the quickest way to see what a
// server sends an admin client. The four prefixes - and the two colons the chat
// and company lines have and the date and client lines do not - are Java's.
func TestConsoleLogsOneLinePerEvent(t *testing.T) {
	var capture = NewLogCapture()
	var fake, configuration = NewFakeAdmin(t)
	var console = demo.NewOpenttdAdminConsole(configuration)
	console.AddListener(event.EVENT_DATE, console)
	console.AddListener(event.EVENT_CHAT, console)
	console.AddListener(event.EVENT_CLIENT, console)
	console.AddListener(event.EVENT_COMPANY, console)
	console.Startup()
	fake.Handshake()
	SendDate(fake, MONDAY_DAY)
	SendClientJoin(fake, 7)
	fake.Send(ChatPacket(7, "hello world"))
	SendCompanyNew(fake, 6)

	var seen = func() bool {
		if !capture.HasLine("Date event ") {
			return false
		}
		if !capture.HasLine("Client event ") {
			return false
		}
		if !capture.HasLine("Chat event : ") {
			return false
		}
		return capture.HasLine("Company event : ")
	}
	openttd.Await(seen, 10, "one log line per event kind")
	// The events print themselves, so the lines carry their text: a join is a
	// CREATE of the client, a new company a CREATE of the company.
	if !capture.HasLine("Client event CREATE:7") {
		t.Errorf("no client line in:\n%s", capture.Text())
	}
	if !capture.HasLine("Company event : CREATE:6") {
		t.Errorf("no company line in:\n%s", capture.Text())
	}
	console.Shutdown()
	fake.Close()
	capture.Close()
}

// TestAdminPingSendsOnePing: the smallest thing the library can do, connect to
// the admin port without the facade and send one ping. The fake server answers
// as the game does, with a PONG echoing the identifier, and the client has no
// game model to put it in - which is Java's behaviour too - so all this test can
// see is that the ping left and the connection survived the answer.
func TestAdminPingSendsOnePing(t *testing.T) {
	var fake, configuration = NewFakeAdmin(t)
	var pinger = demo.NewAdminPing(configuration)
	pinger.Start()
	fake.Handshake()
	openttd.Await(func() bool { return pinger.GetNetworkAdmin().GetSend() != nil }, 10,
		"the admin sender")
	pinger.Ping()
	var ping = fake.AwaitPacket(constant.ADMIN_PACKET_ADMIN_PING, 10)
	if ping == nil {
		t.Errorf("no ADMIN_PACKET_ADMIN_PING reached the server")
	}
	if fake.IsClientDisconnected() {
		t.Errorf("the server's pong ended the connection")
	}
	pinger.Shutdown()
	fake.Close()
}

// TestAdminPingWithoutAConnectionLogs: getSend() is null until a connection has
// been made, where Java's getSend().ping() raises a NullPointerException.
func TestAdminPingWithoutAConnectionLogs(t *testing.T) {
	var capture = NewLogCapture()
	var configuration = core.NewConfiguration()
	var pinger = demo.NewAdminPing(configuration)
	pinger.Ping()
	if !capture.HasLine("no admin sender yet") {
		t.Errorf("a ping with no connection logged:\n%s", capture.Text())
	}
	capture.Close()
}
