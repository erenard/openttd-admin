// EventRecorder mirrors com.openttd.server.EventRecorder: it implements every
// listener interface, so a test can assert on all the functional events the
// facade dispatches instead of only the three that were covered before.
//
// Java's registerWith(OpenttdAdmin) passes `this` to addListener; this dialect
// has no self reference, so it is the free RegisterWith below, with the recorder
// as an explicit argument.

package server

import "com.openttd/admin"
import "com.openttd/admin/event"

type EventRecorder struct {
	chats []*event.ChatEvent
	clients []*event.ClientEvent
	companies []*event.CompanyEvent
	consoles []*event.ConsoleEvent
	dates []*event.DateEvent
	gameScripts []*event.GameScriptEvent
	pongs []*event.PongEvent
	rCons []*event.RConEvent
	rConEnds []*event.RConEndEvent
}

// NewEventRecorder is new EventRecorder().
func NewEventRecorder() *EventRecorder {
	return &EventRecorder{}
}

// Clear is clear(): drop everything collected so far, so a test asserts on its
// own packets only.
func (r *EventRecorder) Clear() {
	r.chats = []*event.ChatEvent{}
	r.clients = []*event.ClientEvent{}
	r.companies = []*event.CompanyEvent{}
	r.consoles = []*event.ConsoleEvent{}
	r.dates = []*event.DateEvent{}
	r.gameScripts = []*event.GameScriptEvent{}
	r.pongs = []*event.PongEvent{}
	r.rCons = []*event.RConEvent{}
	r.rConEnds = []*event.RConEndEvent{}
}

// RegisterWith is registerWith(OpenttdAdmin): the recorder listens for every
// event kind, so the dispatcher's map has an entry per kind and a listener that
// is not of the right kind cannot be reached by accident.
func RegisterWith(facade *admin.OpenttdAdmin, r *EventRecorder) {
	facade.AddListener(event.EVENT_CHAT, r)
	facade.AddListener(event.EVENT_CLIENT, r)
	facade.AddListener(event.EVENT_COMPANY, r)
	facade.AddListener(event.EVENT_CONSOLE, r)
	facade.AddListener(event.EVENT_DATE, r)
	facade.AddListener(event.EVENT_GAME_SCRIPT, r)
	facade.AddListener(event.EVENT_PONG, r)
	facade.AddListener(event.EVENT_RCON, r)
	facade.AddListener(event.EVENT_RCON_END, r)
}

// GetChatEvents is getChatEvents.
func (r *EventRecorder) GetChatEvents() []*event.ChatEvent { return r.chats }

// GetClientEvents is getClientEvents.
func (r *EventRecorder) GetClientEvents() []*event.ClientEvent { return r.clients }

// GetCompanyEvents is getCompanyEvents.
func (r *EventRecorder) GetCompanyEvents() []*event.CompanyEvent { return r.companies }

// GetConsoleEvents is getConsoleEvents.
func (r *EventRecorder) GetConsoleEvents() []*event.ConsoleEvent { return r.consoles }

// GetDateEvents is getDateEvents.
func (r *EventRecorder) GetDateEvents() []*event.DateEvent { return r.dates }

// GetGameScriptEvents is getGameScriptEvents.
func (r *EventRecorder) GetGameScriptEvents() []*event.GameScriptEvent { return r.gameScripts }

// GetPongEvents is getPongEvents.
func (r *EventRecorder) GetPongEvents() []*event.PongEvent { return r.pongs }

// GetRConEvents is getRConEvents.
func (r *EventRecorder) GetRConEvents() []*event.RConEvent { return r.rCons }

// GetRConEndEvents is getRConEndEvents.
func (r *EventRecorder) GetRConEndEvents() []*event.RConEndEvent { return r.rConEnds }

// OnChatEvent is onChatEvent.
func (r *EventRecorder) OnChatEvent(chatEvent *event.ChatEvent) {
	r.chats = append(r.chats, chatEvent)
}

// OnClientEvent is onClientEvent.
func (r *EventRecorder) OnClientEvent(clientEvent *event.ClientEvent) {
	r.clients = append(r.clients, clientEvent)
}

// OnCompanyEvent is onCompanyEvent.
func (r *EventRecorder) OnCompanyEvent(companyEvent *event.CompanyEvent) {
	r.companies = append(r.companies, companyEvent)
}

// OnConsoleEvent is onConsoleEvent.
func (r *EventRecorder) OnConsoleEvent(consoleEvent *event.ConsoleEvent) {
	r.consoles = append(r.consoles, consoleEvent)
}

// OnDateEvent is onDateEvent.
func (r *EventRecorder) OnDateEvent(dateEvent *event.DateEvent) {
	r.dates = append(r.dates, dateEvent)
}

// OnGameScriptEvent is onGameScriptEvent.
func (r *EventRecorder) OnGameScriptEvent(gameScriptEvent *event.GameScriptEvent) {
	r.gameScripts = append(r.gameScripts, gameScriptEvent)
}

// OnPongEvent is onPongEvent.
func (r *EventRecorder) OnPongEvent(pongEvent *event.PongEvent) {
	r.pongs = append(r.pongs, pongEvent)
}

// OnRConEvent is onRConEvent.
func (r *EventRecorder) OnRConEvent(rConEvent *event.RConEvent) {
	r.rCons = append(r.rCons, rConEvent)
}

// OnRConEndEvent is onRConEndEvent.
func (r *EventRecorder) OnRConEndEvent(rConEndEvent *event.RConEndEvent) {
	r.rConEnds = append(r.rConEnds, rConEndEvent)
}
