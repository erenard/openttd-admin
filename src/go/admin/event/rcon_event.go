// RConEvent mirrors com.openttd.admin.event.RConEvent: a line of output from a
// remote console command. RConEventListener is com.openttd.admin.event
// .RConEventListener, kept in this file because it is one method.

package event

import "strconv"

// RConEventListener is RConEventListener.
type RConEventListener interface {
	OnRConEvent(event *RConEvent)
}

type RConEvent struct {
	openttd Openttd
	// color is Java's char: the colour byte of the rcon output. It is an int
	// here, as it is in NetworkAdminEvent, so ToString prints the number rather
	// than the character that byte happens to be.
	color   int
	message string
}

// NewRConEvent is RConEvent(Game openttd, char color, String message).
func NewRConEvent(openttd Openttd, color int, message string) *RConEvent {
	return &RConEvent{openttd: openttd, color: color, message: message}
}

// GetOpenttd is getOpenttd, typed as the Openttd seam rather than Game.
func (e *RConEvent) GetOpenttd() Openttd { return e.openttd }

// GetColor is getColor.
func (e *RConEvent) GetColor() int { return e.color }

// GetMessage is getMessage.
func (e *RConEvent) GetMessage() string { return e.message }

// ToString is toString, with the colour as a number (see the color field).
func (e *RConEvent) ToString() string {
	return "RCon (" + strconv.Itoa(e.color) + ") " + e.message
}

// Kind is the RConEvent.class key this event is queued under.
func (e *RConEvent) Kind() EventKind { return EVENT_RCON }

// Notify is notify(RConEventListener).
func (e *RConEvent) Notify(listener EventListener) {
	var target, ok = listener.(RConEventListener)
	if ok {
		target.OnRConEvent(e)
		return
	}
	eventLog.Error(EventKindName(EVENT_RCON) + " cannot notify a listener that is not a RConEventListener: " + e.ToString())
}
