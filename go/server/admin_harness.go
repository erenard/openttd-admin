// AdminHarness mirrors com.openttd.server.AdminHarness: the OpenttdAdmin facade
// wired to a FakeAdminServer, so no OpenTTD process is needed and the whole
// admin protocol is exercisable in a unit test.
//
// It is what the stage-5 sender tests used to build by hand: the facade now
// exists, so the harness goes through it exactly as the Java one does.

package server

import "com.openttd"
import "com.openttd/admin"
import "com.openttd/admin/model"
import "com.openttd/gamescript"
import "com.openttd/network/core"

// PASSWORD and ADMIN_NAME are AdminHarness.PASSWORD and ADMIN_NAME: what the
// harness joins as, and what the sender test reads back off the wire in its
// join case.
const PASSWORD = "test-password"
const ADMIN_NAME = "openttd-admin-test"

type AdminHarness struct {
	server *FakeAdminServer
	configuration *core.Configuration
	admin *admin.OpenttdAdmin
	recorder *EventRecorder
}

// NewAdminHarness is AdminHarness().
func NewAdminHarness() (*AdminHarness, error) {
	return NewAdminHarnessWith(PASSWORD, ADMIN_NAME)
}

// NewAdminHarnessWith is AdminHarness(String password, String name).
func NewAdminHarnessWith(password string, name string) (*AdminHarness, error) {
	var server, serverError = NewFakeAdminServer()
	if serverError != nil {
		return nil, serverError
	}
	var configuration = core.NewConfiguration()
	configuration.Password = password
	configuration.Name = name
	configuration.Host = "localhost"
	configuration.AdminPort = server.GetPort()

	var facade = admin.NewOpenttdAdmin(configuration)
	var recorder = NewEventRecorder()
	RegisterWith(facade, recorder)
	facade.Startup()
	// The server answers the JOIN with PROTOCOL + WELCOME, exactly as
	// ServerNetworkAdminSocketHandler::SendProtocol() and SendWelcome() do.
	server.Handshake()

	// The game model only exists once a packet has been decoded, so a missing
	// model or a missing GameInfo means "not connected yet", not a protocol
	// failure - which is what Java's caught NullPointerException inside the
	// awaited condition amounts to.
	var connected = func() bool {
		var game = facade.GetGame()
		if game == nil {
			return false
		}
		var gameInfo = game.GetGameInfo()
		if gameInfo == nil {
			return false
		}
		return gameInfo.GetServerName() != ""
	}
	openttd.Await(connected, 10, "the WELCOME packet")

	return &AdminHarness{
		server: server,
		configuration: configuration,
		admin: facade,
		recorder: recorder,
	}, nil
}

// GetServer is getServer.
func (h *AdminHarness) GetServer() *FakeAdminServer { return h.server }

// GetConfiguration is getConfiguration.
func (h *AdminHarness) GetConfiguration() *core.Configuration { return h.configuration }

// GetAdmin is getAdmin.
func (h *AdminHarness) GetAdmin() *admin.OpenttdAdmin { return h.admin }

// GetRecorder is getRecorder.
func (h *AdminHarness) GetRecorder() *EventRecorder { return h.recorder }

// GetGame is getGame(): the game model built from the server's packets. It lives
// in admin/model, a different package from the facade this file calls `admin`.
func (h *AdminHarness) GetGame() *model.Game {
	return h.admin.GetGame()
}

// GetSend is getSend(): the thread sending messages to the server, which the
// facade only has once a connection has been made, hence the await. The type is
// the alias from types.go: this file imports the facade, whose package is also
// called admin, so it cannot import network/admin as well.
func (h *AdminHarness) GetSend() *NetworkAdminSender {
	openttd.Await(func() bool { return h.admin.GetSend() != nil }, 10, "the admin sender thread")
	return h.admin.GetSend()
}

// GetGSExecutor is getGSExecutor.
func (h *AdminHarness) GetGSExecutor() *gamescript.GSExecutor { return h.admin.GetGSExecutor() }

// Close is close(): stop the client, then the server. shutdown() sends the QUIT
// packet and interrupts the reconnect sleep, so it returns in milliseconds.
func (h *AdminHarness) Close() {
	h.admin.Shutdown()
	h.server.Close()
}
