// types.go holds the names this package needs from other packages but cannot
// import here. The OpenttdAdmin facade is package admin (go/admin) and the admin
// protocol's listener and sender are package admin too (go/network/admin): this
// dialect has no import aliasing, so one file can only ever spell one of them as
// `admin`. admin_harness.go needs the facade, so the sender type it returns is
// named through the alias below instead.
//
// The same collision happens twice more on the client side: the OpenttdClient
// facade is package client (go/client) and the game-port listener, its sender
// and its model are package client too (go/network/client). client_harness.go
// imports the facade, so the three listener-side types are aliased here.

package server

import "com.openttd/network/admin"
import "com.openttd/network/client"

// NetworkAdminSender is network/admin's NetworkAdminSender: the thread sending
// messages to the server, which AdminHarness.getSend() hands out.
type NetworkAdminSender = admin.NetworkAdminSender

// NetworkClient is network/client's NetworkClient: the listener thread of the
// game port, which the OpenttdClient facade owns.
type NetworkClient = client.NetworkClient

// NetworkClientSender is network/client's NetworkClientSender: the thread
// sending the client's packets, which ClientHarness.GetSend() hands out.
type NetworkClientSender = client.NetworkClientSender

// NetworkClientModel is network/client's NetworkClientModel: the state decoded
// out of the server's packets, which the client rebuilds per connection.
type NetworkClientModel = client.NetworkClientModel
