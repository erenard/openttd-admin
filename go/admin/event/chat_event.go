// ChatEvent mirrors com.openttd.admin.event.ChatEvent: a chat message the
// server relayed to the admin. ChatEventListener is com.openttd.admin.event
// .ChatEventListener, kept in this file because it is one method.

package event

import "strconv"

// ChatEventListener is ChatEventListener: what a chat listener is asked.
type ChatEventListener interface {
	OnChatEvent(event *ChatEvent)
}

type ChatEvent struct {
	openttd  Openttd
	clientId int
	message  string
}

// NewChatEvent is ChatEvent(Game openttd, long clientId, String message): Java
// narrows the client id to an int, and the port does the same.
func NewChatEvent(openttd Openttd, clientId int64, message string) *ChatEvent {
	return &ChatEvent{openttd: openttd, clientId: int(clientId), message: message}
}

// GetOpenttd is getOpenttd, typed as the Openttd seam rather than Game.
func (e *ChatEvent) GetOpenttd() Openttd { return e.openttd }

// GetClientId is getClientId.
func (e *ChatEvent) GetClientId() int { return e.clientId }

// GetMessage is getMessage.
func (e *ChatEvent) GetMessage() string { return e.message }

// ToString is toString.
func (e *ChatEvent) ToString() string {
	return strconv.Itoa(e.clientId) + " says " + e.message
}

// Kind is the ChatEvent.class key this event is queued under.
func (e *ChatEvent) Kind() EventKind { return EVENT_CHAT }

// Notify is notify(ChatEventListener): the down-cast Java's generic Event<L>
// made unnecessary. A listener registered for chat events that does not
// implement ChatEventListener is the ClassCastException Java logged from
// EventDispatcher.run(); the port logs it where it happens.
func (e *ChatEvent) Notify(listener EventListener) {
	var target, ok = listener.(ChatEventListener)
	if ok {
		target.OnChatEvent(e)
		return
	}
	eventLog.Error(EventKindName(EVENT_CHAT) + " cannot notify a listener that is not a ChatEventListener: " + e.ToString())
}
