// OpenttdAdminConsole is com.openttd.demo.OpenttdAdminConsole: the facade with
// nothing added but a log line per event, which is the quickest way to see what
// a server actually sends an admin client.
//
// Java extends OpenttdAdmin and implements the four listener interfaces, so the
// one object is both the client and its listener. The port keeps the facade in a
// field and delegates the three methods main() calls on it, which keeps the
// sample's shape while leaving the facade's own API untouched.
//
// There is no Java test for this class, so demo_test.go is new coverage written
// against this source as the specification.

package demo

import "com.openttd/admin"
import "com.openttd/admin/event"
import "com.openttd/network/core"
import "com.openttd/util"

var consoleLog = util.GetLogger("OpenttdAdminConsole")

type OpenttdAdminConsole struct {
	openttd *admin.OpenttdAdmin
}

// NewOpenttdAdminConsole is OpenttdAdminConsole(Configuration), i.e. Java's
// super(configuration). The listeners are registered by the main, as Java's is.
func NewOpenttdAdminConsole(configuration *core.Configuration) *OpenttdAdminConsole {
	return &OpenttdAdminConsole{openttd: admin.NewOpenttdAdmin(configuration)}
}

// AddListener is the inherited addListener(Class, EventListener).
func (c *OpenttdAdminConsole) AddListener(kind event.EventKind, listener event.EventListener) {
	c.openttd.AddListener(kind, listener)
}

// Startup is the inherited startup().
func (c *OpenttdAdminConsole) Startup() { c.openttd.Startup() }

// Shutdown is the inherited shutdown().
func (c *OpenttdAdminConsole) Shutdown() { c.openttd.Shutdown() }

// GetOpenttd is the port's accessor for the facade this class used to inherit.
func (c *OpenttdAdminConsole) GetOpenttd() *admin.OpenttdAdmin { return c.openttd }

// OnDateEvent is onDateEvent(DateEvent).
func (c *OpenttdAdminConsole) OnDateEvent(dateEvent *event.DateEvent) {
	consoleLog.Info("Date event " + dateEvent.ToString())
}

// OnClientEvent is onClientEvent(ClientEvent).
func (c *OpenttdAdminConsole) OnClientEvent(clientEvent *event.ClientEvent) {
	consoleLog.Info("Client event " + clientEvent.ToString())
}

// OnChatEvent is onChatEvent(ChatEvent). The colon is Java's, and only the chat
// and company lines have it.
func (c *OpenttdAdminConsole) OnChatEvent(chatEvent *event.ChatEvent) {
	consoleLog.Info("Chat event : " + chatEvent.ToString())
}

// OnCompanyEvent is onCompanyEvent(CompanyEvent).
func (c *OpenttdAdminConsole) OnCompanyEvent(companyEvent *event.CompanyEvent) {
	consoleLog.Info("Company event : " + companyEvent.ToString())
}

// OpenttdAdminConsoleMain is main(String[] args): the password is the only
// argument this sample needs, since it connects to the default localhost:3977.
//
// DEVIATION FROM JAVA: args[0] is read unguarded there, so an argument-less
// command line ends in ArrayIndexOutOfBoundsException; the port reports the
// usage instead.
func OpenttdAdminConsoleMain(args []string) {
	if len(args) < 1 {
		consoleLog.Error("Usage: openttd-admin-console admin_password")
		return
	}
	var configuration = core.NewConfiguration()
	configuration.Password = args[0]

	var simpleAdmin = NewOpenttdAdminConsole(configuration)
	simpleAdmin.AddListener(event.EVENT_DATE, simpleAdmin)
	simpleAdmin.AddListener(event.EVENT_CHAT, simpleAdmin)
	simpleAdmin.AddListener(event.EVENT_CLIENT, simpleAdmin)
	simpleAdmin.AddListener(event.EVENT_COMPANY, simpleAdmin)
	simpleAdmin.Startup()
	consoleLog.Info("Openttd admin started")
}
