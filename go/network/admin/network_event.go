// NetworkEvent mirrors com.openttd.network.admin.NetworkEvent: the packet
// payload the client side of the library hands to admin/model/Game. It is a
// plain class in Java, not an interface, and the setXEvent helpers are package
// private: only NetworkAdmin calls them, and they ignore the packetServerType
// argument because the constructor already fixed it. Java's long clientId and
// date become int64, its short and char ids become int.

package admin

import "com.openttd/network/constant"

type NetworkEvent struct {
	packetServerType constant.PacketServerType
	clientId         int64
	company          int
	extraId          int
	color            int
	origin           string
	message          string
	date             int64
}

// NewNetworkEvent is the constructor. Java's field is final, so the packet
// type never changes afterwards.
func NewNetworkEvent(packetServerType constant.PacketServerType) *NetworkEvent {
	return &NetworkEvent{packetServerType: packetServerType}
}

// The setXEvent helpers, package private in Java. Each one ignores its
// packetServerType argument, exactly as Java does.
func (e *NetworkEvent) setClientEvent(packetServerType constant.PacketServerType, clientId int64) {
	e.clientId = clientId
}

func (e *NetworkEvent) setCompanyEvent(packetServerType constant.PacketServerType, company int) {
	e.company = company
}

func (e *NetworkEvent) setRConEvent(packetServerType constant.PacketServerType, color int, message string) {
	e.color = color
	e.message = message
}

func (e *NetworkEvent) setDateEvent(packetServerType constant.PacketServerType, date int64) {
	e.date = date
}

func (e *NetworkEvent) setChatEvent(packetServerType constant.PacketServerType, clientId int64, message string) {
	e.clientId = clientId
	e.message = message
}

func (e *NetworkEvent) setGameScriptEvent(packetServerType constant.PacketServerType, json string) {
	e.message = json
}

func (e *NetworkEvent) setConsoleEvent(origin string, message string) {
	e.origin = origin
	e.message = message
}

// The public getters.
func (e *NetworkEvent) GetClientId() int64                             { return e.clientId }
func (e *NetworkEvent) GetCompany() int                                { return e.company }
func (e *NetworkEvent) GetExtraId() int                                { return e.extraId }
func (e *NetworkEvent) GetOrigin() string                              { return e.origin }
func (e *NetworkEvent) GetMessage() string                             { return e.message }
func (e *NetworkEvent) GetDate() int64                                 { return e.date }
func (e *NetworkEvent) GetColor() int                                  { return e.color }
func (e *NetworkEvent) GetPacketServerType() constant.PacketServerType { return e.packetServerType }

// SetClientId and SetExtraId are the two public setters Java has.
func (e *NetworkEvent) SetClientId(clientId int64) { e.clientId = clientId }
func (e *NetworkEvent) SetExtraId(extraId int)     { e.extraId = extraId }
