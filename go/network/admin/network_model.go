// NetworkModel mirrors com.openttd.network.admin.NetworkModel: the three maps
// NetworkAdmin fills in as packets arrive, the lock that guards them while
// NetworkAdminSender hands a snapshot to the application, and the clone that
// makes a snapshot safe to read after the connection moves on.
//
// Java keys clients by Long and companies by Short; both are int here, since
// OpenTTD sends both as uint32 or uint16. Java's HashMap iteration order is
// unspecified; retreiveCompanies and retreiveClients return the values sorted
// by id, so the Go port is at least deterministic.

package admin

import "sync"

type NetworkModel struct {
	gameInfo *GameInfo
	companies map[int]*Company
	clients map[int]*Client
	commandNames map[int]string
	lock sync.Mutex
}

// NewNetworkModel is the constructor: Java's field initializers create the
// GameInfo and the three maps.
func NewNetworkModel() *NetworkModel {
	return &NetworkModel{
		gameInfo: NewGameInfo(),
		companies: map[int]*Company{},
		clients: map[int]*Client{},
		commandNames: map[int]string{},
	}
}

// GetGameInfo is getGameInfo: it refreshes the two counts from the maps before
// handing the GameInfo over, exactly as Java does.
func (m *NetworkModel) GetGameInfo() *GameInfo {
	m.gameInfo.clientsCount = len(m.clients)
	m.gameInfo.companiesCount = len(m.companies)
	return m.gameInfo
}

// SetGameInfo is setGameInfo.
func (m *NetworkModel) SetGameInfo(gameInfo *GameInfo) { m.gameInfo = gameInfo }

// RetreiveCompany is retreiveCompany: create the Company if this is the first
// packet about it, then return it. Java spells the method with the same typo.
func (m *NetworkModel) RetreiveCompany(companyId int) *Company {
	var found, ok = m.companies[companyId]
	if !ok {
		found = NewCompany(companyId)
		m.companies[companyId] = found
	}
	return found
}

// RetreiveCompanies is retreiveCompanies: the values, sorted by id.
func (m *NetworkModel) RetreiveCompanies() []*Company {
	var list []int
	for id, _ := range m.companies {
		list = append(list, id)
	}
	sortInts(list)
	var companies []*Company
	for i := 0; i < len(list); i++ {
		var company, ok = m.companies[list[i]]
		if ok {
			companies = append(companies, company)
		}
	}
	return companies
}

// DeleteCompany is deleteCompany.
func (m *NetworkModel) DeleteCompany(companyId int) {
	delete(m.companies, companyId)
}

// RetreiveClient is retreiveClient: create the Client if this is the first
// packet about it, then return it.
func (m *NetworkModel) RetreiveClient(clientId int) *Client {
	var found, ok = m.clients[clientId]
	if !ok {
		found = NewClient(clientId)
		m.clients[clientId] = found
	}
	return found
}

// RetreiveClients is retreiveClients: the values, sorted by id.
func (m *NetworkModel) RetreiveClients() []*Client {
	var list []int
	for id, _ := range m.clients {
		list = append(list, id)
	}
	sortInts(list)
	var clients []*Client
	for i := 0; i < len(list); i++ {
		var client, ok = m.clients[list[i]]
		if ok {
			clients = append(clients, client)
		}
	}
	return clients
}

// DeleteClient is deleteClient.
func (m *NetworkModel) DeleteClient(clientId int) {
	delete(m.clients, clientId)
}

// PutCommandName is putCommandName: the DoCommand names announced by
// ADMIN_PACKET_SERVER_CMD_NAMES. Not stable across OpenTTD versions, used for
// logging only.
func (m *NetworkModel) PutCommandName(commandId int, name string) {
	m.commandNames[commandId] = name
}

// RetreiveCommandNames is retreiveCommandNames.
func (m *NetworkModel) RetreiveCommandNames() map[int]string {
	return m.commandNames
}

// Lock and Unlock are lock and unlock: Java exposes its ReentrantLock directly.
func (m *NetworkModel) Lock() { m.lock.Lock() }
func (m *NetworkModel) Unlock() { m.lock.Unlock() }

// Snapshot is snapshot: the lock, a clone, then the unlock.
func (m *NetworkModel) Snapshot() *NetworkModel {
	m.lock.Lock()
	var clone = cloneModel(m)
	m.lock.Unlock()
	return clone
}

// Copy is copy: a clone without the lock, which is how NetworkAdmin hands the
// model to Game.updateAdmin while it is still writing. Java logs the
// CloneNotSupportedException it can never get; there is nothing to log here.
func (m *NetworkModel) Copy() *NetworkModel {
	return cloneModel(m)
}

// cloneModel is clone: a new NetworkModel with a cloned GameInfo, cloned
// Companies and Clients, and the command names copied over.
func cloneModel(m *NetworkModel) *NetworkModel {
	var clone = NewNetworkModel()
	clone.SetGameInfo(m.gameInfo.Copy())
	for id, company := range m.companies {
		clone.companies[id] = company.Copy()
	}
	for id, client := range m.clients {
		clone.clients[id] = client.Copy()
	}
	for id, name := range m.commandNames {
		clone.commandNames[id] = name
	}
	return clone
}

// sortInts is the insertion sort Java's HashMap makes unnecessary: Go maps
// iterate in a random order, and the tests and logs read better when the ids
// come back sorted.
func sortInts(list []int) {
	for i := 1; i < len(list); i++ {
		var value = list[i]
		var j = i - 1
		for j >= 0 && list[j] > value {
			list[j+1] = list[j]
			j = j - 1
		}
		list[j+1] = value
	}
}
