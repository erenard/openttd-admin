// ConsoleEvent mirrors com.openttd.admin.event.ConsoleEvent: a line the server
// printed to its console. ConsoleEventListener is com.openttd.admin.event
// .ConsoleEventListener, kept in this file because it is one method.

package event

// ConsoleEventListener is ConsoleEventListener.
type ConsoleEventListener interface {
	OnConsoleEvent(event *ConsoleEvent)
}

type ConsoleEvent struct {
	openttd Openttd
	origin  string
	message string
}

// NewConsoleEvent is ConsoleEvent(Game openttd, String origin, String message).
func NewConsoleEvent(openttd Openttd, origin string, message string) *ConsoleEvent {
	return &ConsoleEvent{openttd: openttd, origin: origin, message: message}
}

// GetOpenttd is getOpenttd, typed as the Openttd seam rather than Game.
func (e *ConsoleEvent) GetOpenttd() Openttd { return e.openttd }

// GetOrigin is getOrigin.
func (e *ConsoleEvent) GetOrigin() string { return e.origin }

// GetMessage is getMessage.
func (e *ConsoleEvent) GetMessage() string { return e.message }

// ToString is toString.
func (e *ConsoleEvent) ToString() string {
	return "Console (" + e.origin + ") " + e.message
}

// Kind is the ConsoleEvent.class key this event is queued under.
func (e *ConsoleEvent) Kind() EventKind { return EVENT_CONSOLE }

// Notify is notify(ConsoleEventListener).
func (e *ConsoleEvent) Notify(listener EventListener) {
	var target, ok = listener.(ConsoleEventListener)
	if ok {
		target.OnConsoleEvent(e)
		return
	}
	eventLog.Error(EventKindName(EVENT_CONSOLE) + " cannot notify a listener that is not a ConsoleEventListener: " + e.ToString())
}
