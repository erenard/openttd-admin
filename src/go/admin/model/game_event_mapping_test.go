// GameEventMappingTest mirrors test/java/com/openttd/admin/model/GameEventMappingTest.java:
// the completeness table for Game.updateAdmin. Every ADMIN_PACKET_SERVER_* id is
// listed with the events it must produce, so a packet that is decoded but forgotten
// (the switch had no case for ADMIN_PACKET_SERVER_COMPANY_STATS, and
// ADMIN_PACKET_SERVER_RCON_END fell through into ADMIN_PACKET_SERVER_CONSOLE) fails
// here.
//
// Java spells a table row as a String, "ChatEvent x1" or "none", and parses it back
// apart; the port keeps the kind and the count in a struct and renders the Java text
// only in the failure messages. The action Java's string carries ("ClientEvent CREATE
// x1") is not asserted by the Java test either - assertEvents() only reads the class
// name and the count - so the port leaves it out of the table and checks it in
// GameTest's own per-event assertions.

package model_test

import "strconv"
import "time"
import "testing"
import "com.openttd"
import "com.openttd/admin/event"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/server"

// SETTLE_MILLIS is how long to wait before asserting a packet dispatched nothing.
// Java asserts immediately, which only passes because the dispatcher thread is
// faster than the assertion; waiting a little makes the "none" rows mean something.
const SETTLE_MILLIS = 200

// Expected is one row of the mapping table. Java spells the action out in the
// string ("ClientEvent CREATE x1") but assertEvents() never read it; the port
// keeps it in the row and does check it, because NetworkAdminServerPacketsTest
// asserts the action of every client and company event.
type Expected struct {
	kind   event.EventKind
	count  int
	action event.Action
}

// None is the row for a packet that dispatches nothing.
func None() Expected {
	return Expected{kind: event.EVENT_NONE, count: 0, action: event.ACTION_NONE}
}

// Once is the row for a packet that dispatches one event of kind, with an
// action the row does not constrain.
func Once(kind event.EventKind) Expected {
	return Expected{kind: kind, count: 1, action: event.ACTION_NONE}
}

// WithAction is the row for a packet that dispatches one event of kind with
// that action.
func WithAction(kind event.EventKind, action event.Action) Expected {
	return Expected{kind: kind, count: 1, action: action}
}

// Describe is the "EventClass x count" text Java stores in the table, action
// included when the row pins one.
func Describe(expected Expected) string {
	if expected.kind == event.EVENT_NONE {
		return "none"
	}
	var text = event.EventKindName(expected.kind)
	if expected.action != event.ACTION_NONE {
		text = text + " " + event.ActionName(expected.action)
	}
	return text + " x" + strconv.Itoa(expected.count)
}

// ExpectedEvents is expectedEvents(): the events each packet type must produce.
func ExpectedEvents() map[constant.PacketServerType]Expected {
	var table = map[constant.PacketServerType]Expected{}
	table[constant.ADMIN_PACKET_SERVER_ERROR] = None()
	table[constant.ADMIN_PACKET_SERVER_PROTOCOL] = None()
	table[constant.ADMIN_PACKET_SERVER_WELCOME] = None()
	table[constant.ADMIN_PACKET_SERVER_NEWGAME] = None()
	table[constant.ADMIN_PACKET_SERVER_DATE] = Once(event.EVENT_DATE)
	table[constant.ADMIN_PACKET_SERVER_CLIENT_JOIN] = WithAction(event.EVENT_CLIENT, event.ACTION_CREATE)
	table[constant.ADMIN_PACKET_SERVER_CLIENT_INFO] = WithAction(event.EVENT_CLIENT, event.ACTION_UPDATE)
	table[constant.ADMIN_PACKET_SERVER_CLIENT_UPDATE] = WithAction(event.EVENT_CLIENT, event.ACTION_UPDATE)
	table[constant.ADMIN_PACKET_SERVER_CLIENT_QUIT] = WithAction(event.EVENT_CLIENT, event.ACTION_DELETE)
	table[constant.ADMIN_PACKET_SERVER_CLIENT_ERROR] = WithAction(event.EVENT_CLIENT, event.ACTION_DELETE)
	table[constant.ADMIN_PACKET_SERVER_COMPANY_NEW] = WithAction(event.EVENT_COMPANY, event.ACTION_CREATE)
	table[constant.ADMIN_PACKET_SERVER_COMPANY_INFO] = WithAction(event.EVENT_COMPANY, event.ACTION_UPDATE)
	table[constant.ADMIN_PACKET_SERVER_COMPANY_UPDATE] = WithAction(event.EVENT_COMPANY, event.ACTION_UPDATE)
	table[constant.ADMIN_PACKET_SERVER_COMPANY_REMOVE] = WithAction(event.EVENT_COMPANY, event.ACTION_DELETE)
	table[constant.ADMIN_PACKET_SERVER_COMPANY_ECONOMY] = WithAction(event.EVENT_COMPANY, event.ACTION_UPDATE)
	table[constant.ADMIN_PACKET_SERVER_COMPANY_STATS] = WithAction(event.EVENT_COMPANY, event.ACTION_UPDATE)
	table[constant.ADMIN_PACKET_SERVER_CHAT] = Once(event.EVENT_CHAT)
	table[constant.ADMIN_PACKET_SERVER_RCON] = Once(event.EVENT_RCON)
	table[constant.ADMIN_PACKET_SERVER_CONSOLE] = Once(event.EVENT_CONSOLE)
	table[constant.ADMIN_PACKET_SERVER_CMD_NAMES] = None()
	table[constant.ADMIN_PACKET_SERVER_CMD_LOGGING_OLD] = None()
	table[constant.ADMIN_PACKET_SERVER_CMD_LOGGING] = None()
	table[constant.ADMIN_PACKET_SERVER_GAMESCRIPT] = Once(event.EVENT_GAME_SCRIPT)
	table[constant.ADMIN_PACKET_SERVER_RCON_END] = Once(event.EVENT_RCON_END)
	table[constant.ADMIN_PACKET_SERVER_PONG] = Once(event.EVENT_PONG)
	return table
}

// SessionEnding is SESSION_ENDING: packets that close the session, covered for
// decoding by network/admin's own test.
func SessionEnding() []constant.PacketServerType {
	return []constant.PacketServerType{
		constant.ADMIN_PACKET_SERVER_FULL,
		constant.ADMIN_PACKET_SERVER_BANNED,
		constant.ADMIN_PACKET_SERVER_SHUTDOWN,
		constant.ADMIN_PACKET_SERVER_AUTH_REQUEST,
		constant.ADMIN_PACKET_SERVER_ENABLE_ENCRYPTION,
		constant.INVALID_ADMIN_PACKET,
	}
}

func IsSessionEnding(packetType constant.PacketServerType) bool {
	var ending = SessionEnding()
	for i := 0; i < len(ending); i++ {
		if packetType == ending[i] {
			return true
		}
	}
	return false
}

// CountOf is countOf(): how often the recorder saw that event kind.
func CountOf(r *server.EventRecorder, kind event.EventKind) int {
	switch kind {
	case event.EVENT_CHAT:
		return len(r.GetChatEvents())
	case event.EVENT_CLIENT:
		return len(r.GetClientEvents())
	case event.EVENT_COMPANY:
		return len(r.GetCompanyEvents())
	case event.EVENT_CONSOLE:
		return len(r.GetConsoleEvents())
	case event.EVENT_DATE:
		return len(r.GetDateEvents())
	case event.EVENT_GAME_SCRIPT:
		return len(r.GetGameScriptEvents())
	case event.EVENT_PONG:
		return len(r.GetPongEvents())
	case event.EVENT_RCON:
		return len(r.GetRConEvents())
	case event.EVENT_RCON_END:
		return len(r.GetRConEndEvents())
	default:
		return -1
	}
}

// EventsOf is the recorder's list for an event kind, widened to the Event
// interface so one helper can serve every kind.
func EventsOf(r *server.EventRecorder, kind event.EventKind) []event.Event {
	var out []event.Event
	switch kind {
	case event.EVENT_CHAT:
		for _, item := range r.GetChatEvents() {
			out = append(out, item)
		}
	case event.EVENT_CLIENT:
		for _, item := range r.GetClientEvents() {
			out = append(out, item)
		}
	case event.EVENT_COMPANY:
		for _, item := range r.GetCompanyEvents() {
			out = append(out, item)
		}
	case event.EVENT_CONSOLE:
		for _, item := range r.GetConsoleEvents() {
			out = append(out, item)
		}
	case event.EVENT_DATE:
		for _, item := range r.GetDateEvents() {
			out = append(out, item)
		}
	case event.EVENT_GAME_SCRIPT:
		for _, item := range r.GetGameScriptEvents() {
			out = append(out, item)
		}
	case event.EVENT_PONG:
		for _, item := range r.GetPongEvents() {
			out = append(out, item)
		}
	case event.EVENT_RCON:
		for _, item := range r.GetRConEvents() {
			out = append(out, item)
		}
	case event.EVENT_RCON_END:
		for _, item := range r.GetRConEndEvents() {
			out = append(out, item)
		}
	default:
	}
	return out
}

// TotalEvents is totalEvents().
func TotalEvents(r *server.EventRecorder) int {
	var total = 0
	for _, kind := range event.EventKindValues() {
		total = total + CountOf(r, kind)
	}
	return total
}

// MinimalPacket is minimalPacket(): the smallest valid payload for a packet type,
// per tcp_admin.h.
func MinimalPacket(packetType constant.PacketServerType) *server.PacketWriter {
	var packet = server.New(packetType)
	switch packetType {
	case constant.ADMIN_PACKET_SERVER_ERROR:
		return packet.WriteUint8(0)
	case constant.ADMIN_PACKET_SERVER_NEWGAME:
		return packet // No payload
	case constant.ADMIN_PACKET_SERVER_PROTOCOL:
		return server.ProtocolPacket()
	case constant.ADMIN_PACKET_SERVER_WELCOME:
		return server.WelcomePacket()
	case constant.ADMIN_PACKET_SERVER_DATE:
		return packet.WriteUint32(uint32(1950*365 + 1))
	case constant.ADMIN_PACKET_SERVER_CLIENT_JOIN, constant.ADMIN_PACKET_SERVER_CLIENT_QUIT:
		return packet.WriteUint32(uint32(7))
	case constant.ADMIN_PACKET_SERVER_CLIENT_INFO:
		return packet.WriteUint32(uint32(7)).WriteString("::1").WriteString("Client").WriteUint8(0).WriteUint32(uint32(1)).WriteUint8(255)
	case constant.ADMIN_PACKET_SERVER_CLIENT_UPDATE:
		return packet.WriteUint32(uint32(7)).WriteString("Client").WriteUint8(255)
	case constant.ADMIN_PACKET_SERVER_CLIENT_ERROR:
		return packet.WriteUint32(uint32(7)).WriteUint8(1)
	case constant.ADMIN_PACKET_SERVER_COMPANY_NEW:
		return packet.WriteUint8(6)
	case constant.ADMIN_PACKET_SERVER_COMPANY_INFO:
		return packet.WriteUint8(6).WriteString("Company").WriteString("Manager").WriteUint8(1).WriteBool8(false).WriteUint32(uint32(1950)).WriteBool8(false).WriteUint8(0).WriteUint8(255).WriteUint8(255).WriteUint8(255).WriteUint8(255)
	case constant.ADMIN_PACKET_SERVER_COMPANY_UPDATE:
		return packet.WriteUint8(6).WriteString("Company").WriteString("Manager").WriteUint8(1).WriteBool8(false).WriteUint8(0).WriteUint8(255).WriteUint8(255).WriteUint8(255).WriteUint8(255)
	case constant.ADMIN_PACKET_SERVER_COMPANY_REMOVE:
		return packet.WriteUint8(6).WriteUint8(0)
	case constant.ADMIN_PACKET_SERVER_COMPANY_ECONOMY:
		return packet.WriteUint8(6).WriteInt64(int64(0)).WriteUint64(uint64(0)).WriteInt64(int64(0)).WriteUint16(0).WriteUint64(uint64(0)).WriteUint16(0).WriteUint16(0).WriteUint64(uint64(0)).WriteUint16(0).WriteUint16(0)
	case constant.ADMIN_PACKET_SERVER_COMPANY_STATS:
		packet.WriteUint8(6)
		for i := 0; i < constant.NETWORK_VEHICLE_TYPE_LENGTH; i++ {
			packet.WriteUint16(0)
		}
		for i := 0; i < constant.NETWORK_VEHICLE_TYPE_LENGTH; i++ {
			packet.WriteUint16(0)
		}
		return packet
	case constant.ADMIN_PACKET_SERVER_CHAT:
		return packet.WriteUint8(3).WriteUint8(0).WriteUint32(uint32(1)).WriteString("chat").WriteUint64(uint64(0))
	case constant.ADMIN_PACKET_SERVER_RCON:
		return packet.WriteUint16(2).WriteString("output")
	case constant.ADMIN_PACKET_SERVER_CONSOLE:
		return packet.WriteString("console").WriteString("text")
	case constant.ADMIN_PACKET_SERVER_CMD_NAMES:
		return packet.WriteBool8(false)
	case constant.ADMIN_PACKET_SERVER_CMD_LOGGING_OLD:
		return packet.WriteUint32(uint32(1)).WriteUint8(0).WriteUint16(1).WriteUint32(uint32(0)).WriteUint32(uint32(0)).WriteUint32(uint32(0)).WriteString("text").WriteUint32(uint32(0))
	case constant.ADMIN_PACKET_SERVER_CMD_LOGGING:
		return packet.WriteUint32(uint32(1)).WriteUint8(0).WriteUint16(1).WriteBytes([]byte{1}).WriteUint32(uint32(0))
	case constant.ADMIN_PACKET_SERVER_GAMESCRIPT:
		return packet.WriteString("{\"id\":\"0\",\"data\":\"1\"}")
	case constant.ADMIN_PACKET_SERVER_RCON_END:
		return packet.WriteString("getdate")
	case constant.ADMIN_PACKET_SERVER_PONG:
		return packet.WriteUint32(uint32(1))
	default:
		return packet // Session ending, so no payload to build.
	}
}

// NewHarness is the @BeforeEach: a fresh facade on a fresh fake server. The caller
// owns the @AfterEach, so it must call Close() - a helper cannot defer it, that
// would close the harness before the test body ran.
func NewHarness(t *testing.T) *server.AdminHarness {
	var harness, harnessError = server.NewAdminHarness()
	if harnessError != nil {
		t.Errorf("AdminHarness: %s", core.Message(harnessError))
		return nil
	}
	return harness
}

// TestTableCoversEveryPacketType is testTableCoversEveryPacketType.
func TestTableCoversEveryPacketType(t *testing.T) {
	var harness = NewHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var table = ExpectedEvents()
	for _, packetType := range constant.PacketServerTypeValues() {
		var _, has = table[packetType]
		if !has && !IsSessionEnding(packetType) {
			t.Errorf("%s is neither mapped nor session ending", constant.PacketServerTypeName(packetType))
		}
	}
}

// TestEventMapping is testEventMapping(): each packet produces exactly the events
// the table promises.
func TestEventMapping(t *testing.T) {
	var harness = NewHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var table = ExpectedEvents()
	for packetType, expected := range table {
		harness.GetRecorder().Clear()
		harness.GetServer().Send(MinimalPacket(packetType))
		AssertEvents(t, harness.GetRecorder(), packetType, expected)
	}
}

// HasAction is what ClientEvent and CompanyEvent have in common, so one check
// covers both kinds of the mapping table's action column.
type HasAction interface {
	GetAction() event.Action
}

// AssertEvents is assertEvents(): wait for the promised events, then insist that
// exactly one event class fired, exactly `count` times, and - where the row pins
// one - with the action Game.updateAdmin is supposed to give it. A packet that dispatches
// an extra event (RCON_END falling through into CONSOLE) or none (COMPANY_STATS
// having no case) fails here.
func AssertEvents(t *testing.T, r *server.EventRecorder, packetType constant.PacketServerType, expected Expected) {
	if expected.count > 0 {
		var description = constant.PacketServerTypeName(packetType) + " -> " + Describe(expected)
		openttd.Await(func() bool { return CountOf(r, expected.kind) >= expected.count }, 5, description)
	} else {
		time.Sleep(SETTLE_MILLIS * time.Millisecond)
	}
	for _, kind := range event.EventKindValues() {
		var want = 0
		if kind == expected.kind {
			want = expected.count
		}
		var got = CountOf(r, kind)
		if got != want {
			t.Errorf("%s %s events: expected %d, was %d", constant.PacketServerTypeName(packetType), event.EventKindName(kind), want, got)
		}
	}
	if expected.action != event.ACTION_NONE {
		for _, item := range EventsOf(r, expected.kind) {
			var withAction, ok = item.(HasAction)
			if !ok {
				t.Errorf("%s: %s carries no action", constant.PacketServerTypeName(packetType), event.EventKindName(expected.kind))
				continue
			}
			if withAction.GetAction() != expected.action {
				t.Errorf("%s: expected a %s action, was %s",
					constant.PacketServerTypeName(packetType),
					event.ActionName(expected.action),
					event.ActionName(withAction.GetAction()))
			}
		}
	}
}

// TestSessionEndingPacketsDispatchNothing is testSessionEndingPacketsDispatchNothing.
func TestSessionEndingPacketsDispatchNothing(t *testing.T) {
	var harness = NewHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	for _, packetType := range SessionEnding() {
		if packetType == constant.INVALID_ADMIN_PACKET {
			continue // Not a packet id a server sends
		}
		harness.GetRecorder().Clear()
		harness.GetServer().Send(server.New(packetType))
		harness.GetServer().AwaitClientDisconnect(5)
		if TotalEvents(harness.GetRecorder()) != 0 {
			t.Errorf("%s must dispatch no event", constant.PacketServerTypeName(packetType))
		}
	}
}

// TestWelcomeIsNotAnEvent is Java's testWelcomeIsNotAnEvent() and
// testWelcomeRegistersUpdates() in one: WELCOME registers the eight update
// frequencies and the four polls Game.updateAdmin asks for, and it is not a
// functional event.
func TestWelcomeIsNotAnEvent(t *testing.T) {
	var harness = NewHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()
	var frequencies = len(srv.Packets(constant.ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY))
	var polls = len(srv.Packets(constant.ADMIN_PACKET_ADMIN_POLL))
	harness.GetRecorder().Clear()
	srv.Send(server.WelcomePacket())
	openttd.AwaitSize(func() int { return len(srv.Packets(constant.ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY)) }, frequencies+8, 5, "the eight update frequency registrations")
	openttd.AwaitSize(func() int { return len(srv.Packets(constant.ADMIN_PACKET_ADMIN_POLL)) }, polls+4, 5, "the four polls")
	if TotalEvents(harness.GetRecorder()) != 0 {
		t.Errorf("WELCOME is protocol handshake, not game state")
	}
}
