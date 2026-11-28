// ExternalUsers is com.openttd.robot.rule.ExternalUsers: the rule that turns
// the chat commands a player types into an identity in the application's user
// table, and renames the client after that user.
//
// Rules #1, #2 and #3 of the class comment are "!login <token>", "!howto" and
// "!rename <name>". See rule/AbstractRule.java for the base this rule extends.

package rule

import "strings"
import "com.openttd/admin"
import "com.openttd/admin/event"
import "com.openttd/admin/model"
import "com.openttd/util"
import robotmodel "com.openttd/robot/model"
import "com.openttd/robot"

var externalUsersLog = util.GetLogger("ExternalUsers")

// COMPANY_NONE is the 255 Java compares a client's company id against: the
// spectator slot, a player who owns no company.
const COMPANY_NONE = 255

type ExternalUsers struct {
	base *AbstractRule
	// externalUserService is the field initializer Java runs before the
	// constructor body, so the application has to inject the service first.
	externalUserService    robot.ExternalUserService
	externalUserByClientId map[int]*robotmodel.ExternalUser
	// externalUserByCompanyId is externalUserByCompanyId: the owner of a company
	// as the application knows it, keyed by company id.
	externalUserByCompanyId map[int]*robotmodel.ExternalUser
}

// NewExternalUsers is ExternalUsers(OpenttdAdmin openttdAdmin).
func NewExternalUsers(openttd *admin.OpenttdAdmin) *ExternalUsers {
	var users = &ExternalUsers{
		externalUserService:     robot.GetInstance().GetExternalUserService(),
		externalUserByClientId:  map[int]*robotmodel.ExternalUser{},
		externalUserByCompanyId: map[int]*robotmodel.ExternalUser{}}
	users.base = NewAbstractRule(openttd, users, []event.EventKind{event.EVENT_CHAT})
	return users
}

// OnChatEvent is onChatEvent(ChatEvent). The kinds this rule listens for are
// ChatEvent only, passed to NewAbstractRule by the constructor.
func (u *ExternalUsers) OnChatEvent(chatEvent *event.ChatEvent) {
	if externalUsersLog.IsDebugEnabled() {
		externalUsersLog.Debug("ExternalUsers.onChatEvent " + chatEvent.ToString())
	}
	var clientId = chatEvent.GetClientId()
	var message = chatEvent.GetMessage()
	if message == "" {
		// Java's `if (message != null)`: the port's empty string stands in for
		// a missing message.
		return
	}
	message = strings.TrimSpace(message)
	var command = strings.ToLower(message)
	if strings.HasPrefix(command, "!login ") {
		// Rule #1.
		var loginToken = message[7:]
		var user = u.identifyUser(loginToken)
		if user != nil {
			var logged = true
			var previous, found = u.externalUserByClientId[clientId]
			if found {
				if previous.Equals(user) {
					u.showClientAlreadyLogged(clientId)
				} else {
					logged = false
				}
			} else {
				u.showLoginSucceed(clientId)
				u.externalUserByClientId[clientId] = user
				u.renameClient(clientId, user.GetName())
			}
			var game = u.gameOf(chatEvent)
			if game == nil {
				return
			}
			var client = game.GetClient(clientId)
			if client == nil {
				return
			}
			var companyId = client.GetCompany()
			if logged && companyId != COMPANY_NONE {
				var companyOwner = u.GetOwnerOf(companyId)
				var company = game.GetCompany(companyId)
				if company == nil {
					// Java would have thrown: the company is not in the model
					// yet when the login arrives.
					return
				}
				var companyName = company.GetName()
				if companyOwner != nil {
					if companyOwner.Equals(user) {
						u.showCompanyAlreadyOwned(clientId, companyName)
					} else {
						u.showCompanyAlreadyOwnedBy(clientId, companyName, companyOwner)
					}
				} else {
					// A user owns one company only: drop the old ownership first.
					// Java iterates a copy of the entry set while removing from it;
					// the port collects the keys first for the same reason.
					var stale = []int{}
					for companyKey, owner := range u.externalUserByCompanyId {
						externalUsersLog.Info(util.FormatInt(int64(companyKey)) + " " + owner.ToString())
						if owner.Equals(user) {
							stale = append(stale, companyKey)
						}
					}
					for _, companyKey := range stale {
						u.RemoveOwnerOf(companyKey)
					}
					u.SetOwnerOf(companyId, user)
					u.showCompanyOwned(clientId, companyName)
				}
			}
		} else {
			u.showLoginFailed(clientId)
			u.showHowtoLogin(clientId)
		}
	} else if command == "!howto" {
		// Rule #2.
		u.showHowtoLogin(clientId)
	} else if strings.HasPrefix(command, "!rename ") {
		// Rule #3.
		var newName = strings.TrimSpace(message[8:])
		u.renameClient(clientId, newName)
	}
}

// GetExternalUser is getExternalUser(long clientId): the user a client logged in
// as, or nil.
func (u *ExternalUsers) GetExternalUser(clientId int) *robotmodel.ExternalUser {
	var user, found = u.externalUserByClientId[clientId]
	if found {
		return user
	}
	return nil
}

// GetOwnerOf is getOwnerOf(short companyId).
func (u *ExternalUsers) GetOwnerOf(companyId int) *robotmodel.ExternalUser {
	var owner, found = u.externalUserByCompanyId[companyId]
	if found {
		return owner
	}
	return nil
}

// SetOwnerOf is setOwnerOf(short companyId, ExternalUser owner).
func (u *ExternalUsers) SetOwnerOf(companyId int, owner *robotmodel.ExternalUser) {
	u.externalUserByCompanyId[companyId] = owner
}

// RemoveOwnerOf is removeOwnerOf(short companyId).
func (u *ExternalUsers) RemoveOwnerOf(companyId int) {
	delete(u.externalUserByCompanyId, companyId)
}

// identifyUser is externalUserService.identifyUser(token). Java calls it with no
// null check, so a missing service is an NPE; the port logs it and reports "no
// user", which is what the login path then handles as a failed login.
func (u *ExternalUsers) identifyUser(token string) *robotmodel.ExternalUser {
	if u.externalUserService == nil {
		externalUsersLog.Warn("ExternalUsers: no ExternalUserService injected, cannot identify " + token)
		return nil
	}
	return u.externalUserService.IdentifyUser(token)
}

// gameOf is chatEvent.getOpenttd(), down-cast from the event's Openttd seam to
// the admin/model/Game the rules read.
func (u *ExternalUsers) gameOf(chatEvent *event.ChatEvent) *model.Game {
	var game, ok = chatEvent.GetOpenttd().(*model.Game)
	if ok {
		return game
	}
	return nil
}

// renameClient is renameClient(long clientId, String name): the rcon client_name
// behind a name the application allows.
func (u *ExternalUsers) renameClient(clientId int, name string) {
	var send = u.base.GetSend()
	if len(name) > 1 {
		if strings.HasPrefix(name, "Player") {
			send.ChatClient(int64(clientId), "Player... names are not allowed, try again...")
		} else {
			// Java loops while the name holds a space, replacing every one of them.
			var renamed = strings.ReplaceAll(name, " ", "_")
			send.Rcon("client_name " + util.FormatInt(int64(clientId)) + " " + renamed)
		}
	} else {
		send.ChatClient(int64(clientId), "Try : !rename NewName")
	}
}

func (u *ExternalUsers) showCompanyOwned(clientId int, companyName string) {
	u.base.GetSend().ChatClient(int64(clientId), "Congratulation ! You now own "+companyName+".")
}

func (u *ExternalUsers) showCompanyAlreadyOwned(clientId int, companyName string) {
	u.base.GetSend().ChatClient(int64(clientId), "You already own "+companyName+".")
}

// showCompanyAlreadyOwnedBy is the three argument overload of
// showCompanyAlreadyOwned. Java spells the missing space in
// companyName + "is already owned by" the same way, and the port keeps that
// byte-for-byte.
func (u *ExternalUsers) showCompanyAlreadyOwnedBy(clientId int, companyName string, owner *robotmodel.ExternalUser) {
	u.base.GetSend().ChatClient(int64(clientId), companyName+"is already owned by "+owner.GetName())
}

func (u *ExternalUsers) showLoginFailed(clientId int) {
	u.base.GetSend().ChatClient(int64(clientId), "Login failed.")
}

func (u *ExternalUsers) showLoginSucceed(clientId int) {
	u.base.GetSend().ChatClient(int64(clientId), "Login succeed.")
}

// showHowtoLogin is showHowtoLogin(long clientId): the four lines that tell a
// player how to get a token from www.strategyboard.net.
func (u *ExternalUsers) showHowtoLogin(clientId int) {
	var send = u.base.GetSend()
	send.ChatClient(int64(clientId), "How to login ***")
	send.ChatClient(int64(clientId), "1. Goto www.strategyboard.net and register there,")
	send.ChatClient(int64(clientId), "2. Login there and click 'In game login',")
	send.ChatClient(int64(clientId), "3. Copy and paste the whole !login command line into the chat window of the game.")
}

func (u *ExternalUsers) showClientAlreadyLogged(clientId int) {
	u.base.GetSend().ChatClient(int64(clientId), "You are already logged.")
}
