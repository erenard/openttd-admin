// ClientEvent mirrors com.openttd.admin.event.ClientEvent: a client joined,
// changed, or left. ClientEventListener is com.openttd.admin.event
// .ClientEventListener, kept in this file because it is one method.

package event

import "strconv"

// ClientEventListener is ClientEventListener.
type ClientEventListener interface {
	OnClientEvent(event *ClientEvent)
}

type ClientEvent struct {
	openttd  Openttd
	clientId int
	action   Action
}

// NewClientEvent is ClientEvent(Game openttd, long clientId, Action action):
// Java narrows the client id to an int, and the port does the same.
func NewClientEvent(openttd Openttd, clientId int64, action Action) *ClientEvent {
	return &ClientEvent{openttd: openttd, clientId: int(clientId), action: action}
}

// GetOpenttd is getOpenttd, typed as the Openttd seam rather than Game.
func (e *ClientEvent) GetOpenttd() Openttd { return e.openttd }

// GetClientId is getClientId.
func (e *ClientEvent) GetClientId() int { return e.clientId }

// GetAction is getAction.
func (e *ClientEvent) GetAction() Action { return e.action }

// ToString is toString: the action's name, then the client id.
func (e *ClientEvent) ToString() string {
	return ActionName(e.action) + ":" + strconv.Itoa(e.clientId)
}

// Kind is the ClientEvent.class key this event is queued under.
func (e *ClientEvent) Kind() EventKind { return EVENT_CLIENT }

// Notify is notify(ClientEventListener). See ChatEvent.Notify for what the
// down-cast stands for; the same applies here.
func (e *ClientEvent) Notify(listener EventListener) {
	var target, ok = listener.(ClientEventListener)
	if ok {
		target.OnClientEvent(e)
		return
	}
	eventLog.Error(EventKindName(EVENT_CLIENT) + " cannot notify a listener that is not a ClientEventListener: " + e.ToString())
}
