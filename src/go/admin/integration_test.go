// Integration ladder for the admin port against a real OpenTTD dedicated server,
// ported from test/java/com/openttd/admin/OpenttdAdminIntegrationTest.java. The
// helpers below are shared with protocol_integration_test.go, the port of
// OpenttdAdminProtocolIntegrationTest.java.
//
// Both Java classes extend OpenttdServerWrapper, which starts a dedicated server
// only when port 3977 is free. The Go ladder never spawns one: `go test` has no
// @BeforeAll/@AfterAll pair that could own the process for the whole run, and the
// Java notes already say pre-starting stays the reliable option. So every test
// here asks server.CheckAdminPort() and skips when nothing listens - see the
// AGENTS.md integrationTest prerequisites for how to start the server.
//
// Java's TestUtils.await fails the test on a timeout. openttd.Await panics, which
// would abort the whole binary, so awaitTrue below reports through t.Errorf and
// stops only the running test; tryAwait is the polling half with no reporting.
// Callers write `if (!awaitTrue(...)) { return }` so a timed out wait never lets
// the rest of the test read a value that has not arrived.

package admin_test

import "regexp"
import "strconv"
import "testing"
import "time"
import "com.openttd"
import "com.openttd/admin"
import "com.openttd/admin/event"
import "com.openttd/gamescript"
import "com.openttd/network/core"
import "com.openttd/server"

// WELCOME_TIMEOUT_SECONDS is the Java setUp's wait for the WELCOME packet.
const WELCOME_TIMEOUT_SECONDS = 20

// IntegrationFixture is the pair every integration test works with: the facade and
// the EventRecorder registered on it. Java holds the two as fields filled by setUp().
type IntegrationFixture struct {
	admin    *admin.OpenttdAdmin
	recorder *server.EventRecorder
}

// integrationAdmin is Java's setUp() without the WELCOME wait: a configuration with
// the secrets file's password, an EventRecorder registered on the facade, startup().
// It returns nil after skipping when no server listens, so the caller must check.
func integrationAdmin(t *testing.T, password string, name string) *IntegrationFixture {
	if !server.CheckAdminPort() {
		t.Skipf("no OpenTTD dedicated server listening on port %d: start it with "+
			"openttd -D -d 9,console=1,sprite=0,grf=0,sl=0 -c openttd/openttd.cfg", server.ADMIN_PORT)
		return nil
	}

	var configuration = core.NewConfiguration()
	configuration.Password = password
	configuration.Name = name

	var fixture = &IntegrationFixture{
		admin:    admin.NewOpenttdAdmin(configuration),
		recorder: server.NewEventRecorder(),
	}
	server.RegisterWith(fixture.admin, fixture.recorder)
	fixture.admin.Startup()
	// Java's tearDown() / finally block.
	t.Cleanup(fixture.admin.Shutdown)
	return fixture
}

// integrationFixture is setUp(): startup() plus the wait for the WELCOME packet, then
// recorder.Clear(). Game.GetGameInfo() is never nil in Go, so Java's null test becomes
// the empty server name; the game model itself is nil until startup() has built it.
func integrationFixture(t *testing.T, password string, name string) *IntegrationFixture {
	var fixture = integrationAdmin(t, password, name)
	if fixture == nil {
		return nil
	}
	if !awaitTrue(t, func() bool { return welcomeReceived(fixture.admin) }, WELCOME_TIMEOUT_SECONDS,
		"the WELCOME packet") {
		return nil
	}
	fixture.recorder.Clear()
	return fixture
}

func secretsOrDefault(t *testing.T) string {
	var properties, err = openttd.ReadSecrets()
	if err != nil {
		t.Fatalf("cannot read the admin password: %s", core.Message(err))
		return ""
	}
	return properties.GetProperty("admin_password")
}

func welcomeReceived(adminInstance *admin.OpenttdAdmin) bool {
	var game = adminInstance.GetGame()
	if game == nil {
		return false
	}
	var gameInfo = game.GetGameInfo()
	if gameInfo == nil {
		return false
	}
	return gameInfo.GetServerName() != ""
}

// tryAwait is TestUtils.await minus the failure: poll until the condition holds or the
// timeout runs out.
func tryAwait(condition func() bool, timeoutSeconds int) bool {
	var waited = 0
	var limit = timeoutSeconds * 1000
	for waited < limit {
		if condition() {
			return true
		}
		time.Sleep(openttd.POLL_MILLIS * time.Millisecond)
		waited = waited + openttd.POLL_MILLIS
	}
	return false
}

// awaitTrue reports the timeout the way TestUtils.await does, then stops the test.
func awaitTrue(t *testing.T, condition func() bool, timeoutSeconds int, description string) bool {
	if tryAwait(condition, timeoutSeconds) {
		return true
	}
	t.Errorf("timed out after %d s waiting for %s", timeoutSeconds, description)
	t.FailNow()
	return false
}

// awaitSizeTrue is TestUtils.awaitSize: count() must reach size.
func awaitSizeTrue(t *testing.T, count func() int, size int, timeoutSeconds int, description string) bool {
	if !awaitTrue(t, func() bool { return count() >= size }, timeoutSeconds, description) {
		return false
	}
	return true
}

// matchesPattern is Java's String.matches() on a whole string. regexp.MatchString
// returns an error as well as the answer, so a bad pattern is reported as no match.
func matchesPattern(pattern string, text string) bool {
	var matched, err = regexp.MatchString(pattern, text)
	if err != nil {
		return false
	}
	return matched
}

/* The three tests of OpenttdAdminIntegrationTest.java */

// facadeListener is Java's inner OpenttdTestAdmin: the facade cannot be extended in Go,
// so the listener is its own object registered through AddListener. The dispatcher's
// Notify() type-asserts each listener to the interface its event kind needs.
type facadeListener struct {
	pongs       []*event.PongEvent
	rConEnds    []*event.RConEndEvent
	gameScripts []*event.GameScriptEvent
}

func (l *facadeListener) OnPongEvent(pongEvent *event.PongEvent) {
	l.pongs = append(l.pongs, pongEvent)
}

func (l *facadeListener) OnRConEndEvent(rConEndEvent *event.RConEndEvent) {
	l.rConEnds = append(l.rConEnds, rConEndEvent)
}

func (l *facadeListener) OnGameScriptEvent(gameScriptEvent *event.GameScriptEvent) {
	l.gameScripts = append(l.gameScripts, gameScriptEvent)
}

func (l *facadeListener) clear() {
	l.pongs = []*event.PongEvent{}
	l.rConEnds = []*event.RConEndEvent{}
	l.gameScripts = []*event.GameScriptEvent{}
}

// TestPingPong is testPingPong(): ADMIN_PACKET_ADMIN_PING answers with a PONG.
func TestPingPong(t *testing.T) {
	var listener = &facadeListener{}
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin test")
	if fixture == nil {
		return
	}
	fixture.admin.AddListener(event.EVENT_PONG, listener)

	var send = fixture.admin.GetSend()
	if send == nil {
		t.Fatal("a sender once connected")
		return
	}
	for index := int64(0); index < 5; index++ {
		send.Ping(index)
	}
	if !awaitSizeTrue(t, func() int { return len(listener.pongs) }, 5, 10, "five PongEvents") {
		return
	}
	if len(listener.pongs) != 5 {
		t.Errorf("five PongEvents, got %d", len(listener.pongs))
	}
}

// TestRConIsAnswered is testRCon(): every rcon command ends with an RCON_END.
func TestRConIsAnswered(t *testing.T) {
	var listener = &facadeListener{}
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin test")
	if fixture == nil {
		return
	}
	fixture.admin.AddListener(event.EVENT_RCON_END, listener)

	var send = fixture.admin.GetSend()
	if send == nil {
		t.Fatal("a sender once connected")
		return
	}
	for index := int64(0); index < 5; index++ {
		var command = "say message " + strconv.FormatInt(index, 10)
		send.Rcon(command)
		if !awaitSizeTrue(t, func() int { return len(listener.rConEnds) }, int(index)+1, 10,
			"the RConEndEvent of '"+command+"'") {
			return
		}
	}
	if len(listener.rConEnds) != 5 {
		t.Errorf("five RConEndEvents, got %d", len(listener.rConEnds))
	}
}

// TestGameScriptNewsReplies is testGameScript(): AdminCmd answers both news papers, and
// the second one is refused because company 0 does not exist on a dedicated server.
func TestGameScriptNewsReplies(t *testing.T) {
	var listener = &facadeListener{}
	var fixture = integrationFixture(t, secretsOrDefault(t), "openttd-admin test")
	if fixture == nil {
		return
	}
	fixture.admin.AddListener(event.EVENT_GAME_SCRIPT, listener)

	var executor = fixture.admin.GetGSExecutor()
	if executor == nil {
		t.Fatal("an executor once connected")
		return
	}

	executor.Send(gamescript.NewGSNewsPaper(gamescript.NT_GENERAL, "News Broadcasting"))
	if !awaitSizeTrue(t, func() int { return len(listener.gameScripts) }, 1, 15, "the broadcast news reply") {
		return
	}
	executor.Send(gamescript.NewGSNewsPaperForCompany(0, gamescript.NT_GENERAL, "News for company 0"))
	if !awaitSizeTrue(t, func() int { return len(listener.gameScripts) }, 2, 15, "the company news reply") {
		return
	}
	if len(listener.gameScripts) != 2 {
		t.Errorf("two GameScriptEvents, got %d", len(listener.gameScripts))
	}
}

// The listener lists are cleared by setUp(), which integrationFixture does not do: the
// facade listeners are added after it, so nothing can arrive before the test begins.
