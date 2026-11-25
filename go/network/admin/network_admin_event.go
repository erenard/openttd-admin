// NetworkAdminEvent mirrors com.openttd.network.admin.NetworkAdminEvent: what
// NetworkAdmin decodes from a server packet, before admin/model/Game turns it
// into the functional events the application listens for.
//
// Java's companyRemoveReason and clientError are enum fields left null until a
// COMPANY_REMOVE or CLIENT_ERROR packet sets them; here they are ints with -1
// meaning unset, and ToString prints "null" for them, as Java's
// StringBuilder.append(null) does. Java appends its char color as a character;
// the Go port prints the number, which is what the number on the wire is.

package admin

import "strconv"

import "com.openttd/network/constant"

type NetworkAdminEvent struct {
	packetServerType constant.PacketServerType
	clientId int64
	pingId int64
	company int
	color int
	origin string
	message string
	date int64
	companyRemoveReason int
	clientError constant.NetworkErrorCode
}

// NewNetworkAdminEvent is the constructor.
func NewNetworkAdminEvent(packetServerType constant.PacketServerType) *NetworkAdminEvent {
	return &NetworkAdminEvent{
		packetServerType: packetServerType,
		companyRemoveReason: -1,
		clientError: constant.NETWORK_ERROR_NONE,
	}
}

// The setXEvent helpers, package private in Java, each ignoring its packet
// type argument as Java does.
func (e *NetworkAdminEvent) setClientEvent(packetServerType constant.PacketServerType, clientId int64) {
	e.clientId = clientId
}

func (e *NetworkAdminEvent) setCompanyEvent(packetServerType constant.PacketServerType, company int) {
	e.company = company
}

func (e *NetworkAdminEvent) setRConEvent(packetServerType constant.PacketServerType, color int, message string) {
	e.color = color
	e.message = message
}

func (e *NetworkAdminEvent) setRConEndEvent(packetType constant.PacketServerType, rcon string) {
	e.message = rcon
}

func (e *NetworkAdminEvent) setDateEvent(packetServerType constant.PacketServerType, date int64) {
	e.date = date
}

func (e *NetworkAdminEvent) setChatEvent(packetServerType constant.PacketServerType, clientId int64, message string) {
	e.clientId = clientId
	e.message = message
}

func (e *NetworkAdminEvent) setGameScriptEvent(packetServerType constant.PacketServerType, json string) {
	e.message = json
}

func (e *NetworkAdminEvent) setConsoleEvent(origin string, message string) {
	e.origin = origin
	e.message = message
}

func (e *NetworkAdminEvent) setPongEvent(pingId int64) {
	e.pingId = pingId
}

func (e *NetworkAdminEvent) setCompanyRemoveEvent(packetType constant.PacketServerType, company int, reason int) {
	e.company = company
	e.companyRemoveReason = reason
}

func (e *NetworkAdminEvent) setClientErrorEvent(packetType constant.PacketServerType, clientId int64, error int) {
	e.clientId = clientId
	e.clientError = constant.NetworkErrorCodeOf(error)
}

// The public getters.
func (e *NetworkAdminEvent) GetClientId() int64 { return e.clientId }
func (e *NetworkAdminEvent) GetCompany() int { return e.company }
func (e *NetworkAdminEvent) GetPingId() int64 { return e.pingId }
func (e *NetworkAdminEvent) GetOrigin() string { return e.origin }
func (e *NetworkAdminEvent) GetMessage() string { return e.message }
func (e *NetworkAdminEvent) GetDate() int64 { return e.date }
func (e *NetworkAdminEvent) GetColor() int { return e.color }
func (e *NetworkAdminEvent) GetPacketServerType() constant.PacketServerType { return e.packetServerType }

// GetCompanyRemoveReason is getCompanyRemoveReason: -1 when no COMPANY_REMOVE
// packet set it.
func (e *NetworkAdminEvent) GetCompanyRemoveReason() int { return e.companyRemoveReason }

// GetClientError is getClientError, NETWORK_ERROR_NONE when no CLIENT_ERROR
// packet set it.
func (e *NetworkAdminEvent) GetClientError() constant.NetworkErrorCode { return e.clientError }

// ToString is toString: the packet type, then every field labelled.
func (e *NetworkAdminEvent) ToString() string {
	var sb = constant.PacketServerTypeName(e.packetServerType)
	sb = sb + " clientId:" + strconv.FormatInt(e.clientId, 10)
	sb = sb + " pingId:" + strconv.FormatInt(e.pingId, 10)
	sb = sb + " companyId:" + strconv.Itoa(e.company)
	sb = sb + " color:" + strconv.Itoa(e.color)
	sb = sb + " origin:" + e.origin
	sb = sb + " message:" + e.message
	sb = sb + " date:" + strconv.FormatInt(e.date, 10)
	sb = sb + " companyRemoveReason:" + companyRemoveReasonText(e.companyRemoveReason)
	sb = sb + " clientError:" + constant.NetworkErrorCodeName(e.clientError)
	return sb
}

// companyRemoveReasonText is the enum name Java would print, or "null" while
// the field is unset.
func companyRemoveReasonText(reason int) string {
	if reason < 0 {
		return "null"
	}
	return constant.AdminCompanyRemoveReasonName(constant.AdminCompanyRemoveReason(reason))
}
