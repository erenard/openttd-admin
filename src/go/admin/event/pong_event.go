// PongEvent mirrors com.openttd.admin.event.PongEvent: the server answered the
// admin's ping. PongEventListener is com.openttd.admin.event.PongEventListener,
// kept in this file because it is one method.

package event

import "strconv"

// PongEventListener is PongEventListener.
type PongEventListener interface {
	OnPongEvent(event *PongEvent)
}

type PongEvent struct {
	openttd Openttd
	pingId  int64
}

// NewPongEvent is PongEvent(Game openttd, long pingId).
func NewPongEvent(openttd Openttd, pingId int64) *PongEvent {
	return &PongEvent{openttd: openttd, pingId: pingId}
}

// GetOpenttd is getOpenttd, typed as the Openttd seam rather than Game.
func (e *PongEvent) GetOpenttd() Openttd { return e.openttd }

// GetPingId is getPingId.
func (e *PongEvent) GetPingId() int64 { return e.pingId }

// ToString is toString.
func (e *PongEvent) ToString() string {
	return "Pong (" + strconv.FormatInt(e.pingId, 10) + ")"
}

// Kind is the PongEvent.class key this event is queued under.
func (e *PongEvent) Kind() EventKind { return EVENT_PONG }

// Notify is notify(PongEventListener).
func (e *PongEvent) Notify(listener EventListener) {
	var target, ok = listener.(PongEventListener)
	if ok {
		target.OnPongEvent(e)
		return
	}
	eventLog.Error(EventKindName(EVENT_PONG) + " cannot notify a listener that is not a PongEventListener: " + e.ToString())
}
