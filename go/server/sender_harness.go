// SenderHarness is the stage-5 stand-in for com.openttd.server.AdminHarness.
//
// The Java harness goes through the OpenttdAdmin facade, which is a later stage
// of this port, so this one wires the two pieces the sender tests need by hand:
// a core.Socket to a FakeAdminServer and a NetworkAdminSender on top of it. It
// reproduces the part of NetworkAdmin.run() those tests depend on, in the same
// order (NetworkAdmin.java): connect, new sender, setGSExecutor, startup,
// join(configuration), then read the server's PROTOCOL packet and hand the
// decoded Protocol to the sender, which is what unlocks updateFrequency() and
// poll().
//
// Replace this with the full AdminHarness once admin/OpenttdAdmin lands.
package server

import "errors"
import "time"
import "com.openttd/gamescript"
import "com.openttd/network/admin"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/util"

var harnessLog = util.GetLogger("SenderHarness")

// PASSWORD and ADMIN_NAME are AdminHarness.PASSWORD and ADMIN_NAME: what the
// harness joins as, and what testJoin reads back off the wire.
const PASSWORD = "test-password"
const ADMIN_NAME = "openttd-admin-test"

type SenderHarness struct {
	server *FakeAdminServer
	configuration *core.Configuration
	socket *core.Socket
	send *admin.NetworkAdminSender
	executor *gamescript.GSExecutor
}

// NewSenderHarness is new AdminHarness().
func NewSenderHarness() (*SenderHarness, error) {
	return NewSenderHarnessWith(PASSWORD, ADMIN_NAME)
}

// NewSenderHarnessWith is AdminHarness(String password, String name).
func NewSenderHarnessWith(password string, name string) (*SenderHarness, error) {
	var server, serverError = NewFakeAdminServer()
	if serverError != nil {
		return nil, serverError
	}
	var configuration = core.NewConfiguration()
	configuration.Password = password
	configuration.Name = name
	configuration.Host = "localhost"
	configuration.AdminPort = server.GetPort()

	var socket, socketError = core.NewTcpSocket(configuration.Host, configuration.AdminPort)
	if socketError != nil {
		server.Close()
		return nil, socketError
	}

	var harness = &SenderHarness{
		server: server,
		configuration: configuration,
		socket: socket,
	}
	var send = admin.NewNetworkAdminSender(socket)
	harness.send = send
	// OpenttdAdmin's constructor calls executor.clear(), which restarts the
	// request numbering: the goal tests assert the first id is "0".
	harness.executor = gamescript.NewGSExecutor(harness)
	harness.executor.Clear()
	send.SetGSExecutor(harness.executor)
	send.Startup()
	send.Join(configuration)

	// The server answers the JOIN with PROTOCOL + WELCOME, exactly as
	// ServerNetworkAdminSocketHandler::SendProtocol() and SendWelcome() do.
	server.Handshake()
	if !harness.awaitProtocol(10) {
		send.Shutdown()
		socket.Close()
		server.Close()
		return nil, errors.New("No PROTOCOL packet from the fake admin server")
	}
	return harness, nil
}

// GetServer is AdminHarness.getServer().
func (h *SenderHarness) GetServer() *FakeAdminServer { return h.server }

// GetConfiguration is AdminHarness.getConfiguration().
func (h *SenderHarness) GetConfiguration() *core.Configuration { return h.configuration }

// GetSend is AdminHarness.getSend(): the sender of the current connection.
func (h *SenderHarness) GetSend() *admin.NetworkAdminSender { return h.send }

// GetGSExecutor is AdminHarness.getGSExecutor().
func (h *SenderHarness) GetGSExecutor() *gamescript.GSExecutor { return h.executor }

// GetSender is OpenttdAdmin.getSend(), as gamescript.SenderProvider sees it: the
// interface view of the current sender, which GSExecutor pulls out on every
// send so a reconnect's new sender is picked up.
func (h *SenderHarness) GetSender() gamescript.Sender {
	if h.send == nil {
		return nil
	}
	return h.send
}

// awaitProtocol is the receive loop of NetworkAdmin.run() narrowed to the one
// packet the sender needs: read until the PROTOCOL packet arrives, decode it
// the way the ADMIN_PACKET_SERVER_PROTOCOL case does, give it to the sender.
// The WELCOME packet that follows is left unread: the game model is a later
// stage, and nothing in these tests looks at it.
func (h *SenderHarness) awaitProtocol(timeoutSeconds int) bool {
	var deadline = time.Now().Add(time.Duration(timeoutSeconds) * 1000 * time.Millisecond)
	for (time.Now().Before(deadline)) {
		var packets, err = h.socket.Receive()
		if err != nil {
			harnessLog.Error(core.Message(err))
			return false
		}
		for i := 0; i < len(packets); i++ {
			if decodeProtocol(h, packets[i]) {
				return true
			}
		}
	}
	return false
}

// decodeProtocol is the ADMIN_PACKET_SERVER_PROTOCOL case of
// NetworkAdmin.handlePacket(), and false for every other packet. The header byte
// is the wire id, so a server packet's ordinal plus SERVER_PACKET_START_VALUE:
// Id() is what compares, exactly as Java's == ADMIN_PACKET_SERVER_PROTOCOL does
// against the enum it built from the header.
func decodeProtocol(h *SenderHarness, packet *core.Packet) bool {
	var packetTypeId = packet.ReadUint8()
	if packetTypeId != constant.ADMIN_PACKET_SERVER_PROTOCOL.Id() {
		return false
	}
	var protocol = admin.New(packet.ReadUint8())
	for (packet.ReadBool8()) {
		protocol.PutProtocol(packet.ReadUint16(), packet.ReadUint16())
	}
	h.send.SetProtocol(protocol)
	return true
}

// Close is AdminHarness.close(): stop the sender, release the connection, stop
// the server. Java's shutdown() sends a QUIT first; the sender tests call
// send.quit() themselves when they care about that packet, so close() only has
// to stop the threads.
func (h *SenderHarness) Close() {
	if h.send != nil {
		h.send.Shutdown()
		h.send.AwaitStopped(5000)
	}
	if h.socket != nil {
		h.socket.Close()
	}
	if h.server != nil {
		h.server.Close()
	}
}
