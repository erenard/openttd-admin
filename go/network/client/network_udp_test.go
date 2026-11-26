// NetworkUdpTest mirrors the check NetworkUdp.main() only made by hand: one
// PACKET_UDP_CLIENT_FIND_SERVER out, PACKET_UDP_SERVER_RESPONSE back, and
// isServerFound() is true. The fake server is the one thing Java had no
// equivalent of, so this is the first test of the UDP protocol at all.
package client_test

import "testing"
import "com.openttd/network/client"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/server"

// TestFindServer is NetworkUdp's main(): ask a server that is there and expect
// it to be found. Start() is Java's start(), which joins the updater thread, so
// by the time it returns the query has been asked and answered.
func TestFindServer(t *testing.T) {
	var srv, srvError = server.NewFakeUdpServer()
	if srvError != nil {
		t.Fatalf("fake UDP server: %s", core.Message(srvError))
	}
	var udp, udpError = client.NewNetworkUdp("localhost", srv.GetPort())
	if udpError != nil {
		t.Fatalf("UDP socket: %s", core.Message(udpError))
	}

	udp.Start()

	if !udp.IsServerFound() {
		t.Errorf("isServerFound: expected true after the response packet")
	}

	// The server side of the exchange: exactly one query, and it is the one
	// udp.cpp's ServerNetworkUDPSocketHandler answers.
	var asked = srv.AwaitPacket(constant.PACKET_UDP_CLIENT_FIND_SERVER, 10)
	if len(asked) != 1 {
		t.Errorf("find-server queries: expected 1, got %d", len(asked))
	}
	if asked[0].GetWireTypeId() != int(constant.PACKET_UDP_CLIENT_FIND_SERVER) {
		t.Errorf("query id: expected %d, got %d", int(constant.PACKET_UDP_CLIENT_FIND_SERVER), asked[0].GetWireTypeId())
	}

	udp.Close()
	srv.Close()
}

// TestFindServerWithoutServer is the other half of Java's main(): pointing at a
// port with nothing behind it must not hang. The updater's read fails at once,
// which ends its run(), so Start() returns with the server not found.
func TestFindServerWithoutServer(t *testing.T) {
	var srv, srvError = server.NewFakeUdpServer()
	if srvError != nil {
		t.Fatalf("fake UDP server: %s", core.Message(srvError))
	}
	var free = srv.GetPort()
	srv.Close()

	var udp, udpError = client.NewNetworkUdp("localhost", free)
	if udpError != nil {
		t.Fatalf("UDP socket: %s", core.Message(udpError))
	}

	// UDP_JOIN_TIMEOUT_MILLIS bounds the join, so this returns instead of
	// hanging the way Java's unbounded join() would.
	udp.Start()

	if udp.IsServerFound() {
		t.Errorf("isServerFound: expected false with nothing listening")
	}
	udp.Close()
}
