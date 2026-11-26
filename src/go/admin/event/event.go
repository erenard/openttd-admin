// event mirrors com.openttd.admin.event: the functional events the admin facade
// hands to the application, translated out of the technical packets NetworkAdmin
// decodes.
//
// Java's Event<L> is generic over its listener type, so notify() calls
// listener.on*Event() with no cast, and EventDispatcher keys its listener map by
// event.getClass(). This dialect has neither generics nor Class objects, so an
// Event carries an EventKind - the stand-in for that Class - and each event's
// Notify down-casts the EventListener to its own listener interface. That cast
// is where Java's ClassCastException used to happen; here a failed cast is
// logged, which is what EventDispatcher.run()'s catch did with it.
//
// Every event also keeps Java's `openttd` field, an admin/model/Game. That class
// imports this package, so naming it here would be an import cycle: Openttd below
// is the seam instead, narrowed to the one thing the events themselves read.

package event

import "com.openttd/util"

// EventListener is the marker interface every listener implements. It is empty
// in Java too; its only job is to be the type the dispatcher stores.
type EventListener interface{}

// EventKind is the port's stand-in for the Class keys of the dispatcher's
// listener map: one id per event class, in the order the classes are listed in
// the Java package.
type EventKind int

const (
	EVENT_CHAT        EventKind = 0
	EVENT_CLIENT      EventKind = 1
	EVENT_COMPANY     EventKind = 2
	EVENT_CONSOLE     EventKind = 3
	EVENT_DATE        EventKind = 4
	EVENT_GAME_SCRIPT EventKind = 5
	EVENT_PONG        EventKind = 6
	EVENT_RCON        EventKind = 7
	EVENT_RCON_END    EventKind = 8
)

// EVENT_NONE is the sentinel the mapping table uses for "this packet dispatches
// nothing"; no event reports this kind.
const EVENT_NONE EventKind = -1

// EventKindValues is the kinds that can be dispatched, in order, sentinel
// excluded: the list the completeness assertions walk over.
func EventKindValues() []EventKind {
	return []EventKind{
		EVENT_CHAT,
		EVENT_CLIENT,
		EVENT_COMPANY,
		EVENT_CONSOLE,
		EVENT_DATE,
		EVENT_GAME_SCRIPT,
		EVENT_PONG,
		EVENT_RCON,
		EVENT_RCON_END,
	}
}

// eventKindNames is the Java class's simple name for each EventKindValues()
// entry: what the test table spells out, and what the dispatcher logs.
var eventKindNames = []string{
	"ChatEvent",
	"ClientEvent",
	"CompanyEvent",
	"ConsoleEvent",
	"DateEvent",
	"GameScriptEvent",
	"PongEvent",
	"RConEvent",
	"RConEndEvent",
}

// EventKindName is that class's simple name.
func EventKindName(kind EventKind) string {
	if kind == EVENT_NONE {
		return "none"
	}
	if kind < 0 || int(kind) >= len(eventKindNames) {
		return "EVENT_UNKNOWN"
	}
	return eventKindNames[int(kind)]
}

// Event is Event<L>: the kind it belongs to, and the dispatch to a listener of
// that kind.
type Event interface {
	Kind() EventKind
	Notify(listener EventListener)
}

// Openttd is the seam for the events' `openttd` field, an admin/model/Game. Only
// DateEvent reads it (its toString() asks for the game date), so the seam is that
// one method; a listener that wants the full model gets the Game back from
// GetOpenttd() and works through admin/model's own API.
type Openttd interface {
	GetDate() util.Calendar
}

// Action is ClientEvent.Action and CompanyEvent.Action. Java declares the same
// three-item enum inside both classes; one Go type serves both, and the names it
// prints - which is what toString() shows - are the same either way.
type Action int

const (
	ACTION_CREATE Action = 0
	ACTION_UPDATE Action = 1
	ACTION_DELETE Action = 2
	ACTION_NONE   Action = -1 // Port sentinel: "this row of the mapping table has no action"
)

// ActionName is the enum constant's name.
func ActionName(action Action) string {
	switch action {
	case ACTION_CREATE:
		return "CREATE"
	case ACTION_UPDATE:
		return "UPDATE"
	case ACTION_DELETE:
		return "DELETE"
	default:
		return "ACTION_UNKNOWN"
	}
}
