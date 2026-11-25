// The network/admin data model: the classes NetworkAdmin fills in as packets
// arrive, and the snapshot NetworkAdminSender hands to the application.
//
// See test/java/com/openttd/network/admin/, whose assertions about the model
// these tests reproduce: the toString formats the logs print, Company's
// default array lengths, and the fact that a snapshot keeps reading the state
// it was taken at.

package admin

import "strings"
import "testing"

import "com.openttd/network/constant"

// The Java tests never construct a NetworkException directly, but NetworkAdmin
// throws one for every error id the server reports, and the message is the
// enum's name: NetworkErrorCode.valueOf(errorId).toString().
func TestException(t *testing.T) {
	var byId = NewExceptionFromError(18)
	if byId.GetExceptionMessage() != "NETWORK_ERROR_TIMEOUT_MAP" {
		t.Errorf("error 19: expected NETWORK_ERROR_TIMEOUT_MAP, got %v", byId.GetExceptionMessage())
	}
	if byId.Error() != "NETWORK_ERROR_TIMEOUT_MAP" {
		t.Errorf("error 19: getMessage and the error interface must agree")
	}

	// Java's valueOf returns null for an id outside the enum, which leaves the
	// RuntimeException's message null; here it is the empty string.
	if NewExceptionFromError(24).GetExceptionMessage() != "" {
		t.Errorf("error 24: expected the empty message")
	}

	var byMessage = NewException("Connection lost, received data length: 3")
	if byMessage.GetExceptionMessage() != "Connection lost, received data length: 3" {
		t.Errorf("message constructor: %v", byMessage.GetExceptionMessage())
	}

	// The log line NetworkAdmin prints for an error id.
	if DescribeError(0) != "NETWORK_ERROR_GENERAL" {
		t.Errorf("error 0: expected NETWORK_ERROR_GENERAL, got %v", DescribeError(0))
	}
	if DescribeError(22) != "NETWORK_ERROR_NO_AUTHENTICATION_METHOD_AVAILABLE" {
		t.Errorf("error 22: got %v", DescribeError(22))
	}
}

// Client: toString is id + ":" + name + ", " + ip + ", " + companyId, and clone
// carries every field.
func TestClient(t *testing.T) {
	var client = NewClient(1234)
	client.SetName("Eric")
	client.SetIp("127.0.0.1")
	client.SetJoinDate(int64(1700000000000))
	client.SetCompany(2)

	if client.ToString() != "1234:Eric, 127.0.0.1, 2" {
		t.Errorf("client toString: %v", client.ToString())
	}
	if client.GetId() != 1234 {
		t.Errorf("client id: %v", client.GetId())
	}
	if client.GetJoinDate() != int64(1700000000000) {
		t.Errorf("join date: %v", client.GetJoinDate())
	}

	var copy = client.Copy()
	copy.SetName("Other")
	if client.GetName() != "Eric" {
		t.Errorf("clone must be independent, the original reads %v", client.GetName())
	}
	if copy.ToString() != "1234:Other, 127.0.0.1, 2" {
		t.Errorf("client copy toString: %v", copy.ToString())
	}
}

// GameInfo: four PrintWriter.println lines, and the counts NetworkModel
// recomputes.
func TestGameInfo(t *testing.T) {
	var info = NewGameInfo()
	info.SetVersion(3)
	info.SetServerName("Strategyboard")
	info.SetServerRevision("r123")
	info.SetServerDedicated(true)
	info.SetMapSet(2)
	info.SetMapWidth(256)
	info.SetMapHeight(192)
	info.SetStartDate(int64(1950))
	info.SetCurrentDate(int64(1951))
	info.SetClientsCount(4)
	info.SetCompaniesCount(2)

	var expected = "GameInfo v3.\n" +
		"Server:Strategyboard, r123, dedicated:true.\n" +
		"Map:2, 256x192.\n" +
		"Game:1951, started:1950.\n"
	if info.ToString() != expected {
		t.Errorf("game info toString:\n%v\nexpected\n%v", info.ToString(), expected)
	}

	var copy = info.Copy()
	copy.SetServerName("Other")
	if info.GetServerName() != "Strategyboard" {
		t.Errorf("game info clone must be independent")
	}
	if copy.GetClientsCount() != 4 || copy.GetCompaniesCount() != 2 {
		t.Errorf("game info clone counts: %v, %v", copy.GetClientsCount(), copy.GetCompaniesCount())
	}
}

// Company: the default array lengths Java's field initializers give, the
// tab separated toString, and a clone that deep-copies the arrays.
func TestCompany(t *testing.T) {
	var company = NewCompany(2)
	if len(company.GetShareOwners()) != 4 {
		t.Errorf("shareOwners: expected 4 entries, got %v", len(company.GetShareOwners()))
	}
	if len(company.GetVehicules()) != 5 {
		t.Errorf("vehicules: expected NetworkVehicleType.length() entries, got %v", len(company.GetVehicules()))
	}
	if len(company.GetStations()) != 5 {
		t.Errorf("stations: expected 5 entries, got %v", len(company.GetStations()))
	}
	if company.GetLoan() != uint64(0) || company.GetLastValue() != uint64(0) {
		t.Errorf("Java leaves the BigInteger fields null; the Go port reads 0")
	}

	company.SetName("Company")
	company.SetPresident("Eric")
	company.SetInauguratedYear(int64(1950))
	company.SetUsePassword(false)
	company.SetMoney(int64(1000))
	company.SetIncome(int64(20))
	company.SetLastValue(uint64(300))
	company.SetLastPerformance(40)

	var expected = "2:Company, 1950, password:false\t1000\t20\t300\t40"
	if company.ToString() != expected {
		t.Errorf("company toString:\n%v\nexpected\n%v", company.ToString(), expected)
	}

	// The COMPANY_UPDATE packet writes four share owners; the clone must not
	// share that array with the original.
	var shares = []int{1, 2, 3, 4}
	company.SetShareOwners(shares)
	var copy = company.Copy()
	copy.SetShareOwners([]int{9, 9, 9, 9})
	if company.GetShareOwners()[0] != 1 {
		t.Errorf("company clone shares its shareOwners array")
	}
	if copy.GetInauguratedYear() != int64(1950) || copy.GetName() != "Company" {
		t.Errorf("company clone lost a field")
	}
	if copy.GetLastValue() != uint64(300) {
		t.Errorf("company clone lastValue: %v", copy.GetLastValue())
	}
}

// NetworkModel: retreiveCompany and retreiveClient create on demand, getGameInfo
// recomputes the counts, and snapshot() keeps the state it was taken at.
func TestModel(t *testing.T) {
	var model = NewNetworkModel()

	var company = model.RetreiveCompany(2)
	if company.GetId() != 2 {
		t.Errorf("retreiveCompany id: %v", company.GetId())
	}
	if model.RetreiveCompany(2) != company {
		t.Errorf("retreiveCompany must return the same Company for the same id")
	}
	model.RetreiveCompany(1)

	var client = model.RetreiveClient(1234)
	if client.GetId() != 1234 {
		t.Errorf("retreiveClient id: %v", client.GetId())
	}
	if model.RetreiveClient(1234) != client {
		t.Errorf("retreiveClient must return the same Client for the same id")
	}
	model.RetreiveClient(4321)

	// getGameInfo recomputes both counts from the maps.
	var info = model.GetGameInfo()
	if info.GetClientsCount() != 2 {
		t.Errorf("clientsCount: %v", info.GetClientsCount())
	}
	if info.GetCompaniesCount() != 2 {
		t.Errorf("companiesCount: %v", info.GetCompaniesCount())
	}

	// The companies and clients come back sorted by id, so the logs and the
	// tests are deterministic even though Go maps are not.
	var companies = model.RetreiveCompanies()
	if len(companies) != 2 || companies[0].GetId() != 1 || companies[1].GetId() != 2 {
		t.Errorf("retreiveCompanies order")
	}
	var clients = model.RetreiveClients()
	if len(clients) != 2 || clients[0].GetId() != 1234 || clients[1].GetId() != 4321 {
		t.Errorf("retreiveClients order")
	}

	// NetworkAdmin hands Game.updateAdmin a copy, then keeps writing to its own
	// model; the copy must not move.
	client.SetName("Before")
	var snapshot = model.Snapshot()
	client.SetName("After")
	model.DeleteClient(4321)
	model.DeleteCompany(1)

	var snapClient = snapshot.RetreiveClient(1234)
	if snapClient.GetName() != "Before" {
		t.Errorf("snapshot client name: %v", snapClient.GetName())
	}
	if len(snapshot.RetreiveClients()) != 2 {
		t.Errorf("snapshot still has both clients")
	}
	if len(snapshot.RetreiveCompanies()) != 2 {
		t.Errorf("snapshot still has both companies")
	}
	if snapshot.RetreiveCompany(2) == company {
		t.Errorf("snapshot must not share Company objects with the live model")
	}

	// The DoCommand names are copied over too.
	model.PutCommandName(4, "add_building")
	var names = model.RetreiveCommandNames()
	var commandName, hasName = names[4]
	if !hasName || commandName != "add_building" {
		t.Errorf("command names: %v", commandName)
	}
	var withNames = model.Copy()
	withNames.PutCommandName(5, "build_bridge")
	if len(withNames.RetreiveCommandNames()) != 2 {
		t.Errorf("copy command names: %v", len(withNames.RetreiveCommandNames()))
	}
	if len(model.RetreiveCommandNames()) != 1 {
		t.Errorf("copy must not share the command name map")
	}
}

// NetworkAdminEvent: the setXEvent helpers fill in the fields the packet
// carries, and toString is the line Game.updateAdmin logs.
func TestAdminEvent(t *testing.T) {
	var event = NewNetworkAdminEvent(constant.ADMIN_PACKET_SERVER_CLIENT_INFO)
	event.setClientEvent(constant.ADMIN_PACKET_SERVER_CLIENT_INFO, int64(1234))
	if event.GetClientId() != int64(1234) {
		t.Errorf("client event id: %v", event.GetClientId())
	}
	if int(event.GetPacketServerType()) != int(constant.ADMIN_PACKET_SERVER_CLIENT_INFO) {
		t.Errorf("packet type: %v", event.GetPacketServerType())
	}

	// Java prints "null" for the two enum fields no packet has set.
	var text = event.ToString()
	if !strings.Contains(text, "ADMIN_PACKET_SERVER_CLIENT_INFO clientId:1234 pingId:0 companyId:0") {
		t.Errorf("event toString: %v", text)
	}
	if !strings.Contains(text, "companyRemoveReason:null clientError:NETWORK_ERROR_NONE") {
		t.Errorf("event toString unset enums: %v", text)
	}

	// A COMPANY_REMOVE packet sets the reason; CLIENT_ERROR sets the code.
	var removed = NewNetworkAdminEvent(constant.ADMIN_PACKET_SERVER_COMPANY_REMOVE)
	removed.setCompanyRemoveEvent(constant.ADMIN_PACKET_SERVER_COMPANY_REMOVE, 3, 2)
	if removed.GetCompanyRemoveReason() != 2 {
		t.Errorf("company remove reason: %v", removed.GetCompanyRemoveReason())
	}
	if !strings.Contains(removed.ToString(), "companyRemoveReason:ADMIN_CRR_BANKRUPT") {
		t.Errorf("company remove toString: %v", removed.ToString())
	}

	var errored = NewNetworkAdminEvent(constant.ADMIN_PACKET_SERVER_CLIENT_ERROR)
	errored.setClientErrorEvent(constant.ADMIN_PACKET_SERVER_CLIENT_ERROR, int64(7), 18)
	if errored.GetClientError() != constant.NETWORK_ERROR_TIMEOUT_MAP {
		t.Errorf("client error: %v", errored.GetClientError())
	}

	// A PONG only carries the ping id.
	var pong = NewNetworkAdminEvent(constant.ADMIN_PACKET_SERVER_PONG)
	pong.setPongEvent(int64(9))
	if pong.GetPingId() != int64(9) {
		t.Errorf("pong pingId: %v", pong.GetPingId())
	}

	// The console event has no packet type argument in Java: the origin and
	// the message are the whole payload.
	var console = NewNetworkAdminEvent(constant.ADMIN_PACKET_SERVER_CONSOLE)
	console.setConsoleEvent("status", "Players: 1")
	if console.GetOrigin() != "status" || console.GetMessage() != "Players: 1" {
		t.Errorf("console event: %v %v", console.GetOrigin(), console.GetMessage())
	}
}

// NetworkEvent is the same payload on the client side of the library: admin/model
// Game reads it, and it has no toString of its own.
func TestGameEvent(t *testing.T) {
	var event = NewNetworkEvent(constant.ADMIN_PACKET_SERVER_CHAT)
	event.setChatEvent(constant.ADMIN_PACKET_SERVER_CHAT, int64(12), "hello")
	if event.GetClientId() != int64(12) || event.GetMessage() != "hello" {
		t.Errorf("chat event: %v %v", event.GetClientId(), event.GetMessage())
	}

	event.setDateEvent(constant.ADMIN_PACKET_SERVER_CHAT, int64(1950))
	if event.GetDate() != int64(1950) {
		t.Errorf("date event: %v", event.GetDate())
	}

	event.setConsoleEvent("status", "Money: 100")
	if event.GetOrigin() != "status" {
		t.Errorf("console event origin: %v", event.GetOrigin())
	}

	event.setRConEvent(constant.ADMIN_PACKET_SERVER_CHAT, 1, "save")
	if event.GetColor() != 1 {
		t.Errorf("rcon color: %v", event.GetColor())
	}
}
