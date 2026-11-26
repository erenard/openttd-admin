// CompanyLifeCycle is com.openttd.robot.rule.CompanyLifeCycle: the rule that
// polices who may start a company, how many companies one player may own, and
// how a company gets deleted again.
//
// Rules #1..#6 of the class comment are: only a logged player may create; only
// one company per ip address; "!resetme"; "$companies"/"$cps"; "!info";
// "$reset <companyId>". It listens for CompanyEvent and ChatEvent, and leans on
// the ExternalUsers rule for the identity side. See rule/AbstractRule.java for
// the base and rule/ExternalUsers.java for the rule it delegates to.
//
// There is no Java test for this class, so company_life_cycle_test.go is new
// coverage written against this source as the specification.

package rule

import "strings"
import "com.openttd/admin"
import "com.openttd/admin/event"
import "com.openttd/admin/model"
import netadmin "com.openttd/network/admin"
import "com.openttd/util"
import robotmodel "com.openttd/robot/model"

var companyLifeCycleLog = util.GetLogger("CompanyLifeCycle")

type CompanyLifeCycle struct {
	base          *AbstractRule
	externalUsers *ExternalUsers
	// companyIdByIpAddress is companyIdByIpAddress: the company a player's ip
	// address already owns. Java's value is a Short, boxed so the DELETE branch
	// can compare it; the port stores an int and compares directly.
	companyIdByIpAddress map[string]int
	// checkLogin and checkIpAddress are the two public configuration flags.
	checkLogin     bool
	checkIpAddress bool
}

// NewCompanyLifeCycle is CompanyLifeCycle(OpenttdAdmin openttdAdmin,
// ExternalUsers externalUsers). Both configuration flags default to true, as
// Java's field initializers do.
func NewCompanyLifeCycle(openttd *admin.OpenttdAdmin, externalUsers *ExternalUsers) *CompanyLifeCycle {
	var life = &CompanyLifeCycle{
		externalUsers:        externalUsers,
		companyIdByIpAddress: map[string]int{},
		checkLogin:           true,
		checkIpAddress:       true}
	life.base = NewAbstractRule(openttd, life, []event.EventKind{event.EVENT_COMPANY, event.EVENT_CHAT})
	return life
}

// SetCheckLogin is the public field checkLogin.
func (c *CompanyLifeCycle) SetCheckLogin(check bool) { c.checkLogin = check }

// SetCheckIpAddress is the public field checkIpAddress.
func (c *CompanyLifeCycle) SetCheckIpAddress(check bool) { c.checkIpAddress = check }

// OnCompanyEvent is onCompanyEvent(CompanyEvent).
func (c *CompanyLifeCycle) OnCompanyEvent(companyEvent *event.CompanyEvent) {
	if companyLifeCycleLog.IsDebugEnabled() {
		companyLifeCycleLog.Debug("CompanyLifeCycle.onCompanyEvent " + companyEvent.ToString())
	}
	var companyId = companyEvent.GetCompanyId()
	var action = companyEvent.GetAction()
	var game = c.gameOf(companyEvent.GetOpenttd())
	if action == event.ACTION_CREATE {
		if c.applyRules(companyId, game) {
			// Find the company owner: the first client of the new company, if
			// that client is logged in.
			var companyOwner = c.findOwner(companyId, game)
			if companyOwner != nil {
				// First remove ownership of every other company, then own this
				// one. Java iterates the companies while removing from the owner
				// map; the port reads the list first, which is already a copy.
				for _, company := range game.GetCompanies() {
					var owner = c.externalUsers.GetOwnerOf(company.GetId())
					if owner != nil && companyOwner.Equals(owner) {
						c.externalUsers.RemoveOwnerOf(company.GetId())
					}
				}
				c.externalUsers.SetOwnerOf(companyId, companyOwner)
			}
		} else {
			c.deleteCompany(game, companyId)
		}
	} else if action == event.ACTION_DELETE {
		// Keep the local model up to date. Java's DELETE case has no break, but
		// there is no default case to fall into, so nothing else happens.
		c.externalUsers.RemoveOwnerOf(companyId)
		var stale = []string{}
		for ip, ipCompanyId := range c.companyIdByIpAddress {
			if ipCompanyId == companyId {
				stale = append(stale, ip)
			}
		}
		for _, ip := range stale {
			delete(c.companyIdByIpAddress, ip)
		}
	}
}

// findOwner is the block that reads game.getClients(companyId).iterator().next()
// and looks that client up in the users rule.
func (c *CompanyLifeCycle) findOwner(companyId int, game *model.Game) *robotmodel.ExternalUser {
	var clients = game.GetClientsOf(companyId)
	if len(clients) == 0 {
		return nil
	}
	var creator = clients[0]
	if creator == nil {
		return nil
	}
	return c.externalUsers.GetExternalUser(creator.GetId())
}

// applyRules is applyRules(short companyId, Game game): the two creation rules.
// It reports false when the company must be deleted again.
func (c *CompanyLifeCycle) applyRules(companyId int, game *model.Game) bool {
	var send = c.base.GetSend()
	// Find the company owner, its creator and that creator's ip address.
	var companyOwner *robotmodel.ExternalUser = nil
	var creator *netadmin.Client = nil
	var ipAddress = ""
	var clients = game.GetClientsOf(companyId)
	if len(clients) > 0 {
		creator = clients[0]
	}
	if creator != nil {
		ipAddress = creator.GetIp()
		companyOwner = c.externalUsers.GetExternalUser(creator.GetId())
	}
	// Verify Rule #1: only a logged player can create.
	if c.checkLogin && companyOwner == nil {
		if creator != nil {
			send.ChatClient(int64(creator.GetId()), "Company "+util.FormatInt(int64(companyId))+
				" creation rejected: Please loggin before start a new company.")
			c.externalUsers.showHowtoLogin(creator.GetId())
		}
		return false
	}
	// Verify Rule #2: only one company by ip address can exist. An admin is
	// exempt, and so is a player the application does not know.
	if c.checkIpAddress && ipAddress != "" {
		if companyOwner == nil || !companyOwner.IsAdmin() {
			var previousCompanyId, found = c.companyIdByIpAddress[ipAddress]
			if found {
				var previousCompany = game.GetCompany(previousCompanyId)
				if previousCompany != nil {
					var company = game.GetCompany(companyId)
					// DEVIATION FROM JAVA: Java calls getCompany(companyId).getName()
					// without a null check, which throws when the new company is not
					// in the model yet; the port reads the name only when it is there.
					var companyName = ""
					if company != nil {
						companyName = company.GetName()
					}
					if creator != nil {
						send.ChatClient(int64(creator.GetId()), "Company '"+companyName+
							"' creation rejected: You ("+ipAddress+") already own '"+
							previousCompany.GetName()+"'.")
					}
					return false
				}
			}
		}
		// Add the company.
		c.companyIdByIpAddress[ipAddress] = companyId
	}
	return true
}

// deleteCompany is deleteCompany(Game game, short companyId): the rcon commands
// that empty a company and hand its players back to the spectator slot.
func (c *CompanyLifeCycle) deleteCompany(game *model.Game, companyId int) {
	var send = c.base.GetSend()
	for _, client := range game.GetClientsOf(companyId) {
		send.Rcon("move " + util.FormatInt(int64(client.GetId())) + " 255")
	}
	send.Rcon("reset_company " + util.FormatInt(int64(companyId+1)))
}

// OnChatEvent is onChatEvent(ChatEvent).
func (c *CompanyLifeCycle) OnChatEvent(chatEvent *event.ChatEvent) {
	if companyLifeCycleLog.IsDebugEnabled() {
		companyLifeCycleLog.Debug("CompanyLifeCycle.onChatEvent " + chatEvent.ToString())
	}
	var clientId = chatEvent.GetClientId()
	var message = chatEvent.GetMessage()
	var game = c.gameOf(chatEvent.GetOpenttd())
	if message == "" {
		// Java's `if (message != null)`.
		return
	}
	var command = strings.ToLower(strings.TrimSpace(message))
	if command == "!resetme" {
		// Rule #3.
		var client = game.GetClient(clientId)
		if client != nil {
			var companyId = client.GetCompany()
			if companyId >= 0 && companyId < COMPANY_NONE {
				c.deleteCompany(game, companyId)
			}
		}
	} else if command == "!info" {
		// Rule #5.
		c.showInfo(game, clientId)
	} else {
		var user = c.externalUsers.GetExternalUser(clientId)
		if user != nil && user.IsAdmin() {
			if command == "$companies" || command == "$cps" {
				// Rule #4.
				c.showCompanies(game, clientId)
			} else if strings.HasPrefix(command, "$reset") || strings.HasPrefix(command, "$r") {
				// Rule #6. Java catches only NumberFormatException; a message
				// with no argument would raise ArrayIndexOutOfBoundsException there,
				// and the port sends the usage line instead.
				var argument = c.argumentAfter(command)
				var companyId, ok = ParseShort(argument)
				if !ok {
					c.showMessage(clientId, "Usage: $reset companyId (try $companies to find companyIds)")
					return
				}
				c.deleteCompany(game, companyId)
			}
		}
	}
}

// argumentAfter is message.split(" ")[1].trim(): the first argument after the
// command, or the empty string when the message has none, which parseShort
// reports as unparsable.
func (c *CompanyLifeCycle) argumentAfter(command string) string {
	var parts = strings.Split(command, " ")
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func (c *CompanyLifeCycle) showMessage(clientId int, message string) {
	c.base.GetSend().ChatClient(int64(clientId), message)
}

// showCompanies is showCompanies(Game openttd, Integer clientId): the leaderboard
// of companies, one chat line each, in Java's "# %3d, %255s, %255s, %4d" layout.
func (c *CompanyLifeCycle) showCompanies(game *model.Game, clientId int) {
	var send = c.base.GetSend()
	send.ChatClient(int64(clientId), "#Id, Name, CEO, Inauguration.")
	for _, company := range game.GetCompanies() {
		if company != nil {
			var owner = c.externalUsers.GetOwnerOf(company.GetId())
			// DEVIATION FROM JAVA: Java calls getOwnerOf(...).getName() with no
			// null check, so an unowned company throws; the port shows an empty
			// CEO column for it.
			var ownerName = ""
			if owner != nil {
				ownerName = owner.GetName()
			}
			send.ChatClient(int64(clientId), "#"+
				"# "+util.PadLeft(util.FormatInt(int64(company.GetId())), 3)+
				", "+util.PadRight(company.GetName(), 255)+
				", "+util.PadRight(ownerName, 255)+
				", "+util.PadLeft(util.FormatInt(company.GetInauguratedYear()), 4))
		}
	}
}

// showInfo is showInfo(Game openttd, Integer clientId): what a player is told
// about themselves.
func (c *CompanyLifeCycle) showInfo(game *model.Game, clientId int) {
	var send = c.base.GetSend()
	var client = game.GetClient(clientId)
	if client == nil {
		// DEVIATION FROM JAVA: Java dereferences the client without checking.
		return
	}
	var userName = "Anonymous"
	var user = c.externalUsers.GetExternalUser(clientId)
	if user != nil {
		userName = user.GetName()
	}
	send.ChatClient(int64(clientId), "User: "+userName+", play as ("+client.GetName()+").")
	var companyId = client.GetCompany()
	if companyId == COMPANY_NONE {
		send.ChatClient(int64(clientId), "Spectator.")
	} else {
		var company = game.GetCompany(companyId)
		var owner = c.externalUsers.GetOwnerOf(companyId)
		if company != nil && owner != nil {
			send.ChatClient(int64(clientId), "Company: "+company.GetName()+", owned by "+owner.GetName()+".")
		}
	}
}

// gameOf is companyEvent.getOpenttd() / chatEvent.getOpenttd(), down-cast from
// the event's Openttd seam to the admin/model/Game the rules read.
func (c *CompanyLifeCycle) gameOf(source event.Openttd) *model.Game {
	var game, ok = source.(*model.Game)
	if ok {
		return game
	}
	return nil
}
