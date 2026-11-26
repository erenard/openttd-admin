// NetworkAdmin mirrors com.openttd.network.admin.NetworkAdmin: the listening
// thread that owns a NetworkAdminSender, reconnects every 5 s, and decodes
// every ADMIN_PACKET_SERVER_* packet into the NetworkModel and a
// NetworkAdminEvent for the application's Game.
//
// Java extends Thread, so startup() is the port's stand-in for starting the
// thread, AwaitStopped for join(5000), and Wake for interrupt(): Java's
// shutdown() interrupts the reconnect sleep so it does not wait out the 5 s, and
// Go has no interrupt, so the sleep polls a token instead.
//
// Java's game field is the concrete admin/model/Game, which is a later stage of
// this port and would be an import cycle (Game already imports this package), so
// it is the Game interface below: anything with updateAdmin() can be handed over,
// and admin/model/Game will satisfy it when it lands.

package admin

import "strconv"
import "sync"
import "time"

import "com.openttd/gamescript"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/util"

var adminLog = util.GetLogger("NetworkAdmin")

// RECONNECT_DELAY_MILLIS is Java's Thread.sleep(5000) between two connection
// attempts.
const RECONNECT_DELAY_MILLIS = 5000

// SLEEP_POLL_MILLIS is how often the reconnect sleep checks whether shutdown()
// woke it, standing in for Java's interrupt.
const SLEEP_POLL_MILLIS = 10

// Game is the admin/model/Game field: the application's listener, which turns
// the technical NetworkAdminEvent into its own functional events. admin/model
// imports this package, so the port cannot name that class here and this
// interface is the seam instead.
type Game interface {
	UpdateAdmin(event *NetworkAdminEvent, networkModel *NetworkModel, send *NetworkAdminSender)
}

type NetworkAdmin struct {
	configuration *core.Configuration
	gamescript    *gamescript.GSExecutor

	// stateLock guards running, which run() reads on its own thread while
	// shutdown() writes it. Java leaves the plain boolean field unsynchronized.
	stateLock sync.Mutex
	running   bool

	// game is volatile in Java: setGame() comes from the application's thread
	// while run() reads it on the listener thread.
	gameLock sync.Mutex
	game     Game

	// sendLock guards send, which getSend() hands to the application while
	// run() replaces it on every reconnect. Java's field is volatile.
	sendLock sync.Mutex
	send     *NetworkAdminSender

	// networkModel is Java's field, reassigned per connection and read by
	// handlePacket on the listener thread.
	modelLock    sync.Mutex
	networkModel *NetworkModel

	// socket is Java's field: the current connection, closed on shutdown() to
	// unblock a receive() that is waiting for data.
	socketLock sync.Mutex
	socket     *core.Socket
}

// NewNetworkAdmin is the constructor: it keeps the configuration and starts
// listening. The game and the gamescript executor are handed over afterwards,
// as Java's setGame() and setGSExecutor() do.
func NewNetworkAdmin(configuration *core.Configuration) *NetworkAdmin {
	return &NetworkAdmin{
		configuration: configuration,
		stateLock:     sync.Mutex{},
		gameLock:      sync.Mutex{},
		sendLock:      sync.Mutex{},
		modelLock:     sync.Mutex{},
		socketLock:    sync.Mutex{},
	}
}

// startup is start: the listener thread.
func (a *NetworkAdmin) Startup() {
	a.setRunning(true)
	go a.run()
}

// shutdown is shutdown: stop the client, quit and shut the sender down, close
// the socket so a blocked receive() returns, wake the reconnect sleep, then
// wait for the thread.
func (a *NetworkAdmin) Shutdown() {
	a.setRunning(false)
	var send = a.GetSend()
	if send != nil {
		send.Quit()
		send.Shutdown()
	}
	// receive() blocks this thread, so the socket has to be closed here.
	var socket = a.getSocket()
	if socket != nil {
		socket.Close()
	}
	a.awaitStopped(RECONNECT_DELAY_MILLIS)
}

// IsRunning is the loop condition of run().
func (a *NetworkAdmin) IsRunning() bool {
	a.stateLock.Lock()
	var value = a.running
	a.stateLock.Unlock()
	return value
}

func (a *NetworkAdmin) setRunning(running bool) {
	a.stateLock.Lock()
	a.running = running
	a.stateLock.Unlock()
}

// run is run: reconnect until asked to stop. A fresh Socket, NetworkAdminSender
// and NetworkModel per attempt, as in Java, so a reconnect never inherits the
// previous game's model.
func (a *NetworkAdmin) run() {
	adminLog.Debug("started")
	a.setRunning(true)
	for a.IsRunning() {
		a.attempt()
		if !a.IsRunning() {
			break
		}
		if !a.sleep() {
			break
		}
	}
	adminLog.Debug("stopped")
}

// attempt is one turn of Java's while loop: the try, the receive loop and the
// finally block. Every failure Java catches is logged here and swallowed, so
// the reconnect loop carries on, exactly as the catch clauses do.
func (a *NetworkAdmin) attempt() {
	adminLog.Debug("connecting...")
	var socket, err = core.NewTcpSocket(a.configuration.Host, a.configuration.AdminPort)
	if err != nil {
		adminLog.Error(core.Message(err))
		a.disconnect(nil)
		return
	}
	a.setSocket(socket)

	var send = NewNetworkAdminSender(socket)
	a.setSend(send)
	send.SetGSExecutor(a.gamescript)
	send.SetProtocol(nil)
	send.Startup()
	a.setNetworkModel(NewNetworkModel())
	send.Join(a.configuration)

	// Java's inner loop is while (this.running), and handlePacket's NetworkException
	// is caught by the outer catch: it ends the connection, it does not stop
	// the reconnect loop.
	for a.IsRunning() {
		var packets, readError = socket.Receive()
		if readError != nil {
			if a.IsRunning() {
				adminLog.Error(core.Message(readError))
			}
			break
		}
		var finished = false
		for i := 0; i < len(packets); i++ {
			var failure = a.handlePacket(packets[i])
			if failure != nil {
				adminLog.Error(failure.GetExceptionMessage())
				finished = true
				break
			}
		}
		if finished {
			break
		}
	}

	a.disconnect(socket)
}

// disconnect is the finally block: log, tear the sender down, close the socket.
// Java does not send QUIT here, only shutdown() does.
func (a *NetworkAdmin) disconnect(socket *core.Socket) {
	adminLog.Debug("disconneting...")
	a.setSocket(nil)
	var send = a.GetSend()
	if send != nil {
		send.Shutdown()
	}
	if socket != nil {
		socket.Close()
	}
}

// sleep is the Thread.sleep(5000) between attempts. Java interrupts it from
// shutdown(); here the token is checked every SLEEP_POLL_MILLIS ms, so shutdown
// returns promptly instead of waiting out the 5 s.
func (a *NetworkAdmin) sleep() bool {
	var deadline = time.Now().Add(time.Duration(RECONNECT_DELAY_MILLIS) * time.Millisecond)
	for time.Now().Before(deadline) {
		if !a.IsRunning() {
			return false
		}
		time.Sleep(time.Duration(SLEEP_POLL_MILLIS) * time.Millisecond)
	}
	return true
}

// handlePacket is handlePacket: decode one server packet into the NetworkModel
// and a NetworkAdminEvent, then hand both to the game. Java throws NetworkException
// for the packets that end the connection; the port returns it instead, and
// attempt() logs it and reconnects, as Java's catch clause does.
func (a *NetworkAdmin) handlePacket(packet *core.Packet) *NetworkException {
	var packetTypeId = packet.ReadUint8()
	var packetType = constant.PacketServerTypeOf(packetTypeId)
	adminLog.Debug("Rcv " + constant.PacketServerTypeName(packetType))
	var event = NewNetworkAdminEvent(packetType)
	var networkModel = a.getNetworkModel()
	var send = a.GetSend()
	var failure *NetworkException
	switch packetType {
	case constant.ADMIN_PACKET_SERVER_ERROR:
		var errorId = packet.ReadUint8()
		adminLog.Error(DescribeError(errorId))
	case constant.ADMIN_PACKET_SERVER_PROTOCOL:
		var protocol = New(packet.ReadUint8())
		for packet.ReadBool8() {
			protocol.PutProtocol(packet.ReadUint16(), packet.ReadUint16())
		}
		send.SetProtocol(protocol)
		// The admin protocol version belongs to the game model: it is what
		// tells a consumer which packets this connection carries.
		networkModel.GetGameInfo().SetVersion(protocol.GetVersion())
		adminLog.Debug(protocol.ToString())
	case constant.ADMIN_PACKET_SERVER_WELCOME:
		var gameInfo = NewGameInfo()
		// PROTOCOL arrives before WELCOME, and WELCOME replaces the model's
		// GameInfo, so carry the protocol version over.
		gameInfo.SetVersion(networkModel.GetGameInfo().GetVersion())
		gameInfo.SetServerName(packet.ReadString())
		gameInfo.SetServerRevision(packet.ReadString())
		gameInfo.SetServerDedicated(packet.ReadBool8())
		packet.ReadString() // Always ""
		gameInfo.SetMapSeed(int64(packet.ReadUint32()))
		gameInfo.SetMapSet(packet.ReadUint8())
		gameInfo.SetStartDate(int64(packet.ReadUint32()))
		gameInfo.SetMapWidth(packet.ReadUint16())
		gameInfo.SetMapHeight(packet.ReadUint16())
		networkModel.SetGameInfo(gameInfo)
	case constant.ADMIN_PACKET_SERVER_NEWGAME:
		// Java's `break;`: nothing to do. Go's cases do not fall through, so
		// SHUTDOWN below is unaffected by this empty body.
	case constant.ADMIN_PACKET_SERVER_SHUTDOWN:
		// TODO properly shutdown
		failure = NewException(constant.PacketServerTypeName(packetType))
	case constant.ADMIN_PACKET_SERVER_DATE:
		var date = int64(packet.ReadUint32())
		networkModel.GetGameInfo().SetCurrentDate(date)
		event.setDateEvent(packetType, date)
	case constant.ADMIN_PACKET_SERVER_CLIENT_JOIN:
		// This packet arrives after CLIENT_INFO
		event.setClientEvent(packetType, int64(packet.ReadUint32()))
	case constant.ADMIN_PACKET_SERVER_CLIENT_INFO:
		var clientId = int(packet.ReadUint32())
		event.setClientEvent(packetType, int64(clientId))
		var client = networkModel.RetreiveClient(clientId)
		client.SetIp(packet.ReadString())
		client.SetName(packet.ReadString())
		packet.ReadUint8() // Always 0
		client.SetJoinDate(int64(packet.ReadUint32()))
		client.SetCompany(packet.ReadUint8())
	case constant.ADMIN_PACKET_SERVER_CLIENT_UPDATE:
		var updateClientId = int(packet.ReadUint32())
		event.setClientEvent(packetType, int64(updateClientId))
		var updateClient = networkModel.RetreiveClient(updateClientId)
		updateClient.SetName(packet.ReadString())
		updateClient.SetCompany(packet.ReadUint8())
	case constant.ADMIN_PACKET_SERVER_CLIENT_QUIT:
		var quitClientId = int(packet.ReadUint32())
		event.setClientEvent(packetType, int64(quitClientId))
		networkModel.DeleteClient(quitClientId)
	case constant.ADMIN_PACKET_SERVER_CLIENT_ERROR:
		var errorClientId = int(packet.ReadUint32())
		var clientError = packet.ReadUint8()
		event.setClientErrorEvent(packetType, int64(errorClientId), clientError)
		networkModel.DeleteClient(errorClientId)
	case constant.ADMIN_PACKET_SERVER_COMPANY_NEW:
		// This packet arrives after COMPANY_INFO
		event.setCompanyEvent(packetType, packet.ReadUint8())
	case constant.ADMIN_PACKET_SERVER_COMPANY_INFO:
		var infoCompanyId = packet.ReadUint8()
		event.setCompanyEvent(packetType, infoCompanyId)
		var infoCompany = networkModel.RetreiveCompany(infoCompanyId)
		infoCompany.SetName(packet.ReadString())
		infoCompany.SetPresident(packet.ReadString())
		infoCompany.SetColor(packet.ReadUint8())
		infoCompany.SetUsePassword(packet.ReadBool8())
		infoCompany.SetInauguratedYear(int64(packet.ReadUint32()))
		infoCompany.SetAi(packet.ReadBool8())
		infoCompany.SetBankruptcy(packet.ReadUint8())
		infoCompany.SetShareOwners(readShareOwners(packet))
	case constant.ADMIN_PACKET_SERVER_COMPANY_UPDATE:
		var updateCompanyId = packet.ReadUint8()
		event.setCompanyEvent(packetType, updateCompanyId)
		var updateCompany = networkModel.RetreiveCompany(updateCompanyId)
		updateCompany.SetName(packet.ReadString())
		updateCompany.SetPresident(packet.ReadString())
		updateCompany.SetColor(packet.ReadUint8())
		updateCompany.SetUsePassword(packet.ReadBool8())
		updateCompany.SetBankruptcy(packet.ReadUint8())
		updateCompany.SetShareOwners(readShareOwners(packet))
	case constant.ADMIN_PACKET_SERVER_COMPANY_REMOVE:
		var removeCompanyId = packet.ReadUint8()
		var reason = packet.ReadUint8()
		event.setCompanyRemoveEvent(packetType, removeCompanyId, reason)
		networkModel.DeleteCompany(removeCompanyId)
	case constant.ADMIN_PACKET_SERVER_COMPANY_ECONOMY:
		var economyCompanyId = packet.ReadUint8()
		event.setCompanyEvent(packetType, economyCompanyId)
		var economyCompany = networkModel.RetreiveCompany(economyCompanyId)
		economyCompany.SetMoney(packet.ReadInt64())
		economyCompany.SetLoan(packet.ReadUint64())
		economyCompany.SetIncome(packet.ReadInt64())
		economyCompany.SetDeliveredCargo(packet.ReadUint16())
		economyCompany.SetLastValue(packet.ReadUint64())
		economyCompany.SetLastPerformance(packet.ReadUint16())
		economyCompany.SetLastDeliveredCargo(packet.ReadUint16())
		economyCompany.SetPreviousValue(packet.ReadUint64())
		economyCompany.SetPreviousPerformance(packet.ReadUint16())
		economyCompany.SetPreviousDeliveredCargo(packet.ReadUint16())
	case constant.ADMIN_PACKET_SERVER_COMPANY_STATS:
		var statsCompanyId = packet.ReadUint8()
		event.setCompanyEvent(packetType, statsCompanyId)
		var statsCompany = networkModel.RetreiveCompany(statsCompanyId)
		statsCompany.SetVehicules(readCounts(packet, constant.NETWORK_VEHICLE_TYPE_LENGTH))
		statsCompany.SetStations(readCounts(packet, constant.NETWORK_VEHICLE_TYPE_LENGTH))
	case constant.ADMIN_PACKET_SERVER_CHAT:
		// The action and the destination type are decoded by the sender side
		// only; the event keeps the client and the message, as Java does.
		packet.ReadUint8()
		packet.ReadUint8()
		var chatClientId = int64(packet.ReadUint32())
		var chatMessage = packet.ReadString()
		packet.ReadUint64() // Money, only for a 'give money' action
		event.setChatEvent(packetType, chatClientId, chatMessage)
	case constant.ADMIN_PACKET_SERVER_RCON_END:
		event.setRConEndEvent(packetType, packet.ReadString())
	case constant.ADMIN_PACKET_SERVER_RCON:
		event.setRConEvent(packetType, packet.ReadUint16(), packet.ReadString())
	case constant.ADMIN_PACKET_SERVER_CONSOLE:
		var origin = packet.ReadString()
		var consoleMessage = packet.ReadString()
		event.setConsoleEvent(origin, consoleMessage)
	case constant.ADMIN_PACKET_SERVER_GAMESCRIPT:
		event.setGameScriptEvent(packetType, packet.ReadString())
	case constant.ADMIN_PACKET_SERVER_PONG:
		event.setPongEvent(int64(packet.ReadUint32()))
	case constant.ADMIN_PACKET_SERVER_CMD_NAMES:
		for packet.ReadBool8() {
			networkModel.PutCommandName(packet.ReadUint16(), packet.ReadString())
		}
	case constant.ADMIN_PACKET_SERVER_CMD_LOGGING:
		var loggingClientId = int64(packet.ReadUint32())
		var loggingCompanyId = packet.ReadUint8()
		var commandId = packet.ReadUint16()
		// The command buffer has a variable length, the uint32 frame of
		// execution is the last four bytes of the packet.
		var data = packet.ReadBytes(packet.Remaining() - 4)
		var loggingFrame = int64(packet.ReadUint32())
		adminLog.Debug("Command " + strconv.Itoa(commandId) + " of client " + strconv.FormatInt(loggingClientId, 10) + " (company " + strconv.Itoa(loggingCompanyId) + ", frame " + strconv.FormatInt(loggingFrame, 10) + ", " + strconv.Itoa(len(data)) + " bytes of data)")
	case constant.ADMIN_PACKET_SERVER_CMD_LOGGING_OLD:
		var oldClientId = int64(packet.ReadUint32())
		var oldCompanyId = packet.ReadUint8()
		var oldCommandId = packet.ReadUint16()
		var p1 = int64(packet.ReadUint32())
		var p2 = int64(packet.ReadUint32())
		var tile = int64(packet.ReadUint32())
		var text = packet.ReadString()
		var oldFrame = int64(packet.ReadUint32())
		adminLog.Debug("Command " + strconv.Itoa(oldCommandId) + " of client " + strconv.FormatInt(oldClientId, 10) + " (company " + strconv.Itoa(oldCompanyId) + ", p1 " + strconv.FormatInt(p1, 10) + ", p2 " + strconv.FormatInt(p2, 10) + ", tile " + strconv.FormatInt(tile, 10) + ", text \"" + text + "\", frame " + strconv.FormatInt(oldFrame, 10) + ")")
	case constant.ADMIN_PACKET_SERVER_AUTH_REQUEST:
		var method = packet.ReadUint8()
		var parameters = packet.ReadBytes(packet.Remaining())
		adminLog.Error("The server requested secure admin authentication (method " + strconv.Itoa(method) + ", " + strconv.Itoa(len(parameters)) + " bytes), which is not implemented.")
		failure = NewException(constant.PacketServerTypeName(packetType))
	case constant.ADMIN_PACKET_SERVER_ENABLE_ENCRYPTION:
		var nonce = packet.ReadBytes(packet.Remaining())
		adminLog.Error("The server enabled admin encryption (" + strconv.Itoa(len(nonce)) + " bytes of nonce), which is not implemented.")
		failure = NewException(constant.PacketServerTypeName(packetType))
	case constant.ADMIN_PACKET_SERVER_BANNED, constant.ADMIN_PACKET_SERVER_FULL, constant.INVALID_ADMIN_PACKET:
		failure = NewException(constant.PacketServerTypeName(packetType))
	}
	if failure == nil {
		var game = a.getGame()
		if game != nil {
			game.UpdateAdmin(event, networkModel.Copy(), send)
		}
	}
	return failure
}

// GetSend is getSend: the sender of the current connection, which the application
// speaks to. Java's field is volatile; the lock is the Go equivalent.
func (a *NetworkAdmin) GetSend() *NetworkAdminSender {
	a.sendLock.Lock()
	var value = a.send
	a.sendLock.Unlock()
	return value
}

func (a *NetworkAdmin) setSend(send *NetworkAdminSender) {
	a.sendLock.Lock()
	a.send = send
	a.sendLock.Unlock()
}

// getSocket is the socket field: the current connection. shutdown() closes it
// so a blocked receive() returns and the listener thread can leave its loop.
func (a *NetworkAdmin) getSocket() *core.Socket {
	a.socketLock.Lock()
	var value = a.socket
	a.socketLock.Unlock()
	return value
}

func (a *NetworkAdmin) setSocket(socket *core.Socket) {
	a.socketLock.Lock()
	a.socket = socket
	a.socketLock.Unlock()
}

// getNetworkModel is the networkModel field, reassigned per connection.
func (a *NetworkAdmin) getNetworkModel() *NetworkModel {
	a.modelLock.Lock()
	var value = a.networkModel
	a.modelLock.Unlock()
	return value
}

func (a *NetworkAdmin) setNetworkModel(networkModel *NetworkModel) {
	a.modelLock.Lock()
	a.networkModel = networkModel
	a.modelLock.Unlock()
}

// getGame is the game field, read on the listener thread while setGame() writes
// it from the application's.
func (a *NetworkAdmin) getGame() Game {
	a.gameLock.Lock()
	var value = a.game
	a.gameLock.Unlock()
	return value
}

// SetGame is setGame.
func (a *NetworkAdmin) SetGame(game Game) {
	a.gameLock.Lock()
	a.game = game
	a.gameLock.Unlock()
}

// SetGSExecutor is setGSExecutor: attempt() hands the executor to every new
// sender, so goal commands carry a request id their reply can be correlated
// with, and a reconnect keeps using the same executor.
func (a *NetworkAdmin) SetGSExecutor(executor *gamescript.GSExecutor) {
	a.gamescript = executor
}

// awaitStopped is Thread.join(5000): wait for the listener thread to leave its
// reconnect loop.
func (a *NetworkAdmin) awaitStopped(timeoutMillis int) {
	var deadline = time.Now().Add(time.Duration(timeoutMillis) * time.Millisecond)
	for time.Now().Before(deadline) {
		if !a.IsRunning() {
			return
		}
		time.Sleep(time.Duration(SLEEP_POLL_MILLIS) * time.Millisecond)
	}
}

// readShareOwners is the short[4] share owners of a COMPANY_INFO or COMPANY_UPDATE
// packet: four uint8s, as Java's loop reads them.
func readShareOwners(packet *core.Packet) []int {
	var values []int
	for i := 0; i < 4; i++ {
		values = append(values, packet.ReadUint8())
	}
	return values
}

// readCounts is a run of uint16 counts, used for the NetworkVehicleType.length()
// vehicles and stations of a COMPANY_STATS packet.
func readCounts(packet *core.Packet, count int) []int {
	var values []int
	for i := 0; i < count; i++ {
		values = append(values, packet.ReadUint16())
	}
	return values
}
