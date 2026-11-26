// ServerAnnouncer is com.openttd.robot.rule.ServerAnnouncer: the text a spectator
// is shown when they join, and the answers to the two commands every player may
// use.
//
// Rules #1 and #2 of the class comment are "!rules"/"!r" and "!help"/"!h"; the
// third thing it does is the welcome message on ClientEvent CREATE. Unlike the
// other rules nothing here is admin gated, so it needs no ExternalUsers.
//
// There is no Java test for this class, so server_announcer_test.go is new
// coverage written against this source as the specification.

package rule

import "strings"
import "com.openttd/admin"
import "com.openttd/admin/event"
import "com.openttd/util"

var serverAnnouncerLog = util.GetLogger("ServerAnnouncer")

type ServerAnnouncer struct {
	base *AbstractRule
}

// NewServerAnnouncer is ServerAnnouncer(OpenttdAdmin openttdAdmin). It listens
// for ChatEvent and ClientEvent, the kinds listEventTypes() returns.
func NewServerAnnouncer(openttd *admin.OpenttdAdmin) *ServerAnnouncer {
	var announcer = &ServerAnnouncer{}
	announcer.base = NewAbstractRule(openttd, announcer, []event.EventKind{event.EVENT_CHAT, event.EVENT_CLIENT})
	return announcer
}

// OnClientEvent is onClientEvent(ClientEvent): only CREATE does anything, and
// Java's default case is empty, so nothing else is.
func (a *ServerAnnouncer) OnClientEvent(clientEvent *event.ClientEvent) {
	if serverAnnouncerLog.IsDebugEnabled() {
		serverAnnouncerLog.Debug("ServerAnnouncer.onClientEvent " + clientEvent.ToString())
	}
	if clientEvent.GetAction() == event.ACTION_CREATE {
		a.showWelcome(clientEvent.GetClientId())
	}
}

// OnChatEvent is onChatEvent(ChatEvent).
func (a *ServerAnnouncer) OnChatEvent(chatEvent *event.ChatEvent) {
	if serverAnnouncerLog.IsDebugEnabled() {
		serverAnnouncerLog.Debug("ServerAnnouncer.onChatEvent " + chatEvent.ToString())
	}
	var clientId = chatEvent.GetClientId()
	var message = chatEvent.GetMessage()
	if message == "" {
		// Java's `if (message != null)`.
		return
	}
	var command = strings.ToLower(strings.TrimSpace(message))
	if command == "!rules" || command == "!r" {
		// Rule #1.
		a.showRules(clientId)
	} else if command == "!help" || command == "!h" {
		// Rule #2.
		a.showHelp(clientId)
	}
}

// showWelcome is showWelcome(int clientId).
func (a *ServerAnnouncer) showWelcome(clientId int) {
	var send = a.base.GetSend()
	send.ChatClient(int64(clientId), "Welcome to www.strategyboard.net goal server.")
	send.ChatClient(int64(clientId), "Type: !howto or !help.")
}

// showRules is showRules(int clientId): the eleven numbered rules of the goal
// server, then the link to the site.
func (a *ServerAnnouncer) showRules(clientId int) {
	var send = a.base.GetSend()
	send.ChatClient(int64(clientId), "Any violation of these rules will make you banned.")
	send.ChatClient(int64(clientId), "# 0: Retaliation doesn't allow you to break any of these rules.")
	send.ChatClient(int64(clientId), "# 1: Stay friendly and respect each other! Any racism/rude language will be sanctioned.")
	send.ChatClient(int64(clientId), "# 2: Play fair! No reserving! No rating exploit! No blockage! Don't build useless structures!")
	send.ChatClient(int64(clientId), "# 3: No heavy terraforming/Town Destroying!")
	send.ChatClient(int64(clientId), "# 4: No 'Ghost companies' for extra money (e.g. to prepare your future build).")
	send.ChatClient(int64(clientId), "# 5: No multiple companies for one person! (One person => One company).")
	send.ChatClient(int64(clientId), "# 6: No one way trains!")
	send.ChatClient(int64(clientId), "# 7: Don't buy other companies!")
	send.ChatClient(int64(clientId), "# 8: Nobody can claim an industry(or goods)!")
	send.ChatClient(int64(clientId), "# 9: No 'Next,Next,Turn'!")
	send.ChatClient(int64(clientId), "#10: No catchment area exploit!")
	send.ChatClient(int64(clientId), "http://www.strategyboard.net/ for more info")
}

// showHelp is showHelp(int clientId).
func (a *ServerAnnouncer) showHelp(clientId int) {
	var send = a.base.GetSend()
	send.ChatClient(int64(clientId), "Valid commands are !goal, !rules, !login, !howto, !score, !rename and !resetme.")
	send.ChatClient(int64(clientId), "Short commands are !g, !r, !cv or !cp (equ. to !score)")
}
