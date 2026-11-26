// NetworkClientSender mirrors com.openttd.network.client.NetworkClientSender:
// every PACKET_CLIENT_* command a joined client sends, one method each. Payload
// layouts are copied from git show 15.3:src/network/core/tcp_game.h ("Receive_*"
// doc comments).
//
// See network_client.cpp for updates, search "::Send".
//
// Java extends NetworkSender; this port has no inheritance, so it owns a
// core.NetworkSender and delegates startup()/shutdown() to it, passing its own
// isQuitPacket() and debugPacket() as the two hooks the base loop calls, exactly
// as NetworkAdminSender does.
package client

import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/util"

var log = util.GetLogger("NetworkClientSender")

type NetworkClientSender struct {
	// sender is the base class: the queue packets are offered to and the thread
	// that drains it.
	sender *core.NetworkSender
}

// NewNetworkClientSender is NetworkClientSender(Socket socket). The two abstract
// methods of Java's base class are bound to this instance here.
func NewNetworkClientSender(socket *core.Socket) *NetworkClientSender {
	var s = &NetworkClientSender{}
	s.sender = core.NewNetworkSender(socket, quitHook(s), debugHook(s))
	return s
}

// quitHook and debugHook bind the base loop's hooks to this sender, which is
// how the port replaces Java's protected abstract methods.
func quitHook(s *NetworkClientSender) func(packet *core.Packet) bool {
	return func(packet *core.Packet) bool { return s.IsQuitPacket(packet) }
}

func debugHook(s *NetworkClientSender) func(packet *core.Packet) {
	return func(packet *core.Packet) { s.DebugPacket(packet) }
}

// Startup is startup(): running = true and the sender thread on its way.
func (s *NetworkClientSender) Startup() { s.sender.Startup() }

// Shutdown is shutdown(): stop the loop and release the queue's waiting poll.
func (s *NetworkClientSender) Shutdown() { s.sender.Shutdown() }

// IsRunning is the base loop's condition.
func (s *NetworkClientSender) IsRunning() bool { return s.sender.IsRunning() }

// AwaitStopped is the base class's join(timeoutMillis).
func (s *NetworkClientSender) AwaitStopped(timeoutMillis int) { s.sender.AwaitStopped(timeoutMillis) }

// IsDone is the port's stand-in for !thread.isAlive().
func (s *NetworkClientSender) IsDone() bool { return s.sender.IsDone() }

// GetQueue is the queue Java's subclass offers packets to directly; the tests
// use it to see what a command put on the wire.
func (s *NetworkClientSender) GetQueue() *core.PacketQueue { return s.sender.Queue }

// Join is join(Configuration): tell the server we would like to join. Java's
// copy is package-private, which is fine: only NetworkClient calls it, and that
// is in this same package.
//
// The payload is the revision triple OpenTTD uses to decide whether the client
// may join at all: the version string, the NewGRF checksum, then the client's
// own name, the "play as" mask and the retired language field.
func (s *NetworkClientSender) Join(configuration *core.Configuration) {
	var toSend = core.PacketToSend(constant.PACKET_CLIENT_JOIN)
	toSend.WriteString(constant.OPENTTD)
	toSend.WriteUint32(uint32(constant.NEWGRF))
	toSend.WriteString(configuration.Name)
	toSend.WriteUint8(255) // PlayAs
	toSend.WriteUint8(0) // Used to be language
	s.sender.Queue.Offer(toSend)
}

// NewGRFsOk is newGRFsOk(): tell the server we got all the NewGRFs.
func (s *NetworkClientSender) NewGRFsOk() {
	s.sender.Queue.Offer(core.PacketToSend(constant.PACKET_CLIENT_NEWGRFS_CHECKED))
}

// GamePassword is gamePassword(String): the (hashed) game password the server
// asked for.
func (s *NetworkClientSender) GamePassword(password string) {
	var packet = core.PacketToSend(constant.PACKET_CLIENT_GAME_PASSWORD)
	packet.WriteString(password)
	s.sender.Queue.Offer(packet)
}

// CompanyPassword is companyPassword(String): the (hashed) company password.
func (s *NetworkClientSender) CompanyPassword(password string) {
	var packet = core.PacketToSend(constant.PACKET_CLIENT_COMPANY_PASSWORD)
	packet.WriteString(password)
	s.sender.Queue.Offer(packet)
}

// GetMap is getMap(): request the actual map.
func (s *NetworkClientSender) GetMap() {
	s.sender.Queue.Offer(core.PacketToSend(constant.PACKET_CLIENT_GETMAP))
}

// MapOk is mapOk(): tell the server the whole map arrived.
func (s *NetworkClientSender) MapOk() {
	s.sender.Queue.Offer(core.PacketToSend(constant.PACKET_CLIENT_MAP_OK))
}

// Ack is ack(long frame, short token): tell the server which frame this client
// has executed, so it may advance to it.
func (s *NetworkClientSender) Ack(frame int64, token int) {
	var packet = core.PacketToSend(constant.PACKET_CLIENT_ACK)
	packet.WriteUint32(uint32(frame))
	packet.WriteUint8(token)
	s.sender.Queue.Offer(packet)
}

// Chat is chat(NetworkAction, DestType, long dest, String message): uint8
// Action such as NETWORK_ACTION_CHAT_CLIENT (see #NetworkAction), uint8
// Destination type such as DESTTYPE_BROADCAST (see #DestType), uint32 ID of the
// destination such as company or client id, string Message.
//
// The uint64 payload upstream appends for NETWORK_ACTION_GIVE_MONEY is still
// the TODO Java left it as.
func (s *NetworkClientSender) Chat(action constant.NetworkAction, destType constant.DestType, dest int64, message string) {
	var packet = core.PacketToSend(constant.PACKET_CLIENT_CHAT)
	packet.WriteUint8(int(action))
	packet.WriteUint8(int(destType))
	packet.WriteUint32(uint32(dest))
	packet.WriteString(message)
	s.sender.Queue.Offer(packet)
}

// Quit is quit(): the client tells the server it is going away. It is also the
// packet that stops the sender thread, see IsQuitPacket.
func (s *NetworkClientSender) Quit() {
	s.sender.Queue.Offer(core.PacketToSend(constant.PACKET_CLIENT_QUIT))
}

// RCon is rcon(String command, String password): execute a command on the
// server's console. The password comes first on the wire, which is the order
// upstream reads it in.
func (s *NetworkClientSender) RCon(command string, password string) {
	var packet = core.PacketToSend(constant.PACKET_CLIENT_RCON)
	packet.WriteString(password)
	packet.WriteString(command)
	s.sender.Queue.Offer(packet)
}

// ChatClient is chatClient(long clientId, String message): a message addressed
// to one client. A negative id means "nobody in particular", which falls back
// to a broadcast.
func (s *NetworkClientSender) ChatClient(clientId int64, message string) {
	if clientId < 0 {
		s.ChatBroadcast(message)
	} else {
		s.Chat(constant.NETWORK_ACTION_SERVER_MESSAGE, constant.DESTTYPE_CLIENT, clientId, message)
	}
}

// ChatCompany is chatCompany(short companyId, String message): chat to a team.
func (s *NetworkClientSender) ChatCompany(companyId int, message string) {
	s.Chat(constant.NETWORK_ACTION_SERVER_MESSAGE, constant.DESTTYPE_TEAM, int64(companyId), message)
}

// ChatBroadcast is chatBroadcast(String message): chat to all.
func (s *NetworkClientSender) ChatBroadcast(message string) {
	s.Chat(constant.NETWORK_ACTION_SERVER_MESSAGE, constant.DESTTYPE_BROADCAST, int64(0), message)
}

// IsQuitPacket is isQuitPacket(Packet packet): PACKET_CLIENT_QUIT is the last
// packet of a clean shutdown, so the sender thread stops after it.
func (s *NetworkClientSender) IsQuitPacket(packet *core.Packet) bool {
	return packet.TypeId() == int(constant.PACKET_CLIENT_QUIT)
}

// DebugPacket is debugPacket(Packet packet): the base loop calls it only when
// debug logging is on.
func (s *NetworkClientSender) DebugPacket(packet *core.Packet) {
	log.Debug("Snd "+constant.PacketGameTypeName(constant.PacketGameTypeOf(packet.TypeId())))
}
