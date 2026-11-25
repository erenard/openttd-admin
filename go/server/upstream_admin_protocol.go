// UpstreamAdminProtocol mirrors com.openttd.server.UpstreamAdminProtocol: the
// values copied from OpenTTD 15.3 that the harness scripts on the wire.
//
// - git show 15.3:src/network/core/config.h: NETWORK_GAME_ADMIN_VERSION = 3
// - git show 15.3:src/network/network_admin.cpp: _admin_update_type_frequencies
//
// Keep in sync with the OpenTTD version in GameVersion; a version bump is one
// commit per release (see AGENTS.md).
package server

import "com.openttd/network/constant"

const GAME_ADMIN_VERSION = 3

// Masks is masks(AdminUpdateFrequency...): the bit mask one update type is
// advertised with.
func Masks(frequencies ...constant.AdminUpdateFrequency) int {
	var mask = 0
	for i := 0; i < len(frequencies); i++ {
		mask |= frequencies[i].Mask()
	}
	return mask
}

// Frequency is FREQUENCIES.get(type), i.e. _admin_update_type_frequencies. The
// four types without a poll bit answer a poll with
// ADMIN_PACKET_SERVER_ERROR (NETWORK_ERROR_ILLEGAL_PACKET).
func Frequency(updateType constant.AdminUpdateType) int {
	switch updateType {
	case constant.ADMIN_UPDATE_DATE:
		return Masks(constant.ADMIN_FREQUENCY_POLL, constant.ADMIN_FREQUENCY_DAILY,
			constant.ADMIN_FREQUENCY_WEEKLY, constant.ADMIN_FREQUENCY_MONTHLY,
			constant.ADMIN_FREQUENCY_QUARTERLY, constant.ADMIN_FREQUENCY_ANUALLY)
	case constant.ADMIN_UPDATE_CLIENT_INFO:
		return Masks(constant.ADMIN_FREQUENCY_POLL, constant.ADMIN_FREQUENCY_AUTOMATIC)
	case constant.ADMIN_UPDATE_COMPANY_INFO:
		return Masks(constant.ADMIN_FREQUENCY_POLL, constant.ADMIN_FREQUENCY_AUTOMATIC)
	case constant.ADMIN_UPDATE_COMPANY_ECONOMY:
		return Masks(constant.ADMIN_FREQUENCY_POLL, constant.ADMIN_FREQUENCY_WEEKLY,
			constant.ADMIN_FREQUENCY_MONTHLY, constant.ADMIN_FREQUENCY_QUARTERLY,
			constant.ADMIN_FREQUENCY_ANUALLY)
	case constant.ADMIN_UPDATE_COMPANY_STATS:
		return Masks(constant.ADMIN_FREQUENCY_POLL, constant.ADMIN_FREQUENCY_WEEKLY,
			constant.ADMIN_FREQUENCY_MONTHLY, constant.ADMIN_FREQUENCY_QUARTERLY,
			constant.ADMIN_FREQUENCY_ANUALLY)
	case constant.ADMIN_UPDATE_CHAT:
		return Masks(constant.ADMIN_FREQUENCY_AUTOMATIC)
	case constant.ADMIN_UPDATE_CONSOLE:
		return Masks(constant.ADMIN_FREQUENCY_AUTOMATIC)
	case constant.ADMIN_UPDATE_CMD_NAMES:
		return Masks(constant.ADMIN_FREQUENCY_POLL)
	case constant.ADMIN_UPDATE_CMD_LOGGING:
		return Masks(constant.ADMIN_FREQUENCY_AUTOMATIC)
	case constant.ADMIN_UPDATE_GAMESCRIPT:
		return Masks(constant.ADMIN_FREQUENCY_AUTOMATIC)
	default:
		// ADMIN_UPDATE_END is never advertised.
		return 0
	}
}

// FrequencyTypes is FREQUENCIES.keySet(): the update types the server advertises,
// in wire order, ADMIN_UPDATE_END excluded.
func FrequencyTypes() []constant.AdminUpdateType {
	var types []constant.AdminUpdateType
	var all = constant.AdminUpdateTypeValues()
	for i := 0; i < len(all); i++ {
		if all[i] != constant.ADMIN_UPDATE_END {
			types = append(types, all[i])
		}
	}
	return types
}

// Frequencies is the mask of each FrequencyTypes(), in the same order: the
// Java tests' entrySet iteration becomes two parallel lists here.
func Frequencies() []int {
	var masks []int
	var types = FrequencyTypes()
	for i := 0; i < len(types); i++ {
		masks = append(masks, Frequency(types[i]))
	}
	return masks
}

// ProtocolPacket is the PROTOCOL packet as sent by
// ServerNetworkAdminSocketHandler::SendProtocol().
func ProtocolPacket() *PacketWriter {
	var packet = New(constant.ADMIN_PACKET_SERVER_PROTOCOL)
	packet.WriteUint8(GAME_ADMIN_VERSION)
	var types = constant.AdminUpdateTypeValues()
	for i := 0; i < len(types); i++ {
		if types[i] == constant.ADMIN_UPDATE_END {
			break
		}
		packet.WriteBool8(true)
		packet.WriteUint16(int(types[i]))
		packet.WriteUint16(Frequency(types[i]))
	}
	packet.WriteBool8(false)
	return packet
}

// WelcomePacket is the WELCOME packet as sent by
// ServerNetworkAdminSocketHandler::SendWelcome().
func WelcomePacket() *PacketWriter {
	var packet = New(constant.ADMIN_PACKET_SERVER_WELCOME)
	packet.WriteString("Strategyboard test server")
	packet.WriteString("15.3")
	packet.WriteBool8(true)
	packet.WriteString("") // Used to be map-name
	packet.WriteUint32(uint32(123456789))
	packet.WriteUint8(1) // Landscape: Temperate
	packet.WriteUint32(uint32(1950 * 365 + 1)) // 1950-01-01
	packet.WriteUint16(64)
	packet.WriteUint16(64)
	return packet
}
