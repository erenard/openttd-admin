// NetworkUdp mirrors com.openttd.network.client.NetworkUdp: the one-shot
// "is there a server behind this port?" query on the game port (3979). It sends
// PACKET_UDP_CLIENT_FIND_SERVER and waits for PACKET_UDP_SERVER_RESPONSE. See
// udp.h for changes.
//
// Java's main() for manual testing (localhost:3979) has no counterpart here: a
// library package cannot own a main, and the same check lives in
// network_udp_test.go as TestFindServer.
package client

import "sync"
import "time"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/util"

// udpLog is this file's logger. Package client already has two (the sender's
// "log" and the listener's "clientLog"), and a package-level var may not be
// declared twice.
var udpLog = util.GetLogger("NetworkUdp")

// UDP_JOIN_TIMEOUT_MILLIS bounds Java's unbounded updater.join(). A find-server
// query with nothing behind it would otherwise hang the caller forever, which
// is not something a test can wait out; the updater's own socket read is what
// decides when the query is over.
const UDP_JOIN_TIMEOUT_MILLIS = 10000

// JOIN_POLL_MILLIS is how often that wait checks for the thread to finish.
const JOIN_POLL_MILLIS = 10

type NetworkUdp struct {
	socket *core.Socket

	// foundLock guards serverFound and finished, which the updater thread writes
	// while Start() reads them. Java's serverFound field is unsynchronized.
	foundLock   sync.Mutex
	serverFound bool
	finished    bool
}

// NewNetworkUdp is the constructor: the UDP socket for host:port. Java declares
// "throws IOException", which becomes the error this returns.
func NewNetworkUdp(host string, port int) (*NetworkUdp, error) {
	var socket, socketError = core.NewUdpSocket(host, port)
	if socketError != nil {
		return nil, socketError
	}
	return &NetworkUdp{socket: socket, foundLock: sync.Mutex{}}, nil
}

// IsServerFound is isServerFound.
func (u *NetworkUdp) IsServerFound() bool {
	u.foundLock.Lock()
	var found = u.serverFound
	u.foundLock.Unlock()
	return found
}

// Start is start(): updater.start() then updater.join(). The updater runs on
// its own thread; the join waits for it to leave run(), bounded by
// UDP_JOIN_TIMEOUT_MILLIS (see the deviation note on that constant).
func (u *NetworkUdp) Start() {
	go u.update()
	u.awaitFinished(UDP_JOIN_TIMEOUT_MILLIS)
}

// awaitFinished is Thread.join().
func (u *NetworkUdp) awaitFinished(timeoutMillis int) {
	var deadline = time.Now().Add(time.Duration(timeoutMillis) * time.Millisecond)
	for time.Now().Before(deadline) {
		if u.isFinished() {
			return
		}
		time.Sleep(time.Duration(JOIN_POLL_MILLIS) * time.Millisecond)
	}
}

func (u *NetworkUdp) isFinished() bool {
	u.foundLock.Lock()
	var finished = u.finished
	u.foundLock.Unlock()
	return finished
}

// update is Updater.run(): ask, then read until the server answers. Java's
// "break receive" leaves the labelled loop; here the loop condition is
// serverFound, so the inner break and the outer one agree.
func (u *NetworkUdp) update() {
	u.socket.SendEmptyPacket(constant.PACKET_UDP_CLIENT_FIND_SERVER)
	for !u.IsServerFound() {
		var packets, readError = u.socket.Receive()
		if readError != nil {
			// Java's catch (IOException): log and leave run().
			udpLog.Error(core.Message(readError))
			break
		}
		for i := 0; i < len(packets); i++ {
			var packetUdpType = constant.PacketUdpTypeOf(packets[i].ReadUint8())
			udpLog.Debug(constant.PacketUdpTypeName(packetUdpType))
			if packetUdpType == constant.PACKET_UDP_SERVER_RESPONSE {
				u.setServerFound()
				break
			}
		}
	}
	u.setFinished()
}

func (u *NetworkUdp) setServerFound() {
	u.foundLock.Lock()
	u.serverFound = true
	u.foundLock.Unlock()
}

func (u *NetworkUdp) setFinished() {
	u.foundLock.Lock()
	u.finished = true
	u.foundLock.Unlock()
}

// Close releases the socket. Java has no such method: its socket is garbage
// collected with the object. A test needs it to stop leaking the file descriptor.
func (u *NetworkUdp) Close() {
	u.socket.Close()
}
