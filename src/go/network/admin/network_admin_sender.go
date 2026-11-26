// NetworkAdminSender mirrors com.openttd.network.admin.NetworkAdminSender: every
// ADMIN_PACKET_ADMIN_* command the admin port accepts, one method each. Payload
// layouts are copied from git show 15.3:src/network/core/tcp_admin.h ("Receive_ADMIN_*"
// doc comments); the field length limits come from
// git show 15.3:src/network/core/config.h.
//
// Java extends NetworkSender; this port has no inheritance, so it owns a
// core.NetworkSender and delegates startup()/shutdown() to it, passing its own
// isQuitPacket() and debugPacket() as the two hooks the base loop calls.
//
// See network_admin.cpp for updates, search "Receive_ADMIN_".
package admin

import "strconv"
import "com.openttd/gamescript"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/util"

var log = util.GetLogger("NetworkAdminSender")

// LONG_MAX is Java's Long.MAX_VALUE, which the poll helpers use for "every
// client/company": the wire field is uint32, so it reaches OpenTTD as UINT32_MAX.
const LONG_MAX = 9223372036854775807

type NetworkAdminSender struct {
	// sender is the base class: the queue packets are offered to and the thread
	// that drains it.
	sender *core.NetworkSender
	// protocol is what the server announced in its PROTOCOL packet; null until
	// that packet has been decoded, which is what gates updateFrequency().
	protocol *Protocol
	// executor correlates gamescript requests with their replies; null until
	// NetworkAdmin hands one over, which is what gates the goal commands.
	executor *gamescript.GSExecutor
}

// NewNetworkAdminSender is NetworkAdminSender(Socket socket). The two abstract
// methods of Java's base class are bound to this instance here.
func NewNetworkAdminSender(socket *core.Socket) *NetworkAdminSender {
	var s = &NetworkAdminSender{}
	s.sender = core.NewNetworkSender(socket, quitHook(s), debugHook(s))
	return s
}

// quitHook and debugHook bind the base loop's hooks to this sender, which is
// how the port replaces Java's protected abstract methods.
func quitHook(s *NetworkAdminSender) func(packet *core.Packet) bool {
	return func(packet *core.Packet) bool { return s.IsQuitPacket(packet) }
}

func debugHook(s *NetworkAdminSender) func(packet *core.Packet) {
	return func(packet *core.Packet) { s.DebugPacket(packet) }
}

// Startup is startup(): running = true and the sender thread on its way.
func (s *NetworkAdminSender) Startup() { s.sender.Startup() }

// Shutdown is shutdown(): stop the loop and release the queue's waiting poll.
func (s *NetworkAdminSender) Shutdown() { s.sender.Shutdown() }

// IsRunning is the running flag the base loop tests.
func (s *NetworkAdminSender) IsRunning() bool { return s.sender.IsRunning() }

// AwaitStopped is the inherited Thread.join(timeoutMillis): wait for the sender
// thread to stop. Java spells it join(long), next to this class's own
// join(Configuration); the two overloads get separate names here.
func (s *NetworkAdminSender) AwaitStopped(timeoutMillis int) { s.sender.AwaitStopped(timeoutMillis) }

// GetQueue is the protected queue field, for the tests that count what was
// offered but not sent yet.
func (s *NetworkAdminSender) GetQueue() *core.PacketQueue { return s.sender.Queue }

// SetProtocol is setProtocol(Protocol protocol).
func (s *NetworkAdminSender) SetProtocol(protocol *Protocol) { s.protocol = protocol }

// GetProtocol is the protocol field, null before the PROTOCOL packet.
func (s *NetworkAdminSender) GetProtocol() *Protocol { return s.protocol }

// SetGSExecutor is setGSExecutor(GSExecutor gamescript).
func (s *NetworkAdminSender) SetGSExecutor(executor *gamescript.GSExecutor) { s.executor = executor }

// GetGSExecutor is the gamescript field.
func (s *NetworkAdminSender) GetGSExecutor() *gamescript.GSExecutor { return s.executor }

// Rcon is rcon(String command): ADMIN_PACKET_ADMIN_RCON, string command.
// NETWORK_RCONCOMMAND_LENGTH is 500 bytes including '\0': a longer command makes
// the server reject the packet, so callers must keep commands short; the sender
// itself does not cap them.
func (s *NetworkAdminSender) Rcon(command string) {
	var packet = core.PacketToSend(constant.ADMIN_PACKET_ADMIN_RCON)
	packet.WriteString(command)
	s.sender.Queue.Offer(packet)
}

// GameScript is gameScript(String json): ADMIN_PACKET_ADMIN_GAMESCRIPT, string
// json.
//
// @return false if packet was too long
//
// Java's first branch tests json for null; Go has no null string, so the empty
// string stands in for it. The difference is only visible for a caller that
// would send an empty payload, which AdminCmd answers "Malformed json." for
// anyway.
func (s *NetworkAdminSender) GameScript(json string) bool {
	if json == "" {
		return false
	} else if len(json) > core.MTU-3 {
		// Java measures json.length() in UTF-16 code units; Go's len() counts
		// bytes, which is the stricter of the two and the one that matters, since
		// the limit is the packet's byte budget.
		log.Error(json + " is " + strconv.Itoa(len(json)) + " long, max: " + strconv.Itoa(core.MTU-3))
		return false
	}
	var packet = core.PacketToSend(constant.ADMIN_PACKET_ADMIN_GAMESCRIPT)
	packet.WriteString(json)
	s.sender.Queue.Offer(packet)
	return true
}

// Ping is ping(long pingId): ADMIN_PACKET_ADMIN_PING, uint32 identifier.
func (s *NetworkAdminSender) Ping(pingId int64) {
	var packet = core.PacketToSend(constant.ADMIN_PACKET_ADMIN_PING)
	packet.WriteUint32(uint32(pingId))
	s.sender.Queue.Offer(packet)
}

// Join is join(Configuration configuration): ADMIN_PACKET_ADMIN_JOIN, string
// password, string name, string version.
//
// NETWORK_PASSWORD_LENGTH / NETWORK_CLIENT_NAME_LENGTH / NETWORK_REVISION_LENGTH
// are 33, 25 and 33 bytes including the terminator, so a name or password longer
// than that overflows the packet instead of being truncated.
func (s *NetworkAdminSender) Join(configuration *core.Configuration) {
	var toSend = core.PacketToSend(constant.ADMIN_PACKET_ADMIN_JOIN)
	toSend.WriteString(configuration.Password)
	toSend.WriteString(configuration.Name)
	toSend.WriteString(constant.OPENTTD)
	s.sender.Queue.Offer(toSend)
}

// Quit is quit(): ADMIN_PACKET_ADMIN_QUIT, no payload, the notification that
// this admin is leaving.
func (s *NetworkAdminSender) Quit() {
	s.sender.Queue.Offer(core.PacketToSend(constant.ADMIN_PACKET_ADMIN_QUIT))
}

// UpdateFrequency is updateFrequency(AdminUpdateType, AdminUpdateFrequency):
// ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY, uint16 update type, uint16 frequency mask.
//
// @param adminUpdateType Update type
// @param adminUpdateFrequency Update frequency, setting
// ADMIN_FREQUENCY_POLL is always ignored
func (s *NetworkAdminSender) UpdateFrequency(adminUpdateType constant.AdminUpdateType, adminUpdateFrequency constant.AdminUpdateFrequency) {
	if s.protocol == nil {
		log.Error("Protocol not received yet, " + constant.AdminUpdateTypeName(adminUpdateType) + " not sent.")
		return
	}
	if !s.protocol.HasProtocol(adminUpdateType, adminUpdateFrequency) {
		// ADMIN_FREQUENCY_POLL is not advertised for every update type (not
		// for chat, console, command logging or gamescript), and the server
		// closes the connection on a frequency it does not know.
		var advertised = s.protocol.FirstSupported(adminUpdateType)
		if advertised == constant.ADMIN_FREQUENCY_NONE {
			log.Error("The server does not support " + constant.AdminUpdateTypeName(adminUpdateType) + " at all, " + constant.AdminUpdateFrequencyName(adminUpdateFrequency) + " not sent.")
			return
		}
		log.Error("The server does not support " + constant.AdminUpdateFrequencyName(adminUpdateFrequency) + " for " + constant.AdminUpdateTypeName(adminUpdateType) + ", " + constant.AdminUpdateFrequencyName(advertised) + " used instead.")
		adminUpdateFrequency = advertised
	}

	var packet = core.PacketToSend(constant.ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY)
	packet.WriteUint16(int(adminUpdateType))
	packet.WriteUint16(adminUpdateFrequency.Mask())
	s.sender.Queue.Offer(packet)
}

// Poll is poll(AdminUpdateType, long data): ADMIN_PACKET_ADMIN_POLL, uint8
// update type, uint32 id.
//
// @param adminUpdateType the server should answer for, only if
// AdminUpdateFrequency ADMIN_FREQUENCY_POLL is advertised in the PROTOCOL packet.
// @param data ID relevant to the packet type, e.g. - the client ID for
// ADMIN_UPDATE_CLIENT_INFO. Use UINT32_MAX to show all clients. - the company ID
// for ADMIN_UPDATE_COMPANY_INFO. Use UINT32_MAX to show all companies.
//
// Java dereferences the protocol field here without checking it, so polling
// before the PROTOCOL packet throws NullPointerException. The port logs the
// same "does not support" error and sends nothing instead: the caller's thread
// is the admin's own, and an exception there would only be caught and logged by
// NetworkAdmin.run() anyway.
func (s *NetworkAdminSender) Poll(adminUpdateType constant.AdminUpdateType, data int64) {
	if s.protocol == nil || !s.protocol.HasProtocol(adminUpdateType, constant.ADMIN_FREQUENCY_POLL) {
		log.Error("The server does not support " + constant.PacketAdminTypeName(constant.ADMIN_PACKET_ADMIN_POLL) + " for " + constant.AdminUpdateTypeName(adminUpdateType))
		return
	}

	var packet = core.PacketToSend(constant.ADMIN_PACKET_ADMIN_POLL)
	packet.WriteUint8(int(adminUpdateType))
	packet.WriteUint32(uint32(data))
	s.sender.Queue.Offer(packet)
}

// Chat is chat(NetworkAction, DestType, long dest, String message): ADMIN_PACKET_ADMIN_CHAT,
// uint8 action, uint8 desttype, uint32 dest, string message.
//
// @param action Action such as NETWORK_ACTION_CHAT_CLIENT (see NetworkAction).
// @param type Destination type such as DESTTYPE_BROADCAST (see DestType).
// @param dest ID of the destination such as company or client id.
// @param message Message.
func (s *NetworkAdminSender) Chat(action constant.NetworkAction, destType constant.DestType, dest int64, message string) {
	var packet = core.PacketToSend(constant.ADMIN_PACKET_ADMIN_CHAT)
	packet.WriteUint8(int(action))
	packet.WriteUint8(int(destType))
	packet.WriteUint32(uint32(dest))
	packet.WriteString(message)
	s.sender.Queue.Offer(packet)
}

// ExternalChat is externalChat(String source, char colour, String user, String
// message): ADMIN_PACKET_ADMIN_EXTERNAL_CHAT, string source, uint16 colour,
// string user, string message.
//
// @param source Source identifier.
// @param colour Message colour.
// @param user User name.
// @param message Message.
func (s *NetworkAdminSender) ExternalChat(source string, colour int, user string, message string) {
	var packet = core.PacketToSend(constant.ADMIN_PACKET_ADMIN_EXTERNAL_CHAT)
	packet.WriteString(source)
	packet.WriteUint16(colour)
	packet.WriteString(user)
	packet.WriteString(message)
	s.sender.Queue.Offer(packet)
}

// Utility methods

// PollDate is pollDate().
func (s *NetworkAdminSender) PollDate() { s.Poll(constant.ADMIN_UPDATE_DATE, int64(0)) }

// PollCmdNames is pollCmdNames().
func (s *NetworkAdminSender) PollCmdNames() { s.Poll(constant.ADMIN_UPDATE_CMD_NAMES, int64(0)) }

// PollClientInfo is pollClientInfo(long clientId).
func (s *NetworkAdminSender) PollClientInfo(clientId int64) {
	if clientId < 0 {
		clientId = LONG_MAX
	}
	s.Poll(constant.ADMIN_UPDATE_CLIENT_INFO, clientId)
}

// PollCompanyInfo is pollCompanyInfo(short companyId).
func (s *NetworkAdminSender) PollCompanyInfo(companyId int) {
	var longId = int64(companyId)
	if longId < 0 {
		longId = LONG_MAX
	}
	s.Poll(constant.ADMIN_UPDATE_COMPANY_INFO, longId)
}

// PollCompanyEconomy is pollCompanyEconomy(short companyId).
func (s *NetworkAdminSender) PollCompanyEconomy(companyId int) {
	var longId = int64(companyId)
	if longId < 0 {
		longId = LONG_MAX
	}
	s.Poll(constant.ADMIN_UPDATE_COMPANY_ECONOMY, longId)
}

// PollCompanyStats is pollCompanyStats(short companyId).
func (s *NetworkAdminSender) PollCompanyStats(companyId int) {
	var longId = int64(companyId)
	if longId < 0 {
		longId = LONG_MAX
	}
	s.Poll(constant.ADMIN_UPDATE_COMPANY_STATS, longId)
}

// ChatClient is chatClient(long clientId, String message).
func (s *NetworkAdminSender) ChatClient(clientId int64, message string) {
	if clientId < 0 {
		s.ChatBroadcast(message)
	} else {
		s.Chat(constant.NETWORK_ACTION_SERVER_MESSAGE, constant.DESTTYPE_CLIENT, clientId, message)
	}
}

// ChatCompany is chatCompany(short companyId, String message): chat to a team.
//
// @param companyId
// @param message
func (s *NetworkAdminSender) ChatCompany(companyId int, message string) {
	s.Chat(constant.NETWORK_ACTION_SERVER_MESSAGE, constant.DESTTYPE_TEAM, int64(companyId), message)
}

// ChatBroadcast is chatBroadcast(String message): chat to all.
//
// @param message
func (s *NetworkAdminSender) ChatBroadcast(message string) {
	s.Chat(constant.NETWORK_ACTION_SERVER_MESSAGE, constant.DESTTYPE_BROADCAST, int64(0), message)
}

// AddCompanyGoal is addCompanyGoal(short companyId, String text, GoalType, int
// destinationId): the gamescript's addGoal command.
//
// @param companyId
// @param text Text displayed for this goal
// @param goalType Type of the goal's destination
// @param destinationId Id of the goal's destination
func (s *NetworkAdminSender) AddCompanyGoal(companyId int, text string, goalType constant.GoalType, destinationId int) {
	var arguments = []string{
		strconv.Itoa(companyId),
		gamescript.JsonString(text),
		strconv.Itoa(int(goalType)),
		strconv.Itoa(destinationId),
	}
	s.sendGameScriptCommand(constant.GS_COMMAND_ADD_GOAL, arguments)
}

// AddGlobalGoal is addGlobalGoal(String text, GoalType goalType, int
// destinationId): a global goal is a goal of company -1.
//
// @param text Text displayed for this goal
// @param goalType Type of the goal's destination
// @param destinationId Id of the goal's destination
func (s *NetworkAdminSender) AddGlobalGoal(text string, goalType constant.GoalType, destinationId int) {
	s.AddCompanyGoal(-1, text, goalType, destinationId)
}

// RemoveGoal is removeGoal(int goalId).
//
// @param goalId
func (s *NetworkAdminSender) RemoveGoal(goalId int) {
	s.sendGameScriptCommand(constant.GS_COMMAND_REMOVE_GOAL, []string{strconv.Itoa(goalId)})
}

// RemoveAllGoal is removeAllGoal(): clear the goal list.
func (s *NetworkAdminSender) RemoveAllGoal() {
	var arguments []string
	s.sendGameScriptCommand(constant.GS_COMMAND_REMOVE_ALL_GOAL, arguments)
}

// SetTownCargoGoal is setTownCargoGoal(int townId, TownEffect townEffect, int
// goal): a cargo goal for a town.
//
// @param townId
// @param townEffect
// @param goal
func (s *NetworkAdminSender) SetTownCargoGoal(townId int, townEffect constant.TownEffect, goal int) {
	var arguments = []string{
		strconv.Itoa(townId),
		strconv.Itoa(int(townEffect)),
		strconv.Itoa(goal),
	}
	s.sendGameScriptCommand(constant.GS_COMMAND_SET_TOWN_CARGO_GOAL, arguments)
}

// SendGameScriptCommand is sendGameScriptCommand(GSCommand, JsonArray): a gamescript
// command with a request id, so that GSExecutor can correlate the reply: AdminCmd
// requires "i", "c" and "a".
//
// @param command
// @param arguments
func (s *NetworkAdminSender) sendGameScriptCommand(command constant.GSCommand, arguments []string) {
	if s.executor == nil {
		log.Error("No gamescript executor, " + constant.GSCommandName(command) + " not sent.")
		return
	}
	s.executor.Send(gamescript.NewGSJsonRequest(command, arguments))
}

// IsQuitPacket is isQuitPacket(Packet packet): ADMIN_PACKET_ADMIN_QUIT is the
// last packet of a clean shutdown, so the sender thread stops after it.
func (s *NetworkAdminSender) IsQuitPacket(packet *core.Packet) bool {
	return packet.TypeId() == int(constant.ADMIN_PACKET_ADMIN_QUIT)
}

// DebugPacket is debugPacket(Packet packet): one line per sent packet, only
// asked for when debug logging is on.
func (s *NetworkAdminSender) DebugPacket(packet *core.Packet) {
	log.Debug("Snd " + constant.PacketAdminTypeName(constant.PacketAdminTypeOf(packet.TypeId())))
}
