// OpenttdAdmin mirrors com.openttd.admin.OpenttdAdmin: the facade the embedding
// applications use. It owns the three layers - the EventDispatcher that hands out
// functional events, the Game that translates packets into them, and the
// NetworkAdmin that decodes the admin protocol - and wires them together.
//
// Java's addListener(Class, EventListener) becomes AddListener(EventKind, ...)
// here, since the port has no Class objects; the kind is what an event's Kind()
// answers. isConnected() reads IsRunning() rather than Java's isAlive(): the
// port's NetworkAdmin exposes the running flag, which is what that thread's
// liveness amounts to.
//
// Note the import of com.openttd/network/admin, whose package is also called
// admin: this package's own names are unqualified, so `admin.X` below always
// means the listener layer.

package admin

import "com.openttd/admin/event"
import "com.openttd/admin/model"
import "com.openttd/gamescript"
import "com.openttd/network/admin"
import "com.openttd/network/core"

type OpenttdAdmin struct {
	configuration *core.Configuration
	networkAdmin *admin.NetworkAdmin
	executor *gamescript.GSExecutor
	eventDispatcher *event.EventDispatcher
	game *model.Game
}

// NewOpenttdAdmin is OpenttdAdmin(Configuration): the client layer only. The
// network layer and the model arrive in startup(), which is why getGame() and
// getSend() can be nil before it.
func NewOpenttdAdmin(configuration *core.Configuration) *OpenttdAdmin {
	var facade = &OpenttdAdmin{configuration: configuration}
	facade.eventDispatcher = event.NewEventDispatcher()
	// Java passes `this`: the executor reads the current sender out of the facade
	// on every send, so a reconnect's new sender is picked up. GetSender below
	// is the SenderProvider gamescript asks for.
	facade.executor = gamescript.NewGSExecutor(facade)
	facade.eventDispatcher.AddListener(event.EVENT_GAME_SCRIPT, facade.executor)
	return facade
}

// GetConfiguration is getConfiguration.
func (o *OpenttdAdmin) GetConfiguration() *core.Configuration { return o.configuration }

// GetSender is the SenderProvider side: getSend(), as gamescript sees it. The
// nil checks matter: a nil NetworkAdminSender in a gamescript.Sender is a non-
// nil interface, and the executor would then call a method on nothing.
func (o *OpenttdAdmin) GetSender() gamescript.Sender {
	if o.networkAdmin == nil {
		return nil
	}
	var send = o.networkAdmin.GetSend()
	if send == nil {
		return nil
	}
	return send
}

// IsConnected is isConnected(): the network layer exists and its listener thread
// is running.
func (o *OpenttdAdmin) IsConnected() bool {
	if o.networkAdmin == nil {
		return false
	}
	return o.networkAdmin.IsRunning()
}

// Startup is startup(): shut anything still running down first, then build a
// fresh model and network layer and start them. executor.Clear() last restarts
// the request numbering, which is what makes the gamescript ids start at "0"
// again for each connection.
func (o *OpenttdAdmin) Startup() {
	o.Shutdown()
	o.game = model.NewGame(o.eventDispatcher)
	o.networkAdmin = admin.NewNetworkAdmin(o.configuration)
	o.networkAdmin.SetGame(o.game)
	o.networkAdmin.SetGSExecutor(o.executor)
	o.eventDispatcher.Startup()
	o.networkAdmin.Startup()
	o.executor.Clear()
}

// Shutdown is shutdown(): stop the threads that are running. Both guards matter
// on the first call, when neither exists yet.
func (o *OpenttdAdmin) Shutdown() {
	if o.networkAdmin != nil && o.networkAdmin.IsRunning() {
		o.networkAdmin.Shutdown()
	}
	if o.eventDispatcher != nil && o.eventDispatcher.IsAlive() {
		o.eventDispatcher.Shutdown()
	}
}

// GetGame is getGame(): the game model built from the server's packets, null
// before startup().
func (o *OpenttdAdmin) GetGame() *model.Game { return o.game }

// GetSend is getSend(): the thread sending messages to the server, null until a
// connection has been made.
func (o *OpenttdAdmin) GetSend() *admin.NetworkAdminSender {
	if o.networkAdmin == nil {
		return nil
	}
	return o.networkAdmin.GetSend()
}

// GetGSExecutor is getGSExecutor(): the thread calling gamescript on the server.
func (o *OpenttdAdmin) GetGSExecutor() *gamescript.GSExecutor { return o.executor }

// AddListener is addListener(Class, EventListener), keyed by EventKind.
func (o *OpenttdAdmin) AddListener(kind event.EventKind, listener event.EventListener) {
	o.eventDispatcher.AddListener(kind, listener)
}

// RemoveListener is removeListener(Class, EventListener).
func (o *OpenttdAdmin) RemoveListener(kind event.EventKind, listener event.EventListener) {
	o.eventDispatcher.RemoveListener(kind, listener)
}

// GetEventDispatcher is not in Java: the facade only delegates the two listener
// methods. The port adds this for the tests, which need to see the queue drain.
func (o *OpenttdAdmin) GetEventDispatcher() *event.EventDispatcher { return o.eventDispatcher }
