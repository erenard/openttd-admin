// HelloWorldDemo is com.openttd.demo.HelloWorldDemo: the smallest rule this
// library can teach - one listener, one event kind, one chat message - wrapped
// in the smallest robot that runs it. Its inner HelloWorld class is the sample
// the other demos' comments point at: extending AbstractRule and implementing
// DateEventListener is all it takes.
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

type HelloWorldDemo struct {
	openttdAdmin *admin.OpenttdAdmin
	helloworld   *HelloWorld
}

// NewHelloWorldDemo is HelloWorldDemo(Configuration): the facade and the rule
// Java's constructor builds.
func NewHelloWorldDemo(configuration *core.Configuration) *HelloWorldDemo {
	var robot = &HelloWorldDemo{openttdAdmin: admin.NewOpenttdAdmin(configuration)}
	robot.helloworld = NewHelloWorld(robot.openttdAdmin)
	return robot
}

// Startup is startup().
func (d *HelloWorldDemo) Startup() { d.openttdAdmin.Startup() }

// Shutdown is shutdown().
func (d *HelloWorldDemo) Shutdown() { d.openttdAdmin.Shutdown() }

// GetOpenttd is the port's accessor for the facade field, which Java's users
// reach through the inheritance this port replaced with a field.
func (d *HelloWorldDemo) GetOpenttd() *admin.OpenttdAdmin { return d.openttdAdmin }

// GetRule is the port's accessor for the rule field.
func (d *HelloWorldDemo) GetRule() *HelloWorld { return d.helloworld }

type HelloWorld struct {
	base *rule.AbstractRule
}

// NewHelloWorld is HelloWorld(OpenttdAdmin): the super(openttdAdmin) call, which
// in the port means registering the kinds listEventTypes() returns - DateEvent
// only.
func NewHelloWorld(openttd *admin.OpenttdAdmin) *HelloWorld {
	var helloworld = &HelloWorld{}
	helloworld.base = rule.NewAbstractRule(openttd, helloworld, []event.EventKind{event.EVENT_DATE})
	return helloworld
}

// OnDateEvent is onDateEvent(DateEvent): the greeting once a week. The date comes
// from the game model the event carries, which is the event.Openttd seam's one
// method, so unlike the robot's rules this one needs no down-cast.
func (h *HelloWorld) OnDateEvent(dateEvent *event.DateEvent) {
	var date = dateEvent.GetOpenttd().GetDate()
	if date.DayOfWeek() == util.MONDAY {
		h.base.GetSend().ChatBroadcast("Hello world !!!")
	}
}

// HelloWorldDemoMain is main(String[] args): connect, wait a minute,
// disconnect. The name carries the class, because Java gave every class its own
// main() and a Go package has one namespace.
//
// DEVIATION FROM JAVA: parseArguments() returns whether the command line was
// usable, and the port stops when it was not, where Java keeps going and
// connects to the localhost:3977 defaults.
func HelloWorldDemoMain(args []string) {
	var configuration = core.NewConfiguration()
	if !ParseArguments(args, configuration) {
		return
	}
	var robot = NewHelloWorldDemo(configuration)
	robot.Startup()
	time.Sleep(time.Duration(HELLO_WORLD_WAIT_MILLIS) * time.Millisecond)
	robot.Shutdown()
}

// HELLO_WORLD_WAIT_MILLIS is Java's Thread.sleep(60000): the minute the demo stays
// connected for.
const HELLO_WORLD_WAIT_MILLIS = 60000
