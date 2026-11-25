// GameInfo mirrors com.openttd.network.admin.GameInfo: the state the WELCOME
// packet and the later date packets describe. Java's long day numbers become
// int64 and its short mapSet becomes int. clientsCount and companiesCount are
// public fields in Java because NetworkModel.getGameInfo() recomputes them;
// here they have getters and setters, which is how the Go port reads them.

package admin

import "strconv"

type GameInfo struct {
	version int
	serverName string
	serverRevision string
	serverDedicated bool
	mapSeed int64
	mapSet int
	startDate int64
	mapWidth int
	mapHeight int
	currentDate int64
	clientsCount int
	companiesCount int
}

// NewGameInfo is the constructor.
func NewGameInfo() *GameInfo {
	return &GameInfo{}
}

// The getters, in Java's order.
func (g *GameInfo) GetVersion() int { return g.version }
func (g *GameInfo) GetServerName() string { return g.serverName }
func (g *GameInfo) GetServerRevision() string { return g.serverRevision }
func (g *GameInfo) IsServerDedicated() bool { return g.serverDedicated }
func (g *GameInfo) GetMapSeed() int64 { return g.mapSeed }
func (g *GameInfo) GetMapSet() int { return g.mapSet }
func (g *GameInfo) GetStartDate() int64 { return g.startDate }
func (g *GameInfo) GetMapWidth() int { return g.mapWidth }
func (g *GameInfo) GetMapHeight() int { return g.mapHeight }
func (g *GameInfo) GetCurrentDate() int64 { return g.currentDate }
func (g *GameInfo) GetClientsCount() int { return g.clientsCount }
func (g *GameInfo) GetCompaniesCount() int { return g.companiesCount }

// The setters.
func (g *GameInfo) SetVersion(version int) { g.version = version }
func (g *GameInfo) SetServerName(serverName string) { g.serverName = serverName }
func (g *GameInfo) SetServerRevision(serverRevision string) { g.serverRevision = serverRevision }
func (g *GameInfo) SetServerDedicated(serverDedicated bool) { g.serverDedicated = serverDedicated }
func (g *GameInfo) SetMapSeed(mapSeed int64) { g.mapSeed = mapSeed }
func (g *GameInfo) SetMapSet(mapSet int) { g.mapSet = mapSet }
func (g *GameInfo) SetStartDate(startDate int64) { g.startDate = startDate }
func (g *GameInfo) SetMapWidth(mapWidth int) { g.mapWidth = mapWidth }
func (g *GameInfo) SetMapHeight(mapHeight int) { g.mapHeight = mapHeight }
func (g *GameInfo) SetCurrentDate(currentDate int64) { g.currentDate = currentDate }
func (g *GameInfo) SetClientsCount(clientsCount int) { g.clientsCount = clientsCount }
func (g *GameInfo) SetCompaniesCount(companiesCount int) { g.companiesCount = companiesCount }

// ToString is toString: four PrintWriter.println lines, so four lines ending
// with a newline.
func (g *GameInfo) ToString() string {
	return "GameInfo v" + strconv.Itoa(g.version) + ".\n" +
		"Server:" + g.serverName + ", " + g.serverRevision + ", dedicated:" + boolText(g.serverDedicated) + ".\n" +
		"Map:" + strconv.Itoa(g.mapSet) + ", " + strconv.Itoa(g.mapWidth) + "x" + strconv.Itoa(g.mapHeight) + ".\n" +
		"Game:" + strconv.FormatInt(g.currentDate, 10) + ", started:" + strconv.FormatInt(g.startDate, 10) + ".\n"
}

// Copy is clone: a new GameInfo with every field carried over.
func (g *GameInfo) Copy() *GameInfo {
	return &GameInfo{
		version: g.version,
		serverName: g.serverName,
		serverRevision: g.serverRevision,
		serverDedicated: g.serverDedicated,
		mapSeed: g.mapSeed,
		mapSet: g.mapSet,
		startDate: g.startDate,
		mapWidth: g.mapWidth,
		mapHeight: g.mapHeight,
		currentDate: g.currentDate,
		clientsCount: g.clientsCount,
		companiesCount: g.companiesCount,
	}
}

