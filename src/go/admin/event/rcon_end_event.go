// RConEndEvent mirrors com.openttd.admin.event.RConEndEvent: a remote console
// command finished. RConEndEventListener is com.openttd.admin.event
// .RConEndEventListener, kept in this file because it is one method.

package event

// RConEndEventListener is RConEndEventListener.
type RConEndEventListener interface {
	OnRConEndEvent(event *RConEndEvent)
}

type RConEndEvent struct {
	openttd Openttd
	command string
}

// NewRConEndEvent is RConEndEvent(Game openttd, String command).
func NewRConEndEvent(openttd Openttd, command string) *RConEndEvent {
	return &RConEndEvent{openttd: openttd, command: command}
}

// GetOpenttd is getOpenttd, typed as the Openttd seam rather than Game.
func (e *RConEndEvent) GetOpenttd() Openttd { return e.openttd }

// GetCommand is getCommand: the command that finished.
func (e *RConEndEvent) GetCommand() string { return e.command }

// ToString is toString, including the odd spacing Java has.
func (e *RConEndEvent) ToString() string {
	return "RCon ended :" + e.command
}

// Kind is the RConEndEvent.class key this event is queued under.
func (e *RConEndEvent) Kind() EventKind { return EVENT_RCON_END }

// Notify is notify(RConEndEventListener).
func (e *RConEndEvent) Notify(listener EventListener) {
	var target, ok = listener.(RConEndEventListener)
	if ok {
		target.OnRConEndEvent(e)
		return
	}
	eventLog.Error(EventKindName(EVENT_RCON_END) + " cannot notify a listener that is not a RConEndEventListener: " + e.ToString())
}
