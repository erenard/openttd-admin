// AdminPing is com.openttd.demo.AdminPing: the smallest thing this library can
// do - connect to the admin port without the OpenttdAdmin facade, and send one
// ping. It builds the two pieces by hand, the EventDispatcher and the NetworkAdmin,
// which is what the facade does for its users.
//
// Like the Java original the dispatcher it creates is never handed to the
// network layer, so nothing this demo does reaches it: the listener thread decodes
// the packets and drops them, because NetworkAdmin has no game model to update.
// That is Java's behaviour too, and the port keeps it rather than quietly
// improving the sample.
//
// There is no Java test for this class, so demo_test.go is new coverage written
// against this source as the specification.

package demo

import "time"
import "com.openttd/admin/event"
import "com.openttd/network/admin"
import "com.openttd/network/core"
import "com.openttd/util"

var adminPingLog = util.GetLogger("AdminPing")

type AdminPing struct {
	networkAdmin    *admin.NetworkAdmin
	eventDispatcher *event.EventDispatcher
}

// NewAdminPing is AdminPing(Configuration): the two pieces Java's constructor
// builds, the network layer and the dispatcher it never connects.
func NewAdminPing(configuration *core.Configuration) *AdminPing {
	return &AdminPing{
		eventDispatcher: event.NewEventDispatcher(),
		networkAdmin:    admin.NewNetworkAdmin(configuration),
	}
}

// Start is start(): the dispatcher first, then the listener thread.
func (p *AdminPing) Start() {
	p.eventDispatcher.Startup()
	p.networkAdmin.Startup()
}

// Ping is ping(): one ADMIN_PACKET_ADMIN_PING carrying the current time, which
// is all Java's version does - the identifier only has to be unique enough for
// the server's pong to be recognisable.
//
// DEVIATION FROM JAVA: getSend() is null until a connection has been made, where
// Java's getSend().ping() raises a NullPointerException. The port logs and does
// nothing, which is what the other nil guards of this port do.
func (p *AdminPing) Ping() {
	var send = p.networkAdmin.GetSend()
	if send == nil {
		adminPingLog.Info("AdminPing: no admin sender yet, ping not sent.")
		return
	}
	send.Ping(time.Now().UnixMilli())
}

// Shutdown is the port's addition: Java's demo never stops its threads, so a
// test (or a caller that wants its process to end) has no way in. It is the
// pair of Start(), in the order the facade's shutdown() uses.
func (p *AdminPing) Shutdown() {
	p.networkAdmin.Shutdown()
	p.eventDispatcher.Shutdown()
}

// GetNetworkAdmin is the port's accessor for the private field, so a caller can
// await the sender or stop the client itself.
func (p *AdminPing) GetNetworkAdmin() *admin.NetworkAdmin { return p.networkAdmin }

// GetEventDispatcher is the port's accessor for the dispatcher field.
func (p *AdminPing) GetEventDispatcher() *event.EventDispatcher { return p.eventDispatcher }

// AdminPingMain is main(String[] args): the console demo of the ping. The name
// carries the class, because Java gave every class its own main() and a Go
// package has one namespace.
func AdminPingMain(args []string) {
	var configuration = core.NewConfiguration()
	configuration.Password = "lepasswd"
	var simpleAdmin = NewAdminPing(configuration)
	simpleAdmin.Start()
	simpleAdmin.Ping()
	adminPingLog.Info("Openttd admin started")
}
