// The decoding half of test/java/com/openttd/network/admin/NetworkAdminServerPacketsTest.java:
// every ADMIN_PACKET_SERVER_* packet the admin network can receive, put on the
// wire by server.FakeAdminServer and read back by NetworkAdmin.handlePacket.
//
// The Java test asserts through AdminHarness.getGame() and EventRecorder, i.e.
// through the functional events admin/model/Game raises, which is a later stage
// of this port. Here the Game seam is filled by the Recorder below, so these
// cases check the same wire decoding and the same NetworkModel state; the
// event-level assertions come with the port of Game.

package admin_test

import "testing"
import "sync"
import "com.openttd"
import "com.openttd/network/admin"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/server"

// Recorder is the Game seam: it keeps every NetworkAdminEvent and the model
// copy NetworkAdmin handed over with it, in arrival order.
type Recorder struct {
	lock sync.Mutex
	events []*admin.NetworkAdminEvent
	models []*admin.NetworkModel
}

func (r *Recorder) UpdateAdmin(event *admin.NetworkAdminEvent, networkModel *admin.NetworkModel, send *admin.NetworkAdminSender) {
	r.lock.Lock()
	r.events = append(r.events, event)
	r.models = append(r.models, networkModel)
	r.lock.Unlock()
}

// EventsOf is how often a packet type arrived, which is what the Java test's
// recorder.getXEvents().size() checks.
func (r *Recorder) EventsOf(packetType constant.PacketServerType) int {
	r.lock.Lock()
	var count = 0
	for i := 0; i < len(r.events); i++ {
		if r.events[i].GetPacketServerType() == packetType {
			count = count + 1
		}
	}
	r.lock.Unlock()
	return count
}

// LastEvent is the most recent event of a packet type, for checking the fields
// the decoder filled in.
func (r *Recorder) LastEvent(packetType constant.PacketServerType) *admin.NetworkAdminEvent {
	r.lock.Lock()
	var latest *admin.NetworkAdminEvent
	for i := 0; i < len(r.events); i++ {
		if r.events[i].GetPacketServerType() == packetType {
			latest = r.events[i]
		}
	}
	r.lock.Unlock()
	return latest
}

// LastModel is the most recent model copy, i.e. the state after the last packet.
func (r *Recorder) LastModel() *admin.NetworkModel {
	r.lock.Lock()
	var model *admin.NetworkModel
	if len(r.models) > 0 {
		model = r.models[len(r.models)-1]
	}
	r.lock.Unlock()
	return model
}

func (r *Recorder) Count() int {
	r.lock.Lock()
	var count = len(r.events)
	r.lock.Unlock()
	return count
}

// Admin is one connection under test: the fake server, the configuration that
// points at it, the listening NetworkAdmin and the Recorder that sees what it
// decoded. setUp() in the Java test builds the same thing through AdminHarness.
type Admin struct {
	server *server.FakeAdminServer
	listener *admin.NetworkAdmin
	recorder *Recorder
}

// newAdmin is setUp(): a fresh connection per test. The listener sends the JOIN
// itself, so the test waits for that connection before the server answers it.
func newAdmin(t *testing.T) *Admin {
	var fake, serverError = server.NewFakeAdminServer()
	if serverError != nil {
		t.Errorf("FakeAdminServer: %s", core.Message(serverError))
		return nil
	}

	var configuration = core.NewConfiguration()
	configuration.Host = "localhost"
	configuration.AdminPort = fake.GetPort()
	configuration.Password = server.PASSWORD
	configuration.Name = server.ADMIN_NAME

	var recorder = &Recorder{lock: sync.Mutex{}}
	var listener = admin.NewNetworkAdmin(configuration)
	listener.SetGSExecutor(nil)
	listener.SetGame(recorder)
	listener.Startup()

	var self = &Admin{server: fake, listener: listener, recorder: recorder}
	fake.AwaitConnectionCount(1, 10)
	fake.Handshake()
	return self
}

// close is tearDown(): stop listening and shut the fake server down.
func (a *Admin) close() {
	a.listener.Shutdown()
	a.server.Close()
}

// send is the test's send(PacketWriter packet) helper.
func (a *Admin) send(packet *server.PacketWriter) {
	a.server.Send(packet)
}

// awaitEvent waits for one arrival of a packet type, as TestUtils.awaitSize does.
func (a *Admin) awaitEvent(packetType constant.PacketServerType, timeoutSeconds int, description string) {
	openttd.Await(func() bool { return a.recorder.EventsOf(packetType) > 0 }, timeoutSeconds, description)
}

// assertStillConnected is the Java helper of the same name: the connection
// survived, so a PING round trips. The recorder sees the PONG, not a PongEvent.
func (a *Admin) assertStillConnected(t *testing.T) {
	var send = a.listener.GetSend()
	if send == nil {
		t.Errorf("no sender on the current connection")
		return
	}
	send.Ping(int64(9))
	openttd.Await(func() bool { return a.recorder.EventsOf(constant.ADMIN_PACKET_SERVER_PONG) > 0 }, 5, "a PONG")
}

// ADMIN_PACKET_SERVER_WELCOME: string server name, string revision, bool
// dedicated, string (used to be map name), uint32 seed, uint8 landscape, uint32
// start date, uint16 width, uint16 height.
func TestWelcome(t *testing.T) {
	var a = newAdmin(t)
	if a == nil {
		return
	}
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_WELCOME, 10, "a WELCOME")
	var gameInfo = a.recorder.LastModel().GetGameInfo()
	if gameInfo.GetVersion() != server.GAME_ADMIN_VERSION {
		t.Errorf("protocol version: expected %d, got %d", server.GAME_ADMIN_VERSION, gameInfo.GetVersion())
	}
	if gameInfo.GetServerName() != "Strategyboard test server" {
		t.Errorf("server name: %s", gameInfo.GetServerName())
	}
	if gameInfo.GetServerRevision() != "15.3" {
		t.Errorf("server revision: %s", gameInfo.GetServerRevision())
	}
	if !gameInfo.IsServerDedicated() {
		t.Errorf("expected a dedicated server")
	}
	if gameInfo.GetMapSeed() != int64(123456789) {
		t.Errorf("map seed: expected 123456789, got %d", gameInfo.GetMapSeed())
	}
	if gameInfo.GetMapSet() != 1 {
		t.Errorf("map set: expected 1, got %d", gameInfo.GetMapSet())
	}
	if gameInfo.GetStartDate() != int64(1950*365+1) {
		t.Errorf("start date: expected %d, got %d", 1950*365+1, gameInfo.GetStartDate())
	}
	if gameInfo.GetMapWidth() != 64 || gameInfo.GetMapHeight() != 64 {
		t.Errorf("map size: expected 64x64, got %dx%d", gameInfo.GetMapWidth(), gameInfo.GetMapHeight())
	}
	// PROTOCOL comes before WELCOME, and WELCOME replaces the model's GameInfo,
	// so the version must survive the replacement.
	a.assertStillConnected(t)
	a.close()
}

// ADMIN_PACKET_SERVER_DATE: uint32 current game date.
func TestDate(t *testing.T) {
	var a = newAdmin(t)
	if a == nil {
		return
	}
	var date = int64(1950*365+100)
	a.send(server.New(constant.ADMIN_PACKET_SERVER_DATE).WriteUint32(uint32(date)))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_DATE, 5, "a date packet")
	var gameInfo = a.recorder.LastModel().GetGameInfo()
	if gameInfo.GetCurrentDate() != date {
		t.Errorf("current date: expected %d, got %d", date, gameInfo.GetCurrentDate())
	}
	a.assertStillConnected(t)
	a.close()
}

// ADMIN_PACKET_SERVER_CLIENT_JOIN / _INFO / _UPDATE / _QUIT.
func TestClientPackets(t *testing.T) {
	var a = newAdmin(t)
	if a == nil {
		return
	}
	// CLIENT_INFO: uint32 id, string ip, string name, uint8 language, uint32
	// join date, uint8 company.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_CLIENT_INFO).
		WriteUint32(uint32(7)).
		WriteString("127.0.0.1").
		WriteString("Test client").
		WriteUint8(0).
		WriteUint32(uint32(1950*365+1)).
		WriteUint8(255))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_CLIENT_INFO, 5, "a client info packet")
	var clients = a.recorder.LastModel().RetreiveClients()
	if len(clients) != 1 {
		t.Errorf("expected one client, got %d", len(clients))
	}
	if clients[0].GetName() != "Test client" || clients[0].GetIp() != "127.0.0.1" {
		t.Errorf("client: %s", clients[0].ToString())
	}
	if clients[0].GetJoinDate() != int64(1950*365+1) || clients[0].GetCompany() != 255 {
		t.Errorf("client join date or company: %s", clients[0].ToString())
	}

	// CLIENT_JOIN: uint32 id, and it arrives after CLIENT_INFO.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_CLIENT_JOIN).WriteUint32(uint32(7)))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_CLIENT_JOIN, 5, "a client join packet")

	// CLIENT_UPDATE: uint32 id, string name, uint8 company.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_CLIENT_UPDATE).
		WriteUint32(uint32(7)).
		WriteString("Renamed client").
		WriteUint8(1))
	openttd.Await(func() bool {
		var list = a.recorder.LastModel().RetreiveClients()
		if len(list) == 0 {
			return false
		}
		return list[0].GetName() == "Renamed client"
	}, 5, "a client update packet")
	if a.recorder.LastModel().RetreiveClients()[0].GetCompany() != 1 {
		t.Errorf("expected the client to play as company 1")
	}

	// CLIENT_QUIT: uint32 id.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_CLIENT_QUIT).WriteUint32(uint32(7)))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_CLIENT_QUIT, 5, "a client quit packet")
	if len(a.recorder.LastModel().RetreiveClients()) != 0 {
		t.Errorf("expected the client to be gone from the model")
	}
	a.assertStillConnected(t)
	a.close()
}
// ADMIN_PACKET_SERVER_COMPANY_NEW / _INFO / _UPDATE / _ECONOMY / _STATS /
// _REMOVE, the layouts of the six company packets.
func TestCompanyPackets(t *testing.T) {
	var a = newAdmin(t)
	if a == nil {
		return
	}
	// COMPANY_INFO: uint8 id, string name, string president, uint8 colour, bool
	// password, uint32 inaugurated year, bool ai, uint8 bankruptcy, 4x uint8
	// share owners.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_COMPANY_INFO).
		WriteUint8(1).
		WriteString("Test company").
		WriteString("Test president").
		WriteUint8(2).
		WriteBool8(true).
		WriteUint32(uint32(1980)).
		WriteBool8(false).
		WriteUint8(3).
		WriteUint8(7).WriteUint8(8).WriteUint8(9).WriteUint8(10))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_COMPANY_INFO, 5, "a company info packet")
	var companies = a.recorder.LastModel().RetreiveCompanies()
	if len(companies) != 1 {
		t.Errorf("expected one company, got %d", len(companies))
	}
	var company = companies[0]
	if company.GetName() != "Test company" || company.GetPresident() != "Test president" {
		t.Errorf("company: %s", company.ToString())
	}
	if company.GetColor() != 2 || !company.IsUsePassword() || company.IsAi() {
		t.Errorf("company colour, password or ai: %s", company.ToString())
	}
	if company.GetInauguratedYear() != int64(1980) || company.GetBankruptcy() != 3 {
		t.Errorf("company year or bankruptcy: %s", company.ToString())
	}
	var owners = company.GetShareOwners()
	if len(owners) != 4 || owners[0] != 7 || owners[3] != 10 {
		t.Errorf("expected four share owners 7,8,9,10, got %d,%d,%d,%d", owners[0], owners[1], owners[2], owners[3])
	}

	// COMPANY_NEW: uint8 id, and it arrives after COMPANY_INFO.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_COMPANY_NEW).WriteUint8(1))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_COMPANY_NEW, 5, "a company new packet")

	// COMPANY_ECONOMY: int64 money, uint64 loan, int64 income, uint16 delivered,
	// then the last and previous quarter's value, performance and delivery.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_COMPANY_ECONOMY).
		WriteUint8(1).
		WriteInt64(int64(1234567890)).
		WriteUint64(uint64(9876543210)).
		WriteInt64(int64(-1)).
		WriteUint16(2).
		WriteUint64(uint64(3)).WriteUint16(4).WriteUint16(5).
		WriteUint64(uint64(6)).WriteUint16(7).WriteUint16(8))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_COMPANY_ECONOMY, 5, "a company economy packet")
	var economy = a.recorder.LastModel().RetreiveCompanies()[0]
	if economy.GetMoney() != int64(1234567890) || economy.GetIncome() != int64(-1) {
		t.Errorf("company money or income: %d, %d", economy.GetMoney(), economy.GetIncome())
	}
	if economy.GetLoan() != uint64(9876543210) {
		t.Errorf("company loan: expected 9876543210, got %d", economy.GetLoan())
	}
	if economy.GetDeliveredCargo() != 2 || economy.GetLastValue() != uint64(3) || economy.GetPreviousValue() != uint64(6) {
		t.Errorf("company economy quarters: %s", economy.ToString())
	}

	// COMPANY_STATS: NetworkVehicleType.length() vehicle counts, then the same
	// number of station counts.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_COMPANY_STATS).
		WriteUint8(1).
		WriteUint16(11).WriteUint16(12).WriteUint16(13).WriteUint16(14).WriteUint16(15).
		WriteUint16(21).WriteUint16(22).WriteUint16(23).WriteUint16(24).WriteUint16(25))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_COMPANY_STATS, 5, "a company stats packet")
	var stats = a.recorder.LastModel().RetreiveCompanies()[0]
	if len(stats.GetVehicules()) != constant.NETWORK_VEHICLE_TYPE_LENGTH || stats.GetVehicules()[0] != 11 {
		t.Errorf("expected %d vehicle counts, got %s", constant.NETWORK_VEHICLE_TYPE_LENGTH, stats.ToString())
	}
	if len(stats.GetStations()) != constant.NETWORK_VEHICLE_TYPE_LENGTH || stats.GetStations()[4] != 25 {
		t.Errorf("expected %d station counts, got %s", constant.NETWORK_VEHICLE_TYPE_LENGTH, stats.ToString())
	}

	// COMPANY_UPDATE: uint8 id, string name, string president, uint8 colour,
	// bool password, uint8 bankruptcy, 4x uint8 share owners.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_COMPANY_UPDATE).
		WriteUint8(1).
		WriteString("Renamed company").
		WriteString("Renamed president").
		WriteUint8(5).
		WriteBool8(false).
		WriteUint8(0).
		WriteUint8(1).WriteUint8(1).WriteUint8(1).WriteUint8(1))
	openttd.Await(func() bool {
		var list = a.recorder.LastModel().RetreiveCompanies()
		if len(list) == 0 {
			return false
		}
		return list[0].GetName() == "Renamed company"
	}, 5, "a company update packet")

	// COMPANY_REMOVE: uint8 id, uint8 reason.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_COMPANY_REMOVE).
		WriteUint8(1).
		WriteUint8(int(constant.ADMIN_CRR_AUTOCLEAN)))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_COMPANY_REMOVE, 5, "a company remove packet")
	if len(a.recorder.LastModel().RetreiveCompanies()) != 0 {
		t.Errorf("expected the company to be gone from the model")
	}
	a.assertStillConnected(t)
	a.close()
}

// ADMIN_PACKET_SERVER_CHAT, _CONSOLE, _GAMESCRIPT, _RCON and _RCON_END: the
// text carrying packets.
func TestTextPackets(t *testing.T) {
	var a = newAdmin(t)
	if a == nil {
		return
	}
	// CHAT: uint8 action, uint8 destination type, uint32 client id, string
	// message, uint64 money.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_CHAT).
		WriteUint8(int(constant.NETWORK_ACTION_CHAT)).
		WriteUint8(int(constant.DESTTYPE_BROADCAST)).
		WriteUint32(uint32(7)).
		WriteString("hello").
		WriteUint64(uint64(0)))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_CHAT, 5, "a chat packet")
	var chat = a.recorder.LastEvent(constant.ADMIN_PACKET_SERVER_CHAT)
	if chat.GetClientId() != int64(7) || chat.GetMessage() != "hello" {
		t.Errorf("chat event: %s", chat.ToString())
	}

	// CONSOLE: string origin, string message. Java's setConsoleEvent does not
	 // record the packet type, so the event keeps the origin and the message.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_CONSOLE).
		WriteString("console").
		WriteString("a console line"))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_CONSOLE, 5, "a console packet")
	var console = a.recorder.LastEvent(constant.ADMIN_PACKET_SERVER_CONSOLE)
	if console.GetOrigin() != "console" || console.GetMessage() != "a console line" {
		t.Errorf("console event: %s", console.ToString())
	}

	// GAMESCRIPT: string json.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_GAMESCRIPT).WriteString("{\"id\":\"1\",\"data\":\"42\"}"))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_GAMESCRIPT, 5, "a gamescript packet")

	// RCON: uint16 colour, string message.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_RCON).WriteUint16(1).WriteString("rcon output"))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_RCON, 5, "an rcon packet")

	// RCON_END: string the command as requested.
	a.send(server.New(constant.ADMIN_PACKET_SERVER_RCON_END).WriteString("set_time 1"))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_RCON_END, 5, "an rcon end packet")
	a.assertStillConnected(t)
	a.close()
}

// ADMIN_PACKET_SERVER_CMD_NAMES: repeats of bool data to follow, uint16 id,
// string name, until the packet is full.
func TestCmdNames(t *testing.T) {
	var a = newAdmin(t)
	if a == nil {
		return
	}
	var packet = server.New(constant.ADMIN_PACKET_SERVER_CMD_NAMES)
	packet.WriteBool8(true).WriteUint16(4).WriteString("AddBridge")
	packet.WriteBool8(true).WriteUint16(5).WriteString("AddLevelCrossing")
	packet.WriteBool8(false)
	a.send(packet)
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_CMD_NAMES, 5, "a cmd names packet")
	var names = a.recorder.LastModel().RetreiveCommandNames()
	var first, hasFirst = names[4]
	var second, hasSecond = names[5]
	if !hasFirst || first != "AddBridge" || !hasSecond || second != "AddLevelCrossing" {
		t.Errorf("expected two command names, got %d", len(names))
	}
	a.assertStillConnected(t)
	a.close()
}

// ADMIN_PACKET_SERVER_COMPANY_ECONOMY with values above 2^32: the uint64 fields
// must not lose precision, which is what the Java test pins with BigInteger.
func TestEconomyUint64Boundary(t *testing.T) {
	var a = newAdmin(t)
	if a == nil {
		return
	}
	a.send(server.New(constant.ADMIN_PACKET_SERVER_COMPANY_ECONOMY).
		WriteUint8(4).
		WriteInt64(int64(-123456)).
		WriteUint64(uint64(50000000000)).
		WriteInt64(int64(987654321)).
		WriteUint16(1200).
		WriteUint64(uint64(200000000000)).WriteUint16(10000).WriteUint16(500).
		WriteUint64(uint64(1000000)).WriteUint16(9000).WriteUint16(400))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_COMPANY_ECONOMY, 5, "a company economy packet")
	var company = a.recorder.LastModel().RetreiveCompany(4)
	if company.GetMoney() != int64(-123456) || company.GetIncome() != int64(987654321) {
		t.Errorf("company money or income: %d, %d", company.GetMoney(), company.GetIncome())
	}
	if company.GetLoan() != uint64(50000000000) {
		t.Errorf("company loan: expected 50000000000, got %d", company.GetLoan())
	}
	if company.GetLastValue() != uint64(200000000000) || company.GetPreviousValue() != uint64(1000000) {
		t.Errorf("company quarter values: %d, %d", company.GetLastValue(), company.GetPreviousValue())
	}
	if company.GetDeliveredCargo() != 1200 || company.GetLastPerformance() != 10000 || company.GetLastDeliveredCargo() != 500 {
		t.Errorf("company economy: %s", company.ToString())
	}
	a.assertStillConnected(t)
	a.close()
}

// ADMIN_PACKET_SERVER_PONG: uint32 id echoed from the ping, at the top of the
// uint32 range.
func TestPong(t *testing.T) {
	var a = newAdmin(t)
	if a == nil {
		return
	}
	var send = a.listener.GetSend()
	if send == nil {
		t.Errorf("no sender on the current connection")
		return
	}
	send.Ping(int64(4294967295))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_PONG, 5, "a pong packet")
	var pong = a.recorder.LastEvent(constant.ADMIN_PACKET_SERVER_PONG)
	if pong.GetPingId() != int64(4294967295) {
		t.Errorf("pong id: expected 4294967295, got %d", pong.GetPingId())
	}
	a.close()
}

// ADMIN_PACKET_SERVER_CMD_LOGGING (id 127): uint32 client, uint8 company, uint16
// command, the variable length command buffer, uint32 frame. Logging only: the
// model must not change and the session must survive.
func TestCmdLogging(t *testing.T) {
	var a = newAdmin(t)
	if a == nil {
		return
	}
	var commandBuffer = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}
	a.send(server.New(constant.ADMIN_PACKET_SERVER_CMD_LOGGING).
		WriteUint32(uint32(2)).WriteUint8(1).WriteUint16(37).
		WriteBytes(commandBuffer).WriteUint32(uint32(12345)))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_CMD_LOGGING, 5, "a cmd logging packet")
	if len(a.recorder.LastModel().RetreiveCompanies()) != 0 {
		t.Errorf("command logging must not touch the companies")
	}
	a.assertStillConnected(t)
	a.close()
}

// ADMIN_PACKET_SERVER_CMD_LOGGING_OLD (id 123): the admin protocol version 1
// layout, uint32 client, uint8 company, uint16 command, uint32 p1, uint32 p2,
// uint32 tile, string text, uint32 frame.
func TestCmdLoggingOld(t *testing.T) {
	var a = newAdmin(t)
	if a == nil {
		return
	}
	a.send(server.New(constant.ADMIN_PACKET_SERVER_CMD_LOGGING_OLD).
		WriteUint32(uint32(3)).WriteUint8(1).WriteUint16(42).
		WriteUint32(uint32(1000)).WriteUint32(uint32(2)).WriteUint32(uint32(30000)).
		WriteString("Build station").WriteUint32(uint32(999)))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_CMD_LOGGING_OLD, 5, "a cmd logging (old) packet")
	a.assertStillConnected(t)
	a.close()
}

// ADMIN_PACKET_SERVER_NEWGAME carries no payload and no model change.
func TestNewGame(t *testing.T) {
	var a = newAdmin(t)
	if a == nil {
		return
	}
	a.send(server.New(constant.ADMIN_PACKET_SERVER_NEWGAME))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_NEWGAME, 5, "a new game packet")
	if len(a.recorder.LastModel().RetreiveClients()) != 0 || len(a.recorder.LastModel().RetreiveCompanies()) != 0 {
		t.Errorf("NEWGAME must not touch the model")
	}
	a.assertStillConnected(t)
	a.close()
}

// ADMIN_PACKET_SERVER_ERROR: uint8 NetworkErrorCode. Reported, session survives.
func TestErrorPacket(t *testing.T) {
	var a = newAdmin(t)
	if a == nil {
		return
	}
	a.send(server.New(constant.ADMIN_PACKET_SERVER_ERROR).WriteUint8(int(constant.NETWORK_ERROR_WRONG_PASSWORD)))
	a.awaitEvent(constant.ADMIN_PACKET_SERVER_ERROR, 5, "an error packet")
	a.assertStillConnected(t)
	a.close()
}

// The packets that end the session: NetworkAdmin throws, the connection is torn
// down, and the reconnect loop starts a new one five seconds later.
func TestSessionEndingPackets(t *testing.T) {
	var packetTypes = []constant.PacketServerType{
		constant.ADMIN_PACKET_SERVER_FULL,
		constant.ADMIN_PACKET_SERVER_BANNED,
		constant.ADMIN_PACKET_SERVER_AUTH_REQUEST,
		constant.ADMIN_PACKET_SERVER_ENABLE_ENCRYPTION,
	}
	for i := 0; i < len(packetTypes); i++ {
		var a = newAdmin(t)
		if a == nil {
			return
		}
		var packet = server.New(packetTypes[i])
		if packetTypes[i] == constant.ADMIN_PACKET_SERVER_AUTH_REQUEST {
			packet.WriteUint8(1).WriteBytes(newBytes(32)).WriteBytes(newBytes(24))
		} else if packetTypes[i] == constant.ADMIN_PACKET_SERVER_ENABLE_ENCRYPTION {
			packet.WriteBytes(newBytes(24))
		}
		a.send(packet)
		a.server.AwaitClientDisconnect(10)
		if !a.listener.IsRunning() {
			t.Errorf("expected the listener to survive %s", constant.PacketServerTypeName(packetTypes[i]))
		}
		a.close()
	}
}

// A packet id the protocol does not define must be refused, not read as a payload.
func TestUnknownPacketId(t *testing.T) {
	var a = newAdmin(t)
	if a == nil {
		return
	}
	var bytes = server.New(constant.ADMIN_PACKET_SERVER_FULL).ToBytes()
	bytes[2] = byte(250)
	a.server.SendRaw(bytes)
	a.server.AwaitClientDisconnect(10)
	a.close()
}

// ADMIN_PACKET_SERVER_SHUTDOWN ends the connection, and NetworkAdmin reconnects
// five seconds later, which is what the reconnect loop is for.
func TestShutdownPacketReconnects(t *testing.T) {
	var a = newAdmin(t)
	if a == nil {
		return
	}
	a.send(server.New(constant.ADMIN_PACKET_SERVER_SHUTDOWN))
	a.server.AwaitClientDisconnect(10)
	a.server.AwaitConnectionCount(2, 10)
	if !a.listener.IsRunning() {
		t.Errorf("expected the listener to keep running after the SHUTDOWN packet")
	}
	a.close()
}

// newBytes is the Java test's new byte[n]: the key material of an AUTH_REQUEST.
func newBytes(count int) []byte {
	return make([]byte, count)
}
