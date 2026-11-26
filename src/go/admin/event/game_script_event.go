// GameScriptEvent mirrors com.openttd.admin.event.GameScriptEvent: a GameScript
// reply the server sent back to the admin. GSExecutor registers itself as a
// listener for this event, so its message is the JSON AdminCmd answers with.
// GameScriptEventListener is com.openttd.admin.event.GameScriptEventListener,
// kept in this file because it is one method.

package event

// GameScriptEventListener is GameScriptEventListener.
type GameScriptEventListener interface {
	OnGameScriptEvent(event *GameScriptEvent)
}

type GameScriptEvent struct {
	openttd Openttd
	message string
}

// NewGameScriptEvent is GameScriptEvent(Game openttd, String message).
func NewGameScriptEvent(openttd Openttd, message string) *GameScriptEvent {
	return &GameScriptEvent{openttd: openttd, message: message}
}

// GetOpenttd is getOpenttd, typed as the Openttd seam rather than Game.
func (e *GameScriptEvent) GetOpenttd() Openttd { return e.openttd }

// GetMessage is getMessage: the JSON of the server's reply.
func (e *GameScriptEvent) GetMessage() string { return e.message }

// ToString is toString.
func (e *GameScriptEvent) ToString() string {
	return "GameScript " + e.message
}

// Kind is the GameScriptEvent.class key this event is queued under.
func (e *GameScriptEvent) Kind() EventKind { return EVENT_GAME_SCRIPT }

// Notify is notify(GameScriptEventListener).
func (e *GameScriptEvent) Notify(listener EventListener) {
	var target, ok = listener.(GameScriptEventListener)
	if ok {
		target.OnGameScriptEvent(e)
		return
	}
	eventLog.Error(EventKindName(EVENT_GAME_SCRIPT) + " cannot notify a listener that is not a GameScriptEventListener: " + e.ToString())
}
