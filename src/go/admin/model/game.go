// model mirrors com.openttd.admin.model.Game: the game state the facade keeps,
// and the translation layer from the technical NetworkAdminEvent the listener
// decodes to the functional events the application listens for.
//
// Java's updateAdmin() is the whole contract: it swaps in the model snapshot the
// listener took for this packet, then maps one packet type onto one event, and
// for WELCOME it seeds the polling schedule. The completeness table in
// test/java/com/openttd/admin/model/GameEventMappingTest.java is what pins that
// switch, so the cases below are in the Java order and none of them is missing.

package model

import "com.openttd/admin/event"
import "com.openttd/network/admin"
import "com.openttd/network/constant"
import "com.openttd/util"

var modelLog = util.GetLogger("Game")

type Game struct {
	eventDispatcher *event.EventDispatcher
	// networkModel is the snapshot last handed over by updateAdmin. Java's field
	// is not volatile and is written on the listener thread while the application
	// reads it on its own, so the port keeps Java's (loose) behaviour.
	networkModel *admin.NetworkModel
}

// NewGame is Game(EventDispatcher eventDispatcher).
func NewGame(eventDispatcher *event.EventDispatcher) *Game {
	return &Game{eventDispatcher: eventDispatcher, networkModel: admin.NewNetworkModel()}
}

// GetGameInfo is getGameInfo.
func (g *Game) GetGameInfo() *admin.GameInfo {
	return g.networkModel.GetGameInfo()
}

// GetDate is getDate: the ingame date as a java-style Calendar, which is what
// DateEvent.toString() reads off the Openttd seam.
func (g *Game) GetDate() util.Calendar {
	return util.DayToCalendar(g.networkModel.GetGameInfo().GetCurrentDate())
}

// GetClient is getClient(long clientId).
func (g *Game) GetClient(clientId int) *admin.Client {
	return g.networkModel.RetreiveClient(clientId)
}

// GetClients is getClients().
func (g *Game) GetClients() []*admin.Client {
	return g.networkModel.RetreiveClients()
}

// GetClientsOf is the getClients(short companyId) overload: the clients that
// belong to one company. Go has no overloading, so the name says what the
// argument is.
func (g *Game) GetClientsOf(companyId int) []*admin.Client {
	var clients []*admin.Client
	for _, client := range g.GetClients() {
		if client.GetCompany() == companyId {
			clients = append(clients, client)
		}
	}
	return clients
}

// GetCompany is getCompany(short companyId).
func (g *Game) GetCompany(companyId int) *admin.Company {
	return g.networkModel.RetreiveCompany(companyId)
}

// GetCompanies is getCompanies().
func (g *Game) GetCompanies() []*admin.Company {
	return g.networkModel.RetreiveCompanies()
}

// GetCommandNames is getCommandNames(): the DoCommand names the server announced,
// for logging only, as the ids and names are not stable across OpenTTD versions.
func (g *Game) GetCommandNames() map[int]string {
	return g.networkModel.RetreiveCommandNames()
}

// UpdateAdmin is updateAdmin(NetworkAdminEvent, NetworkModel, NetworkAdminSender)
// and the seam admin/network/admin's Game interface. The snapshot replaces the
// field before the switch, so an event listener that reads the model sees this
// packet's state, not the previous one's.
func (g *Game) UpdateAdmin(e *admin.NetworkAdminEvent, networkModel *admin.NetworkModel, send *admin.NetworkAdminSender) {
	modelLog.Debug("Game.updateAdmin " + e.ToString())
	g.networkModel = networkModel
	switch e.GetPacketServerType() {
	case constant.ADMIN_PACKET_SERVER_CHAT:
		g.eventDispatcher.Dispatch(event.NewChatEvent(g, e.GetClientId(), e.GetMessage()))
	case constant.ADMIN_PACKET_SERVER_DATE:
		g.eventDispatcher.Dispatch(event.NewDateEvent(g))
	case constant.ADMIN_PACKET_SERVER_NEWGAME:
		// Java has this case with an empty body: the new game's WELCOME packet
		// replaces the model, so there is nothing to announce here.
	case constant.ADMIN_PACKET_SERVER_RCON:
		g.eventDispatcher.Dispatch(event.NewRConEvent(g, e.GetColor(), e.GetMessage()))
	case constant.ADMIN_PACKET_SERVER_WELCOME:
		// The schedule a welcome implies: the cheap, always wanted updates
		// automatic, the expensive statistics weekly, and one immediate poll of
		// the four things the application will want before its first tick.
		send.UpdateFrequency(constant.ADMIN_UPDATE_CLIENT_INFO, constant.ADMIN_FREQUENCY_AUTOMATIC)
		send.UpdateFrequency(constant.ADMIN_UPDATE_COMPANY_INFO, constant.ADMIN_FREQUENCY_AUTOMATIC)
		send.UpdateFrequency(constant.ADMIN_UPDATE_COMPANY_STATS, constant.ADMIN_FREQUENCY_WEEKLY)
		send.UpdateFrequency(constant.ADMIN_UPDATE_COMPANY_ECONOMY, constant.ADMIN_FREQUENCY_WEEKLY)
		send.UpdateFrequency(constant.ADMIN_UPDATE_CHAT, constant.ADMIN_FREQUENCY_AUTOMATIC)
		send.UpdateFrequency(constant.ADMIN_UPDATE_GAMESCRIPT, constant.ADMIN_FREQUENCY_AUTOMATIC)
		send.UpdateFrequency(constant.ADMIN_UPDATE_CONSOLE, constant.ADMIN_FREQUENCY_AUTOMATIC)
		send.UpdateFrequency(constant.ADMIN_UPDATE_DATE, constant.ADMIN_FREQUENCY_DAILY)
		send.PollDate()
		send.PollClientInfo(int64(-1))
		send.PollCompanyInfo(-1)
		send.PollCompanyEconomy(-1)
	case constant.ADMIN_PACKET_SERVER_CLIENT_JOIN:
		g.eventDispatcher.Dispatch(event.NewClientEvent(g, e.GetClientId(), event.ACTION_CREATE))
	// Java writes these as fall-through groups; Go's cases do not fall through,
	// so the ids that share a body are listed on one label. Written the Java way
	// the first id of each group would get an empty body and dispatch nothing.
	case constant.ADMIN_PACKET_SERVER_CLIENT_INFO, constant.ADMIN_PACKET_SERVER_CLIENT_UPDATE:
		g.eventDispatcher.Dispatch(event.NewClientEvent(g, e.GetClientId(), event.ACTION_UPDATE))
	case constant.ADMIN_PACKET_SERVER_CLIENT_ERROR, constant.ADMIN_PACKET_SERVER_CLIENT_QUIT:
		g.eventDispatcher.Dispatch(event.NewClientEvent(g, e.GetClientId(), event.ACTION_DELETE))
	case constant.ADMIN_PACKET_SERVER_COMPANY_NEW:
		g.eventDispatcher.Dispatch(event.NewCompanyEvent(g, e.GetCompany(), event.ACTION_CREATE))
	case constant.ADMIN_PACKET_SERVER_COMPANY_INFO, constant.ADMIN_PACKET_SERVER_COMPANY_UPDATE, constant.ADMIN_PACKET_SERVER_COMPANY_ECONOMY, constant.ADMIN_PACKET_SERVER_COMPANY_STATS:
		g.eventDispatcher.Dispatch(event.NewCompanyEvent(g, e.GetCompany(), event.ACTION_UPDATE))
	case constant.ADMIN_PACKET_SERVER_COMPANY_REMOVE:
		g.eventDispatcher.Dispatch(event.NewCompanyEvent(g, e.GetCompany(), event.ACTION_DELETE))
	case constant.ADMIN_PACKET_SERVER_GAMESCRIPT:
		g.eventDispatcher.Dispatch(event.NewGameScriptEvent(g, e.GetMessage()))
	case constant.ADMIN_PACKET_SERVER_RCON_END:
		g.eventDispatcher.Dispatch(event.NewRConEndEvent(g, e.GetMessage()))
	case constant.ADMIN_PACKET_SERVER_CONSOLE:
		g.eventDispatcher.Dispatch(event.NewConsoleEvent(g, e.GetOrigin(), e.GetMessage()))
	case constant.ADMIN_PACKET_SERVER_PONG:
		g.eventDispatcher.Dispatch(event.NewPongEvent(g, e.GetPingId()))
	}
}
