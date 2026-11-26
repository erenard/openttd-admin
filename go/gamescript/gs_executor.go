// GSExecutor mirrors com.openttd.gamescript.GSExecutor: it numbers the gamescript
// requests and matches AdminCmd's replies to them by that id.
//
// Java pulls the current NetworkAdminSender out of OpenttdAdmin on every send,
// so a reconnect's new sender is picked up; Sender and SenderProvider below stand
// in for those two types, which this port has not reached yet.
//
// Java's onGameScriptEvent(GameScriptEvent) is OnGameScriptReply here, taking
// the getMessage() of that event: admin/event is a later stage. Once it landed,
// OnGameScriptEvent below is the listener method the facade registers, and it
// simply hands the message to OnGameScriptReply.

package gamescript

import "strconv"
import "com.openttd/admin/event"
import "com.openttd/util"

var log = util.GetLogger("GSExecutor")

// Sender is the part of NetworkAdminSender GSExecutor sends through:
// gameScript(String json), whose false return means the payload did not fit
// the packet.
type Sender interface {
	GameScript(json string) bool
}

// SenderProvider is the part of OpenttdAdmin it reads that sender from: getSend(),
// which is null until the admin is connected.
type SenderProvider interface {
	GetSender() Sender
}

// nextRequestId is Java's static field: shared by every executor, so the ids
// keep counting up across connections until someone calls Clear().
var nextRequestId = 0

type GSExecutor struct {
	admin SenderProvider
	requestById map[string]GSRequest
}

// NewGSExecutor is GSExecutor(OpenttdAdmin admin).
func NewGSExecutor(admin SenderProvider) *GSExecutor {
	return &GSExecutor{admin: admin, requestById: map[string]GSRequest{}}
}

// Clear is clear(): start numbering again and forget the pending requests.
func (e *GSExecutor) Clear() {
	nextRequestId = 0
	e.requestById = map[string]GSRequest{}
}

// Send is send(GSRequest request): number the request, remember it, put its
// JSON on the wire. Without a connection nothing is sent, and the request is
// left pending, exactly as Java leaves it.
func (e *GSExecutor) Send(request GSRequest) {
	if request == nil {
		return
	}
	if e.admin == nil {
		return
	}
	var sender = e.admin.GetSender()
	if sender == nil {
		return
	}
	var id = strconv.Itoa(nextRequestId)
	nextRequestId = nextRequestId + 1
	request.SetId(id)
	e.requestById[id] = request
	sender.GameScript(request.ToString())
}

// OnGameScriptReply is onGameScriptEvent(GameScriptEvent), given the event's
// getMessage(). AdminCmd answers {id, data} or {id, exception}, and {exception}
// alone for a malformed request, which is the case that used to throw a
// NullPointerException on the listener thread.
func (e *GSExecutor) OnGameScriptReply(payload string) {
	var response = Parse(payload)
	if response == nil || !response.HasId {
		log.Warn("Ignoring gamescript reply without a request id: " + payload)
		return
	}
	var request, has = e.requestById[response.Id]
	if !has {
		log.Warn("Ignoring gamescript reply without a pending request: " + payload)
		return
	}
	delete(e.requestById, response.Id)
	request.SetResponse(response)
}

// OnGameScriptEvent is onGameScriptEvent(GameScriptEvent): the method that makes
// the executor a GameScriptEventListener, so the facade can register it with the
// dispatcher and the server's replies arrive here. It is a one line delegation -
// the event carries nothing but its message.
func (e *GSExecutor) OnGameScriptEvent(gameScriptEvent *event.GameScriptEvent) {
	e.OnGameScriptReply(gameScriptEvent.GetMessage())
}
