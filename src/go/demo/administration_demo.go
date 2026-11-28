// AdministrationDemo is com.openttd.demo.AdministrationDemo: the login demo. It
// installs the two rules a goal server needs for its players to identify
// themselves - ExternalUsers and the admin-gated Administration - plus a rule of
// its own that greets the players once a week, and it puts TestUtil's fake
// services into the ExternalServices singleton so any token works and a token
// starting with "admin" gets moderation powers.
//
// Java extends OpenttdAdmin and installs the rules in its constructor; the port
// keeps the facade in a field, and keeps the three rules as fields too, where
// Java lets them go - a rule that is not kept alive is still registered, but a
// caller reading its own robot back has nothing to read.
//
// There is no Java test for this class, so demo_test.go is new coverage written
// against this source as the specification.

package demo

import "time"
import "com.openttd/admin"
import "com.openttd/admin/event"
import "com.openttd/network/core"
import "com.openttd/robot/rule"
import "com.openttd/util"

type IngameInfo struct {
	base *rule.AbstractRule
}

// NewIngameInfo is IngameInfo(OpenttdAdmin): the super(openttdAdmin) call plus
// the DateEvent kind listEventTypes() returns.
func NewIngameInfo(openttd *admin.OpenttdAdmin) *IngameInfo {
	var info = &IngameInfo{}
	info.base = rule.NewAbstractRule(openttd, info, []event.EventKind{event.EVENT_DATE})
	return info
}

// OnDateEvent is onDateEvent(DateEvent): the four lines that tell a player what
// this server can do for them, once a week.
func (i *IngameInfo) OnDateEvent(dateEvent *event.DateEvent) {
	var date = dateEvent.GetOpenttd().GetDate()
	if date.DayOfWeek() == util.MONDAY {
		var send = i.base.GetSend()
		send.ChatBroadcast("Welcome to this login demo.")
		send.ChatBroadcast("Type '!login username' to log on (any name will work).")
		send.ChatBroadcast("If your username starts with 'admin', you will have moderation powers.")
		send.ChatBroadcast("Once logged, type '$help' to know more about moderation commands.")
	}
}

type AdministrationDemo struct {
	openttdAdmin  *admin.OpenttdAdmin
	externalUsers *rule.ExternalUsers
	// administration is the rule the login demo is named after. Java builds it
	// without keeping it.
	administration *rule.Administration
	ingameInfo     *IngameInfo
}

// NewAdministrationDemo is AdministrationDemo(Configuration), i.e. Java's
// super(configuration) and the three rules its constructor body installs. The
// services have to be in the ExternalServices singleton before it runs: the
// rules read them in their field initializers, exactly as Java's do.
func NewAdministrationDemo(configuration *core.Configuration) *AdministrationDemo {
	var robot = &AdministrationDemo{openttdAdmin: admin.NewOpenttdAdmin(configuration)}
	robot.externalUsers = rule.NewExternalUsers(robot.openttdAdmin)
	robot.administration = rule.NewAdministration(robot.openttdAdmin, robot.externalUsers)
	robot.ingameInfo = NewIngameInfo(robot.openttdAdmin)
	return robot
}

// Startup is the inherited startup().
func (d *AdministrationDemo) Startup() { d.openttdAdmin.Startup() }

// Shutdown is the inherited shutdown().
func (d *AdministrationDemo) Shutdown() { d.openttdAdmin.Shutdown() }

// GetOpenttd is the port's accessor for the facade this class used to inherit.
func (d *AdministrationDemo) GetOpenttd() *admin.OpenttdAdmin { return d.openttdAdmin }

// GetExternalUsers is the port's accessor for the login rule.
func (d *AdministrationDemo) GetExternalUsers() *rule.ExternalUsers { return d.externalUsers }

// GetAdministration is the port's accessor for the moderation rule.
func (d *AdministrationDemo) GetAdministration() *rule.Administration { return d.administration }

// GetIngameInfo is the port's accessor for the greeting rule.
func (d *AdministrationDemo) GetIngameInfo() *IngameInfo { return d.ingameInfo }

// AdministrationDemoMain is main(String[] args).
//
// DEVIATION FROM JAVA, WHICH IS ALSO A FIX: Java's main() builds a HelloWorldDemo
// - a copy-paste slip, so the three rules this class installs are never started
// and the demo waits three minutes showing hello-world instead. The port starts
// the demo it is part of.
func AdministrationDemoMain(args []string) {
	var configuration = core.NewConfiguration()
	if !ParseArguments(args, configuration) {
		return
	}
	FakeExternalUserService()
	var robot = NewAdministrationDemo(configuration)
	robot.Startup()
	time.Sleep(time.Duration(ADMINISTRATION_WAIT_MILLIS) * time.Millisecond)
	robot.Shutdown()
}

// ADMINISTRATION_WAIT_MILLIS is Java's Thread.sleep(180000): the three minutes
// the demo stays connected for.
const ADMINISTRATION_WAIT_MILLIS = 180000
