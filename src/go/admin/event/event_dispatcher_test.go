// EventDispatcher tests: the queue/thread pair that carries events from
// NetworkAdmin's thread to the application's listeners, and the down-casts that
// replaced Java's generic listener type.

package event_test

import "strconv"
import "testing"
import "com.openttd"
import "time"
import "com.openttd/admin/event"
import "com.openttd/util"

// FakeOpenttd stands in for admin/model/Game, the events' `openttd` field. The
// seam is the one method an event reads.
type FakeOpenttd struct {
	day int64
}

func (f *FakeOpenttd) GetDate() util.Calendar {
	return util.DayToCalendar(f.day)
}

type PongRecorder struct {
	ids []int64
}

func (l *PongRecorder) OnPongEvent(pongEvent *event.PongEvent) {
	l.ids = append(l.ids, pongEvent.GetPingId())
}

type ConsoleRecorder struct {
	messages []string
}

func (l *ConsoleRecorder) OnConsoleEvent(consoleEvent *event.ConsoleEvent) {
	l.messages = append(l.messages, consoleEvent.GetMessage())
}

type DateRecorder struct {
	dates []string
}

func (l *DateRecorder) OnDateEvent(dateEvent *event.DateEvent) {
	l.dates = append(l.dates, dateEvent.ToString())
}

func TestDispatchReachesTheListener(t *testing.T) {
	var dispatcher = event.NewEventDispatcher()
	dispatcher.Startup()
	var recorder = &PongRecorder{}
	dispatcher.AddListener(event.EVENT_PONG, recorder)
	var game = &FakeOpenttd{}
	dispatcher.Dispatch(event.NewPongEvent(game, int64(7)))
	openttd.Await(func() bool { return len(recorder.ids) == 1 }, 5, "the pong event")
	if recorder.ids[0] != 7 {
		t.Errorf("pong id %d, expected 7", recorder.ids[0])
	}
	dispatcher.Shutdown()
	if dispatcher.IsAlive() {
		t.Errorf("the dispatcher is still running after shutdown")
	}
}

// A listener registered under a kind it cannot handle is Java's caught
// ClassCastException: it is logged, the other listeners still get the event,
// and the dispatcher keeps running.
func TestMismatchedListenerIsSkipped(t *testing.T) {
	var dispatcher = event.NewEventDispatcher()
	dispatcher.Startup()
	var wrong = &ConsoleRecorder{}
	dispatcher.AddListener(event.EVENT_PONG, wrong)
	var right = &PongRecorder{}
	dispatcher.AddListener(event.EVENT_PONG, right)
	var game = &FakeOpenttd{}
	dispatcher.Dispatch(event.NewPongEvent(game, int64(3)))
	openttd.Await(func() bool { return len(right.ids) == 1 }, 5, "the pong event")
	if len(wrong.messages) != 0 {
		t.Errorf("a console listener received a pong event")
	}
	if !dispatcher.IsAlive() {
		t.Errorf("the dispatcher stopped after a bad listener")
	}
	dispatcher.Shutdown()
}

// RemoveListener takes the listener back out again.
func TestRemoveListener(t *testing.T) {
	var dispatcher = event.NewEventDispatcher()
	dispatcher.Startup()
	var recorder = &PongRecorder{}
	dispatcher.AddListener(event.EVENT_PONG, recorder)
	var game = &FakeOpenttd{}
	dispatcher.Dispatch(event.NewPongEvent(game, int64(1)))
	openttd.Await(func() bool { return len(recorder.ids) == 1 }, 5, "the first pong")
	dispatcher.RemoveListener(event.EVENT_PONG, recorder)
	dispatcher.Dispatch(event.NewPongEvent(game, int64(2)))
	time.Sleep(SETTLE_MILLIS * time.Millisecond)
	if len(recorder.ids) != 1 {
		t.Errorf("the removed listener saw %d events", len(recorder.ids))
	}
	dispatcher.Shutdown()
}

// Nothing is lost, and nothing blocks, when an event has no listeners at all:
// that is the state a library user leaves the facade in.
func TestDispatchWithoutListeners(t *testing.T) {
	var dispatcher = event.NewEventDispatcher()
	dispatcher.Startup()
	var game = &FakeOpenttd{}
	dispatcher.Dispatch(event.NewDateEvent(game))
	dispatcher.Dispatch(event.NewChatEvent(game, int64(7), "hello"))
	dispatcher.Dispatch(event.NewPongEvent(game, int64(1)))
	openttd.Await(func() bool { return dispatcher.QueueSize() == 0 }, 5, "the queue to drain")
	if !dispatcher.IsAlive() {
		t.Errorf("the dispatcher stopped with no listeners")
	}
	dispatcher.Shutdown()
}

// Java's startup() clears the queue and starts a fresh thread, and shutdown()
// clears it again: events nobody consumed are dropped, not replayed. The second
// startup also needs EventQueue.Open(), the undo of the Close() shutdown does.
func TestStartupClearsTheQueue(t *testing.T) {
	var dispatcher = event.NewEventDispatcher()
	var game = &FakeOpenttd{}
	dispatcher.Dispatch(event.NewDateEvent(game))
	if dispatcher.QueueSize() != 1 {
		t.Errorf("queue holds %d before startup, expected 1", dispatcher.QueueSize())
	}
	dispatcher.Startup()
	if dispatcher.QueueSize() != 0 {
		t.Errorf("startup left %d events queued", dispatcher.QueueSize())
	}
	dispatcher.Shutdown()
	if dispatcher.IsAlive() {
		t.Errorf("the dispatcher is still running after shutdown")
	}
	var recorder = &DateRecorder{}
	dispatcher.Startup()
	dispatcher.AddListener(event.EVENT_DATE, recorder)
	dispatcher.Dispatch(event.NewDateEvent(game))
	openttd.Await(func() bool { return len(recorder.dates) == 1 }, 5, "an event after a second startup")
	dispatcher.Shutdown()
}

// Every event prints the way its Java class does.
func TestEventToStrings(t *testing.T) {
	var game = &FakeOpenttd{}
	game.day = int64(1950*365 + 1)
	var chat = event.NewChatEvent(game, int64(7), "hello")
	if chat.ToString() != "7 says hello" {
		t.Errorf("chat event %s", chat.ToString())
	}
	var client = event.NewClientEvent(game, int64(7), event.ACTION_CREATE)
	if client.ToString() != "CREATE:7" {
		t.Errorf("client event %s", client.ToString())
	}
	var company = event.NewCompanyEvent(game, 2, event.ACTION_DELETE)
	if company.ToString() != "DELETE:2" {
		t.Errorf("company event %s", company.ToString())
	}
	var console = event.NewConsoleEvent(game, "::1", "say hi")
	if console.ToString() != "Console (::1) say hi" {
		t.Errorf("console event %s", console.ToString())
	}
	var script = event.NewGameScriptEvent(game, "{\"id\":\"0\"}")
	if script.ToString() != "GameScript {\"id\":\"0\"}" {
		t.Errorf("gamescript event %s", script.ToString())
	}
	var pong = event.NewPongEvent(game, int64(4))
	if pong.ToString() != "Pong (4)" {
		t.Errorf("pong event %s", pong.ToString())
	}
	var rcon = event.NewRConEvent(game, 2, "output")
	if rcon.ToString() != "RCon (2) output" {
		t.Errorf("rcon event %s", rcon.ToString())
	}
	var rconEnd = event.NewRConEndEvent(game, "getdate")
	if rconEnd.ToString() != "RCon ended :getdate" {
		t.Errorf("rcon end event %s", rconEnd.ToString())
	}
	// DateEvent is day-month-year, Month 0 based, read through the seam. The
	// game's day counter is not 365 days a year, so this is January 1st of
	// game year 1950, the date a new map reports in its WELCOME packet.
	var year = 1950
	var startOfYear = year*365 + ((year-1)/4 - (year-1)/100 + (year-1)/400) + 1
	game.day = int64(startOfYear)
	var date = util.DayToCalendar(int64(startOfYear))
	var expected = "DateEvent " + strconv.Itoa(date.DayOfMonth) + "-" + strconv.Itoa(date.Month) + "-" + strconv.Itoa(date.Year)
	if event.NewDateEvent(game).ToString() != expected {
		t.Errorf("date event %s, expected %s", event.NewDateEvent(game).ToString(), expected)
	}
	if event.NewDateEvent(game).ToString() != "DateEvent 1-0-1950" {
		t.Errorf("date event %s, expected DateEvent 1-0-1950", event.NewDateEvent(game).ToString())
	}
	if date.Year != 1950 || date.Month != util.JANUARY || date.DayOfMonth != 1 {
		t.Errorf("game date 1950 reads %d-%d-%d", date.DayOfMonth, date.Month, date.Year)
	}
}

// The kind table is the stand-in for Java's Class keys, so it has to cover every
// event class and name it the way the Java package does.
func TestEventKindTable(t *testing.T) {
	var kinds = event.EventKindValues()
	if len(kinds) != 9 {
		t.Errorf("%d event kinds, expected 9", len(kinds))
	}
	var names = []string{
		"ChatEvent", "ClientEvent", "CompanyEvent", "ConsoleEvent",
		"DateEvent", "GameScriptEvent", "PongEvent", "RConEvent", "RConEndEvent"}
	for i := 0; i < len(kinds); i++ {
		if event.EventKindName(kinds[i]) != names[i] {
			t.Errorf("kind %d is %s, expected %s", kinds[i], event.EventKindName(kinds[i]), names[i])
		}
	}
	if event.EventKindName(event.EVENT_NONE) != "none" {
		t.Errorf("the sentinel is %s", event.EventKindName(event.EVENT_NONE))
	}
	if event.EventKindName(9) != "EVENT_UNKNOWN" {
		t.Errorf("an out of range kind is %s", event.EventKindName(9))
	}
	if event.ActionName(event.ACTION_CREATE) != "CREATE" || event.ActionName(event.ACTION_UPDATE) != "UPDATE" || event.ActionName(event.ACTION_DELETE) != "DELETE" {
		t.Errorf("the action names are wrong")
	}
}

const SETTLE_MILLIS = 200
