// The PROTOCOL packet, as sent by ServerNetworkAdminSocketHandler::SendProtocol()
// in OpenTTD 15.3: uint8 version, then for every AdminUpdateType below
// ADMIN_UPDATE_END a bool plus the uint16 type and uint16 frequency mask, and a
// final bool false. NetworkAdmin.handlePacket decodes exactly that loop into
// Protocol, which gates updateFrequency() and poll().
//
// See test/java/com/openttd/network/admin/ProtocolTest.java.

package admin_test

import "strings"
import "strconv"
import "testing"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/network/admin"
import "com.openttd/server"

// advertisedProtocol rebuilds the same table the way NetworkAdmin does, for the
// other tests.
func advertisedProtocol() *admin.Protocol {
	var protocol = admin.New(server.GAME_ADMIN_VERSION)
	var types = server.FrequencyTypes()
	var masks = server.Frequencies()
	for i := 0; i < len(types); i++ {
		protocol.PutProtocol(int(types[i]), masks[i])
	}
	return protocol
}

// decode is decode: the payload of a scripted packet, positioned after the
// header and its packet type.
func decode(writer *server.PacketWriter) *core.Packet {
	var packet = core.NewForReceive(writer.ToBytes())
	packet.PrepareToRead()
	packet.ReadUint8()
	return packet
}

// maskOf is maskOf: the mask Protocol reports for an update type, rebuilt from
// the individual bits.
func maskOf(protocol *admin.Protocol, updateType constant.AdminUpdateType) int {
	var mask = 0
	var frequencies = constant.AdminUpdateFrequencyValues()
	for i := 0; i < len(frequencies); i++ {
		if protocol.HasProtocol(updateType, frequencies[i]) {
			mask |= frequencies[i].Mask()
		}
	}
	return mask
}

// frequencyOf is frequencyOf: the first frequency a mask contains.
func frequencyOf(mask int) constant.AdminUpdateFrequency {
	var frequencies = constant.AdminUpdateFrequencyValues()
	for i := 0; i < len(frequencies); i++ {
		if (mask & frequencies[i].Mask()) != 0 {
			return frequencies[i]
		}
	}
	panic("mask " + strconv.Itoa(mask))
}

// splitLines is Java's String.split("\n"), which drops the empty element a
// trailing newline would otherwise produce.
func splitLines(text string) []string {
	var lines = strings.Split(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		return lines[0 : len(lines)-1]
	}
	return lines
}

// hasLine is hasLine: toString pads the update type with "%-30s", so compare
// whole lines.
func hasLine(text string, updateType constant.AdminUpdateType, frequency constant.AdminUpdateFrequency) bool {
	var lines = splitLines(text)
	for i := 0; i < len(lines); i++ {
		var parts = strings.Split(lines[i], " - ")
		if len(parts) == 2 && strings.TrimSpace(parts[0]) == constant.AdminUpdateTypeName(updateType) &&
			strings.TrimSpace(parts[1]) == constant.AdminUpdateFrequencyName(frequency) {
			return true
		}
	}
	return false
}

// TestDecodeProtocolPacket is testDecodeProtocolPacket: the PROTOCOL packet
// decoding into the advertised frequency table.
func TestDecodeProtocolPacket(t *testing.T) {
	var packet = decode(server.ProtocolPacket())
	var protocol = admin.New(packet.ReadUint8())
	for packet.ReadBool8() {
		protocol.PutProtocol(packet.ReadUint16(), packet.ReadUint16())
	}
	// The loop consumes the false marker, so nothing is left in the packet.
	if packet.Remaining() != 0 {
		t.Errorf("expected the packet to end with a false marker, %d bytes left", packet.Remaining())
	}
	if protocol.GetVersion() != server.GAME_ADMIN_VERSION {
		t.Errorf("expected protocol version %d, got %d", server.GAME_ADMIN_VERSION, protocol.GetVersion())
	}

	var types = server.FrequencyTypes()
	var masks = server.Frequencies()
	for i := 0; i < len(types); i++ {
		if !protocol.HasProtocol(types[i], frequencyOf(masks[i])) {
			t.Errorf("%s must advertise its first frequency", constant.AdminUpdateTypeName(types[i]))
		}
		if maskOf(protocol, types[i]) != masks[i] {
			t.Errorf("%s mask: expected %d, got %d", constant.AdminUpdateTypeName(types[i]),
				masks[i], maskOf(protocol, types[i]))
		}
	}
}

// TestHasProtocolMatrix is testHasProtocolMatrix: hasProtocol matching the
// advertised bits, and nothing else.
func TestHasProtocolMatrix(t *testing.T) {
	var protocol = advertisedProtocol()
	var types = constant.AdminUpdateTypeValues()
	var frequencies = constant.AdminUpdateFrequencyValues()
	for i := 0; i < len(types); i++ {
		if types[i] == constant.ADMIN_UPDATE_END {
			continue
		}
		var mask = server.Frequency(types[i])
		for j := 0; j < len(frequencies); j++ {
			var expected = (mask & frequencies[j].Mask()) != 0
			var actual = protocol.HasProtocol(types[i], frequencies[j])
			if expected != actual {
				t.Errorf("%s / %s: expected %v, got %v",
					constant.AdminUpdateTypeName(types[i]),
					constant.AdminUpdateFrequencyName(frequencies[j]), expected, actual)
			}
		}
	}
	// ADMIN_UPDATE_END is a sentinel: the server never advertises it.
	if protocol.HasProtocol(constant.ADMIN_UPDATE_END, constant.ADMIN_FREQUENCY_POLL) {
		t.Errorf("ADMIN_UPDATE_END must not be advertised")
	}
}

// TestPollableTypes: only DATE, CLIENT_INFO, COMPANY_INFO, COMPANY_ECONOMY,
// COMPANY_STATS and CMD_NAMES have a Poll bit; polling CHAT, CONSOLE, CMD_LOGGING
// or GAMESCRIPT makes the server reply NETWORK_ERROR_ILLEGAL_PACKET.
func TestPollableTypes(t *testing.T) {
	var protocol = advertisedProtocol()
	var pollable = []constant.AdminUpdateType{
		constant.ADMIN_UPDATE_DATE,
		constant.ADMIN_UPDATE_CLIENT_INFO,
		constant.ADMIN_UPDATE_COMPANY_INFO,
		constant.ADMIN_UPDATE_COMPANY_ECONOMY,
		constant.ADMIN_UPDATE_COMPANY_STATS,
		constant.ADMIN_UPDATE_CMD_NAMES,
	}
	var unpollable = []constant.AdminUpdateType{
		constant.ADMIN_UPDATE_CHAT,
		constant.ADMIN_UPDATE_CONSOLE,
		constant.ADMIN_UPDATE_CMD_LOGGING,
		constant.ADMIN_UPDATE_GAMESCRIPT,
	}
	for i := 0; i < len(pollable); i++ {
		if !protocol.HasProtocol(pollable[i], constant.ADMIN_FREQUENCY_POLL) {
			t.Errorf("%s should be pollable", constant.AdminUpdateTypeName(pollable[i]))
		}
	}
	for i := 0; i < len(unpollable); i++ {
		if protocol.HasProtocol(unpollable[i], constant.ADMIN_FREQUENCY_POLL) {
			t.Errorf("%s should not be pollable", constant.AdminUpdateTypeName(unpollable[i]))
		}
	}
}

// TestFirstSupported: the server rejects an update frequency it does not know by
// closing the connection, so the sender must fall back to an advertised frequency
// - and POLL is not one for chat, console, command logging or gamescript.
func TestFirstSupported(t *testing.T) {
	var protocol = advertisedProtocol()
	var pollable = []constant.AdminUpdateType{
		constant.ADMIN_UPDATE_DATE,
		constant.ADMIN_UPDATE_CLIENT_INFO,
		constant.ADMIN_UPDATE_COMPANY_INFO,
		constant.ADMIN_UPDATE_COMPANY_ECONOMY,
		constant.ADMIN_UPDATE_COMPANY_STATS,
		constant.ADMIN_UPDATE_CMD_NAMES,
	}
	var automatic = []constant.AdminUpdateType{
		constant.ADMIN_UPDATE_CHAT,
		constant.ADMIN_UPDATE_CONSOLE,
		constant.ADMIN_UPDATE_CMD_LOGGING,
		constant.ADMIN_UPDATE_GAMESCRIPT,
	}
	for i := 0; i < len(pollable); i++ {
		if protocol.FirstSupported(pollable[i]) != constant.ADMIN_FREQUENCY_POLL {
			t.Errorf("%s: expected Poll, got %s", constant.AdminUpdateTypeName(pollable[i]),
				constant.AdminUpdateFrequencyName(protocol.FirstSupported(pollable[i])))
		}
	}
	for i := 0; i < len(automatic); i++ {
		if protocol.FirstSupported(automatic[i]) != constant.ADMIN_FREQUENCY_AUTOMATIC {
			t.Errorf("%s: expected Automatic, got %s", constant.AdminUpdateTypeName(automatic[i]),
				constant.AdminUpdateFrequencyName(protocol.FirstSupported(automatic[i])))
		}
	}
	// ADMIN_UPDATE_END is never advertised, so there is nothing to fall back on.
	if protocol.FirstSupported(constant.ADMIN_UPDATE_END) != constant.ADMIN_FREQUENCY_NONE {
		t.Errorf("expected no frequency for ADMIN_UPDATE_END")
	}
}

// TestUnknownUpdateTypeIsNotAdvertised: a type the server does not advertise at
// all (a newer OpenTTD adding one) must not be sent.
func TestUnknownUpdateTypeIsNotAdvertised(t *testing.T) {
	var protocol = admin.New(server.GAME_ADMIN_VERSION)
	if protocol.FirstSupported(constant.ADMIN_UPDATE_DATE) != constant.ADMIN_FREQUENCY_NONE {
		t.Errorf("expected no frequency before the server advertised any")
	}
	if protocol.HasProtocol(constant.ADMIN_UPDATE_DATE, constant.ADMIN_FREQUENCY_DAILY) {
		t.Errorf("expected an unadvertised type to match no frequency")
	}
}

// TestToString lists every advertised type/frequency pair.
func TestToString(t *testing.T) {
	var protocol = advertisedProtocol()
	var advertised = 0
	var masks = server.Frequencies()
	for i := 0; i < len(masks); i++ {
		advertised += bitCount(masks[i])
	}

	var text = protocol.ToString()
	if !strings.Contains(text, "Protocol version '"+strconv.Itoa(server.GAME_ADMIN_VERSION)+"'") {
		t.Errorf("expected the version header in:\n%s", text)
	}
	// One header line plus one line per advertised type/frequency pair; the
	// trailing newline is not a line of its own.
	if len(splitLines(text)) != advertised+1 {
		t.Errorf("expected %d lines, got %d:\n%s", advertised+1, len(splitLines(text)), text)
	}
	if !hasLine(text, constant.ADMIN_UPDATE_DATE, constant.ADMIN_FREQUENCY_ANUALLY) {
		t.Errorf("expected the DATE/ANUALLY line in:\n%s", text)
	}
	if hasLine(text, constant.ADMIN_UPDATE_CHAT, constant.ADMIN_FREQUENCY_DAILY) {
		t.Errorf("unexpected CHAT/DAILY line in:\n%s", text)
	}
}

// TestProtocolPacketId documents which packet type carries the table.
func TestProtocolPacketId(t *testing.T) {
	var wireTypeId = server.New(constant.ADMIN_PACKET_SERVER_PROTOCOL).GetWireTypeId()
	if wireTypeId != 103 {
		t.Errorf("expected wire id 103, got %d", wireTypeId)
	}
}

// bitCount is Integer.bitCount.
func bitCount(value int) int {
	var count = 0
	var bits = value
	for bits != 0 {
		bits = bits & (bits - 1)
		count += 1
	}
	return count
}
