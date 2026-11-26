// Protocol mirrors com.openttd.network.admin.Protocol: the frequency table the
// server sends in its PROTOCOL packet, which gates NetworkAdminSender's
// updateFrequency() and poll().
//
// The wire format is ServerNetworkAdminSocketHandler::SendProtocol() in 15.3:
// uint8 version, then per AdminUpdateType below ADMIN_UPDATE_END a bool8 plus
// the uint16 type and uint16 frequency mask, then a final bool8 false.
package admin

import "strconv"
import "fmt"
import "com.openttd/network/constant"

type Protocol struct {
	version int
	// masks is Java's EnumMap<AdminUpdateType, Integer>, keyed by ordinal: the
	// frequency mask the server advertised for that update type, or -1 when it
	// advertised none (Java's absent key).
	masks []int
}

// New is the constructor: the version the server announced, no frequencies yet.
func New(version int) *Protocol {
	var masks []int
	var types = constant.AdminUpdateTypeValues()
	for i := 0; i < len(types); i++ {
		masks = append(masks, -1)
	}
	return &Protocol{version: version, masks: masks}
}

func (p *Protocol) GetVersion() int { return p.version }

// PutProtocol is putProtocol(key, value): key is the uint16 read off the wire.
// Java's EnumMap throws for an ordinal it does not know, i.e. an update type a
// newer OpenTTD added; the port ignores the entry instead, so it never matches
// and the admin keeps working.
func (p *Protocol) PutProtocol(key int, value int) {
	if key < 0 || key >= len(p.masks) {
		return
	}
	p.masks[key] = value
}

// HasProtocol is hasProtocol: does the mask this server advertised for that
// update type contain the frequency's bit.
func (p *Protocol) HasProtocol(updateType constant.AdminUpdateType, frequency constant.AdminUpdateFrequency) bool {
	if updateType < 0 || int(updateType) >= len(p.masks) {
		return false
	}
	return p.masks[int(updateType)] >= 0 && (p.masks[int(updateType)]&frequency.Mask()) != 0
}

// FirstSupported is firstSupported: the first frequency the server advertises
// for an update type. ADMIN_FREQUENCY_POLL is advertised only for the types the
// server can actually poll, so this is never POLL for ADMIN_UPDATE_CHAT,
// ADMIN_UPDATE_CONSOLE, ADMIN_UPDATE_CMD_LOGGING or ADMIN_UPDATE_GAMESCRIPT.
//
// Returns ADMIN_FREQUENCY_NONE, the port's stand-in for Java's null, when the
// server advertises none.
func (p *Protocol) FirstSupported(updateType constant.AdminUpdateType) constant.AdminUpdateFrequency {
	if updateType < 0 || int(updateType) >= len(p.masks) || p.masks[int(updateType)] < 0 {
		return constant.ADMIN_FREQUENCY_NONE
	}
	var frequencies = constant.AdminUpdateFrequencyValues()
	for i := 0; i < len(frequencies); i++ {
		if (p.masks[int(updateType)] & frequencies[i].Mask()) != 0 {
			return frequencies[i]
		}
	}
	return constant.ADMIN_FREQUENCY_NONE
}

// ToString is toString: the version header, then one padded line per advertised
// type/frequency pair. Java uses a Formatter with "%-30s - %s%n"; fmt's
// %-30s pads the same way.
func (p *Protocol) ToString() string {
	var text = "Protocol version '" + strconv.Itoa(p.version) + "' :\n"
	var types = constant.AdminUpdateTypeValues()
	var frequencies = constant.AdminUpdateFrequencyValues()
	for i := 0; i < len(types); i++ {
		for j := 0; j < len(frequencies); j++ {
			if p.HasProtocol(types[i], frequencies[j]) {
				text += fmt.Sprintf("%-30s - %s\n", constant.AdminUpdateTypeName(types[i]),
					constant.AdminUpdateFrequencyName(frequencies[j]))
			}
		}
	}
	return text
}
