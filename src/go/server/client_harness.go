// ClientHarness is the game-port twin of AdminHarness: the OpenttdClient facade
// wired to a FakeGameServer, so the whole client protocol is exercisable without
// an OpenTTD process.
//
// Java has no such class: the client layer's only test is
// OpenttdClientIntegrationTest, and that is excluded from `gradle integrationTest`
// because the Java TcpGame.PacketGameType is stale for 15.3 (the join is refused
// with NETWORK_ERROR_WRONG_REVISION). The fake server scripts exactly the packets
// that stale enum would have missed, which is what makes this harness worth
// having.
//
// The types from src/go/network/client are spelled through the aliases in types.go:
// this file imports the facade, whose package is also called client, so it
// cannot import the network layer as well.
package server

import "com.openttd"
import "com.openttd/client"
import "com.openttd/network/core"

// CLIENT_NAME is what the harness joins as: the name the sender puts in its
// PACKET_CLIENT_JOIN, which the sender test reads back off the wire.
const CLIENT_NAME = "openttd-client-test"

type ClientHarness struct {
	server        *FakeGameServer
	configuration *core.Configuration
	facade        *client.OpenttdClient
}

// NewClientHarness is the constructor: a fake game server on an ephemeral port,
// a configuration pointing at it, the facade started, and the handshake answered
// the way a 15.3 server that needs no NewGRF check and no game password answers
// a join: straight with PACKET_SERVER_WELCOME.
func NewClientHarness() (*ClientHarness, error) {
	var gameServer, serverError = NewFakeGameServer()
	if serverError != nil {
		return nil, serverError
	}
	var configuration = core.NewConfiguration()
	configuration.Password = PASSWORD
	configuration.Name = CLIENT_NAME
	configuration.Host = "localhost"
	configuration.ClientPort = gameServer.GetPort()

	var facade = client.NewOpenttdClient(configuration)
	facade.Startup()
	gameServer.Handshake()

	// The model only carries the WELCOME's payload, so an empty server id means
	// "not connected yet" rather than a protocol failure.
	var connected = func() bool {
		var model = facade.GetModel()
		if model == nil {
			return false
		}
		return model.PasswordServerId != ""
	}
	openttd.Await(connected, 10, "the game client to receive WELCOME")

	return &ClientHarness{
		server:        gameServer,
		configuration: configuration,
		facade:        facade,
	}, nil
}

// GetServer is getServer: the fake game server, to script the packets a real
// one would send and to read back the ones the client sent.
func (h *ClientHarness) GetServer() *FakeGameServer { return h.server }

// GetConfiguration is getConfiguration.
func (h *ClientHarness) GetConfiguration() *core.Configuration { return h.configuration }

// GetClient is getClient: the facade.
func (h *ClientHarness) GetClient() *client.OpenttdClient { return h.facade }

// GetSend is getSend(): the thread sending the client's packets. The facade only
// has one once a connection has been made, hence the await.
func (h *ClientHarness) GetSend() *NetworkClientSender {
	openttd.Await(func() bool { return h.facade.GetSend() != nil }, 10, "the game client sender thread")
	return h.facade.GetSend()
}

// GetModel is getModel(): the state decoded out of the server's packets.
func (h *ClientHarness) GetModel() *NetworkClientModel {
	return h.facade.GetModel()
}

// Close is close(): stop the client, then the server. Shutdown() sends PACKET
// CLIENT_QUIT, closes the socket so a blocked receive() returns and cuts the
// reconnect sleep short, so it comes back in milliseconds.
func (h *ClientHarness) Close() {
	h.facade.Shutdown()
	h.server.Close()
}
