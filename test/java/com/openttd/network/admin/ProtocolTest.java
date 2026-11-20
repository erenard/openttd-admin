package com.openttd.network.admin;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.IOException;
import java.util.Map;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import com.openttd.network.constant.TcpAdmin.AdminUpdateFrequency;
import com.openttd.network.constant.TcpAdmin.AdminUpdateType;
import com.openttd.network.constant.TcpAdmin.PacketServerType;
import com.openttd.network.core.Packet;
import com.openttd.server.PacketWriter;
import com.openttd.server.UpstreamAdminProtocol;

/**
 * The PROTOCOL packet, as sent by ServerNetworkAdminSocketHandler::SendProtocol()
 * in OpenTTD 15.3: uint8 version, then for every AdminUpdateType below
 * ADMIN_UPDATE_END a bool plus the uint16 type and uint16 frequency mask, and a
 * final bool false. NetworkAdmin.handlePacket decodes exactly that loop into
 * Protocol, which gates updateFrequency() and poll().
 */
public class ProtocolTest {

    private Protocol protocolFromPacket;

    @Test
    @DisplayName("the PROTOCOL packet decodes into the advertised frequency table")
    public void testDecodeProtocolPacket() throws IOException {
        Packet packet = decode(UpstreamAdminProtocol.protocolPacket());
        Protocol protocol = new Protocol(packet.readUint8());
        while (packet.readBool8()) {
            protocol.putProtocol(packet.readUint16(), packet.readUint16());
        }
        // The loop consumes the false marker, so nothing is left in the packet.
        assertEquals(0, packet.remaining(), "the packet ends with a false marker");
        this.protocolFromPacket = protocol;

        assertEquals(UpstreamAdminProtocol.GAME_ADMIN_VERSION, protocol.getVersion());
        for (Map.Entry<AdminUpdateType, Integer> entry : UpstreamAdminProtocol.FREQUENCIES.entrySet()) {
            assertTrue(protocol.hasProtocol(entry.getKey(), frequencyOf(entry.getValue())),
                    entry.getKey() + " must advertise its first frequency");
            assertEquals(entry.getValue().intValue(), maskOf(protocol, entry.getKey()), entry.getKey() + " mask");
        }
    }

    /** Rebuilds the same table the way NetworkAdmin does, for the other tests. */
    private static Protocol advertisedProtocol() {
        Protocol protocol = new Protocol(UpstreamAdminProtocol.GAME_ADMIN_VERSION);
        for (Map.Entry<AdminUpdateType, Integer> entry : UpstreamAdminProtocol.FREQUENCIES.entrySet()) {
            protocol.putProtocol(entry.getKey().ordinal(), entry.getValue());
        }
        return protocol;
    }

    @Test
    @DisplayName("hasProtocol matches the advertised bits, and nothing else")
    public void testHasProtocolMatrix() {
        Protocol protocol = advertisedProtocol();
        for (AdminUpdateType type : AdminUpdateType.values()) {
            if (type == AdminUpdateType.ADMIN_UPDATE_END) {
                continue;
            }
            int mask = UpstreamAdminProtocol.FREQUENCIES.get(type);
            for (AdminUpdateFrequency frequency : AdminUpdateFrequency.values()) {
                assertEquals((mask & frequency.mask) != 0, protocol.hasProtocol(type, frequency),
                        type + " / " + frequency);
            }
        }
        // ADMIN_UPDATE_END is a sentinel: the server never advertises it.
        assertFalse(protocol.hasProtocol(AdminUpdateType.ADMIN_UPDATE_END, AdminUpdateFrequency.ADMIN_FREQUENCY_POLL));
    }

    /**
     * Only DATE, CLIENT_INFO, COMPANY_INFO, COMPANY_ECONOMY, COMPANY_STATS and
     * CMD_NAMES have a Poll bit; polling CHAT, CONSOLE, CMD_LOGGING or
     * GAMESCRIPT makes the server reply NETWORK_ERROR_ILLEGAL_PACKET.
     */
    @Test
    @DisplayName("Poll is advertised for exactly the six pollable update types")
    public void testPollableTypes() {
        Protocol protocol = advertisedProtocol();
        AdminUpdateType[] pollable = {
                AdminUpdateType.ADMIN_UPDATE_DATE,
                AdminUpdateType.ADMIN_UPDATE_CLIENT_INFO,
                AdminUpdateType.ADMIN_UPDATE_COMPANY_INFO,
                AdminUpdateType.ADMIN_UPDATE_COMPANY_ECONOMY,
                AdminUpdateType.ADMIN_UPDATE_COMPANY_STATS,
                AdminUpdateType.ADMIN_UPDATE_CMD_NAMES,
        };
        AdminUpdateType[] unpollable = {
                AdminUpdateType.ADMIN_UPDATE_CHAT,
                AdminUpdateType.ADMIN_UPDATE_CONSOLE,
                AdminUpdateType.ADMIN_UPDATE_CMD_LOGGING,
                AdminUpdateType.ADMIN_UPDATE_GAMESCRIPT,
        };
        for (AdminUpdateType type : pollable) {
            assertTrue(protocol.hasProtocol(type, AdminUpdateFrequency.ADMIN_FREQUENCY_POLL), type + " is pollable");
        }
        for (AdminUpdateType type : unpollable) {
            assertFalse(protocol.hasProtocol(type, AdminUpdateFrequency.ADMIN_FREQUENCY_POLL), type + " is not pollable");
        }
    }

    /**
     * The server rejects an update frequency it does not know by closing the
     * connection, so the sender must fall back to an advertised frequency - and
     * POLL is not one for chat, console, command logging or gamescript.
     */
    @Test
    @DisplayName("firstSupported never returns Poll for the automatic-only types")
    public void testFirstSupported() {
        Protocol protocol = advertisedProtocol();
        assertEquals(AdminUpdateFrequency.ADMIN_FREQUENCY_POLL, protocol.firstSupported(AdminUpdateType.ADMIN_UPDATE_DATE));
        assertEquals(AdminUpdateFrequency.ADMIN_FREQUENCY_POLL, protocol.firstSupported(AdminUpdateType.ADMIN_UPDATE_CLIENT_INFO));
        assertEquals(AdminUpdateFrequency.ADMIN_FREQUENCY_POLL, protocol.firstSupported(AdminUpdateType.ADMIN_UPDATE_COMPANY_INFO));
        assertEquals(AdminUpdateFrequency.ADMIN_FREQUENCY_POLL, protocol.firstSupported(AdminUpdateType.ADMIN_UPDATE_COMPANY_ECONOMY));
        assertEquals(AdminUpdateFrequency.ADMIN_FREQUENCY_POLL, protocol.firstSupported(AdminUpdateType.ADMIN_UPDATE_COMPANY_STATS));
        assertEquals(AdminUpdateFrequency.ADMIN_FREQUENCY_POLL, protocol.firstSupported(AdminUpdateType.ADMIN_UPDATE_CMD_NAMES));
        assertEquals(AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC, protocol.firstSupported(AdminUpdateType.ADMIN_UPDATE_CHAT));
        assertEquals(AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC, protocol.firstSupported(AdminUpdateType.ADMIN_UPDATE_CONSOLE));
        assertEquals(AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC, protocol.firstSupported(AdminUpdateType.ADMIN_UPDATE_CMD_LOGGING));
        assertEquals(AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC, protocol.firstSupported(AdminUpdateType.ADMIN_UPDATE_GAMESCRIPT));
        assertNull(protocol.firstSupported(AdminUpdateType.ADMIN_UPDATE_END));
    }

    /**
     * A type the server does not advertise at all (a newer OpenTTD adding one)
     * must not be sent.
     */
    @Test
    public void testUnknownUpdateTypeIsNotAdvertised() {
        Protocol protocol = new Protocol(UpstreamAdminProtocol.GAME_ADMIN_VERSION);
        assertNull(protocol.firstSupported(AdminUpdateType.ADMIN_UPDATE_DATE));
        assertFalse(protocol.hasProtocol(AdminUpdateType.ADMIN_UPDATE_DATE, AdminUpdateFrequency.ADMIN_FREQUENCY_DAILY));
    }

    @Test
    @DisplayName("toString lists every advertised type/frequency pair")
    public void testToString() {
        Protocol protocol = advertisedProtocol();
        int advertised = 0;
        for (Map.Entry<AdminUpdateType, Integer> entry : UpstreamAdminProtocol.FREQUENCIES.entrySet()) {
            advertised += Integer.bitCount(entry.getValue());
        }
        String text = protocol.toString();
        assertTrue(text.contains("Protocol version '" + UpstreamAdminProtocol.GAME_ADMIN_VERSION + "'"), text);
        // One header line plus one line per advertised type/frequency pair; the
        // trailing newline is not a line of its own.
        assertEquals(advertised + 1, text.split("\n").length, text);
        // toString pads the update type with "%-30s", so compare whole lines.
        assertTrue(hasLine(text, AdminUpdateType.ADMIN_UPDATE_DATE, AdminUpdateFrequency.ADMIN_FREQUENCY_ANUALLY), text);
        assertFalse(hasLine(text, AdminUpdateType.ADMIN_UPDATE_CHAT, AdminUpdateFrequency.ADMIN_FREQUENCY_DAILY), text);
    }

    private static boolean hasLine(String text, AdminUpdateType type, AdminUpdateFrequency frequency) {
        for (String line : text.split("\n")) {
            String[] parts = line.split(" - ");
            if (parts.length == 2 && parts[0].trim().equals(type.toString())
                    && parts[1].trim().equals(frequency.toString())) {
                return true;
            }
        }
        return false;
    }

    private static int maskOf(Protocol protocol, AdminUpdateType type) {
        int mask = 0;
        for (AdminUpdateFrequency frequency : AdminUpdateFrequency.values()) {
            if (protocol.hasProtocol(type, frequency)) {
                mask |= frequency.mask;
            }
        }
        return mask;
    }

    private static AdminUpdateFrequency frequencyOf(int mask) {
        for (AdminUpdateFrequency frequency : AdminUpdateFrequency.values()) {
            if ((mask & frequency.mask) != 0) {
                return frequency;
            }
        }
        throw new IllegalArgumentException("mask " + mask);
    }

    private static Packet decode(PacketWriter writer) {
        Packet packet = Packet.packetToReceive(java.nio.ByteBuffer.wrap(writer.toBytes()));
        packet.prepareToRead();
        packet.readUint8();
        return packet;
    }

    /** Documents which packet type carries the table, for readability. */
    @Test
    public void testProtocolPacketId() {
        assertEquals(103, new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_PROTOCOL).getWireTypeId());
    }
}
