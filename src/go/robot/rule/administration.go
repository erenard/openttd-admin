// Administration is com.openttd.robot.rule.Administration: the chat commands an
// admin types to inspect and discipline the players on the server.
//
// Rules #1..#7 of the class comment are "$clients"/"$cls", "$kick"/"$k",
// "$ban"/"$b", "$warn"/"$w", "$help"/"$h", "$pause" and "$unpause". Every one
// of them is gated on the sender being an admin in the application's user table,
// which is why the rule needs the ExternalUsers rule. See rule/AbstractRule.java
// for the base.
//
// There is no Java test for this class, so administration_test.go is new
// coverage written against this source as the specification.

package rule

import "strings"
import "com.openttd/admin"
import "com.openttd/admin/event"
import "com.openttd/admin/model"
import "com.openttd/util"

var administrationLog = util.GetLogger("Administration")

type Administration struct {
	base          *AbstractRule
	externalUsers *ExternalUsers
	// warningCountByClientId is warningCountByClientId: how often a player has
	// been warned, which decides whether the next warning kicks or bans them.
	warningCountByClientId map[int]int
}

// NewAdministration is Administration(OpenttdAdmin openttdAdmin,
// ExternalUsers externalUsers).
func NewAdministration(openttd *admin.OpenttdAdmin, externalUsers *ExternalUsers) *Administration {
	var administration = &Administration{
		externalUsers:          externalUsers,
		warningCountByClientId: map[int]int{}}
	administration.base = NewAbstractRule(openttd, administration, []event.EventKind{event.EVENT_CHAT})
	return administration
}

// OnChatEvent is onChatEvent(ChatEvent).
func (a *Administration) OnChatEvent(chatEvent *event.ChatEvent) {
	if administrationLog.IsDebugEnabled() {
		administrationLog.Debug("Administration.onChatEvent " + chatEvent.ToString())
	}
	var clientId = chatEvent.GetClientId()
	var message = chatEvent.GetMessage()
	if message == "" {
		// Java's `if (message != null)`.
		return
	}
	var command = strings.ToLower(strings.TrimSpace(message))
	var user = a.externalUsers.GetExternalUser(clientId)
	if user == nil || !user.IsAdmin() {
		// Every command below is admin only.
		return
	}
	if command == "$clients" || command == "$cls" {
		// Rule #1.
		a.showClients(a.gameOf(chatEvent.GetOpenttd()), clientId)
	} else if strings.HasPrefix(command, "$kick") || strings.HasPrefix(command, "$k") {
		// Rule #2.
		var kickedClientId, ok = ParseInt(a.argumentAfter(command))
		if !ok {
			a.showMessage(clientId, "Usage: $kick clientId (try $clients to find clientIds)")
			return
		}
		a.kick(kickedClientId)
	} else if strings.HasPrefix(command, "$ban") || strings.HasPrefix(command, "$b") {
		// Rule #3.
		var bannedClientId, ok = ParseInt(a.argumentAfter(command))
		if !ok {
			a.showMessage(clientId, "Usage: $ban clientId (try $clients to find clientIds)")
			return
		}
		a.ban(bannedClientId)
	} else if strings.HasPrefix(command, "$warn") || strings.HasPrefix(command, "$w") {
		// Rule #4. Java re-reads the un-lower-cased message here, so the warning
		// text keeps whatever case the player typed.
		var warned = strings.TrimSpace(chatEvent.GetMessage())
		var args = strings.Split(warned, " ")
		if len(args) < 2 {
			// DEVIATION FROM JAVA: args[1] would raise
			// ArrayIndexOutOfBoundsException, which its catch does not handle;
			// the port reports the usage line.
			a.showMessage(clientId, "Usage: $warn clientId [message] (try $clients to find clientIds)")
			return
		}
		var warnedClientId, ok = ParseInt(strings.TrimSpace(args[1]))
		if !ok {
			a.showMessage(clientId, "Usage: $warn clientId [message] (try $clients to find clientIds)")
			return
		}
		var warningMessage = ""
		if len(args) > 2 {
			// Java takes the substring from the first occurrence of the third
			// token, which keeps the rest of the line verbatim.
			var third = args[2]
			var start = strings.Index(warned, third)
			if start != -1 {
				warningMessage = warned[start:]
			}
		}
		a.warn(warnedClientId, warningMessage)
	} else if strings.HasPrefix(command, "$help") || command == "$h" {
		// Rule #5.
		a.showHelpAdmin(clientId)
	} else if strings.HasPrefix(command, "$pause") || command == "$p" {
		// Rule #6.
		a.pause()
	} else if strings.HasPrefix(command, "$unpause") || command == "$u" {
		// Rule #7.
		a.unpause()
	}
}

// gameOf is chatEvent.getOpenttd(), down-cast from the event's Openttd seam to
// the admin/model/Game the rules read.
func (a *Administration) gameOf(source event.Openttd) *model.Game {
	var game, ok = source.(*model.Game)
	if ok {
		return game
	}
	return nil
}

// argumentAfter is message.split(" ")[1].trim(): the first argument after the
// command, or the empty string when the message has none.
func (a *Administration) argumentAfter(command string) string {
	var parts = strings.Split(command, " ")
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func (a *Administration) showMessage(clientId int, message string) {
	a.base.GetSend().ChatClient(int64(clientId), message)
}

// showClients is showClients(Game openttd, int clientId): the client table.
func (a *Administration) showClients(game *model.Game, clientId int) {
	var send = a.base.GetSend()
	send.ChatClient(int64(clientId), "#ClientId, Name, IpAddress, CompanyId")
	for _, client := range game.GetClients() {
		send.ChatClient(int64(clientId), "#"+util.FormatInt(int64(client.GetId()))+", "+client.GetName()+", "+client.GetIp()+", "+util.FormatInt(int64(client.GetCompany())))
	}
}

// kick is kick(int clientId): the rcon kick.
func (a *Administration) kick(clientId int) {
	a.base.GetSend().Rcon("kick " + util.FormatInt(int64(clientId)))
}

// ban is ban(int clientId): the rcon ban.
func (a *Administration) ban(clientId int) {
	a.base.GetSend().Rcon("ban " + util.FormatInt(int64(clientId)))
}

// warn is warn(int clientId, String warning): the escalation ladder. The count
// read decides what this warning does, and is always incremented afterwards.
func (a *Administration) warn(clientId int, warning string) {
	var send = a.base.GetSend()
	var warningCount, found = a.warningCountByClientId[clientId]
	if !found {
		warningCount = 0
	}
	// DEVIATION FROM JAVA: the port's empty string stands in for Java's null,
	// so an absent warning adds no parenthesised suffix.
	var suffix = ""
	if warning != "" {
		suffix = " (" + warning + ")"
	}
	if warningCount == 0 {
		send.ChatClient(int64(clientId), "Be warned! No rule breaking!"+suffix)
	} else if warningCount == 1 {
		send.ChatClient(int64(clientId), "Last warning! No rule breaking!"+suffix)
	} else if warningCount == 2 {
		a.kick(clientId)
	} else if warningCount > 2 {
		a.ban(clientId)
	}
	a.warningCountByClientId[clientId] = warningCount + 1
}

// showHelpAdmin is showHelpAdmin(int clientId): the two command lists.
func (a *Administration) showHelpAdmin(clientId int) {
	var send = a.base.GetSend()
	send.ChatClient(int64(clientId), "Valid commands are $clients, $companies, $warn, $kick, $ban, $reset, $pause, $unpause")
	send.ChatClient(int64(clientId), "Short commands are $cls, $cps, $w, $k, $b, $r, $p, $u")
}

// pause is pause(): the rcon pause.
func (a *Administration) pause() {
	a.base.GetSend().Rcon("pause")
}

// unpause is unpause(): the rcon unpause.
func (a *Administration) unpause() {
	a.base.GetSend().Rcon("unpause")
}
