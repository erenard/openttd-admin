// OpenttdClient mirrors com.openttd.client.OpenttdClient: the facade an
// embedding application uses to sit on the game port (3979), the client twin of
// the admin port's OpenttdAdmin. strategyboard's OpenttdClientHandler is the
// consumer that pins this API: new Configuration(); conf.clientPort = port;
// new OpenttdClient(conf); client.startup(); ... client.isConnected().
//
// Unlike the admin facade there is no event layer, no game model and no
// gamescript here: NetworkClient only decodes the server's packets into its own
// NetworkClientModel and answers the handshake, so the facade is the three
// methods Java has.
//
// This package is `client` and the network layer it wraps is `client` too
// (src/go/network/client): an import binds the declared package name, so every
// reference below to `client.` means the network layer, and the types declared
// in this file are spelled unqualified.
package client

import "com.openttd/network/client"
import "com.openttd/network/core"

type OpenttdClient struct {
	configuration *core.Configuration

	// networkClient is Java's field: the listener thread of the current
	// connection, replaced by Startup() and dropped by Shutdown().
	networkClient *client.NetworkClient
}

// NewOpenttdClient is the constructor: it only keeps the configuration, as in
// Java, so the network layer can be rebuilt by Startup() without one.
func NewOpenttdClient(configuration *core.Configuration) *OpenttdClient {
	return &OpenttdClient{configuration: configuration}
}

// IsConnected is isConnected(): networkClient != null && networkClient.isAlive().
// The port's stand-in for isAlive() is IsRunning(), which Startup() sets before
// the thread starts, so this is true as soon as Startup() returns.
func (c *OpenttdClient) IsConnected() bool {
	if c.networkClient == nil {
		return false
	}
	return c.networkClient.IsRunning()
}

// Startup is startup(): safety shutdown of whatever is running, then a fresh
// network layer on the same configuration. A new NetworkClient means a new
// NetworkClientModel, so a reconnect does not inherit the previous game's
// client id or ack frame.
func (c *OpenttdClient) Startup() {
	c.Shutdown()
	c.networkClient = client.NewNetworkClient(c.configuration)
	c.networkClient.Startup()
}

// Shutdown is shutdown(): disconnect and stop the threads. Java guards on
// isAlive(), so a client that never started is left alone.
func (c *OpenttdClient) Shutdown() {
	if c.networkClient != nil && c.networkClient.IsRunning() {
		c.networkClient.Shutdown()
	}
}

// GetSend is not in Java: OpenttdClient keeps its NetworkClient private, and
// the only consumer, strategyboard, never asks for it. The tests need the
// speaking thread to drive PACKET_CLIENT_* commands, so the facade hands it out
// here, null until a connection has been made - the same shape as
// OpenttdAdmin.getSend().
func (c *OpenttdClient) GetSend() *client.NetworkClientSender {
	if c.networkClient == nil {
		return nil
	}
	return c.networkClient.GetSend()
}

// GetModel is not in Java either, for the same reason: the state the server's
// packets have been decoded into (client id, map seed, server id, ack frame) is
// what the client tests assert on.
func (c *OpenttdClient) GetModel() *client.NetworkClientModel {
	if c.networkClient == nil {
		return nil
	}
	return c.networkClient.GetModel()
}
