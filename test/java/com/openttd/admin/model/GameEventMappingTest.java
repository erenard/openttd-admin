package com.openttd.admin.model;

import static org.junit.jupiter.api.Assertions.assertEquals;

import java.io.IOException;
import java.math.BigInteger;
import java.util.EnumMap;
import java.util.Map;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import com.openttd.TestUtils;
import com.openttd.admin.event.ChatEvent;
import com.openttd.admin.event.ClientEvent;
import com.openttd.admin.event.CompanyEvent;
import com.openttd.admin.event.ConsoleEvent;
import com.openttd.admin.event.DateEvent;
import com.openttd.admin.event.GameScriptEvent;
import com.openttd.admin.event.PongEvent;
import com.openttd.admin.event.RConEndEvent;
import com.openttd.admin.event.RConEvent;
import com.openttd.network.constant.NetworkType.NetworkVehicleType;
import com.openttd.network.constant.TcpAdmin.PacketServerType;
import com.openttd.server.AdminHarness;
import com.openttd.server.EventRecorder;
import com.openttd.server.PacketWriter;

/**
 * Game.updateAdmin translates technical packets into functional events. This is
 * the completeness table: every ADMIN_PACKET_SERVER_* id is listed with the
 * events it must produce, so a packet that is decoded but forgotten (the switch
 * had no case for ADMIN_PACKET_SERVER_COMPANY_STATS, and ADMIN_PACKET_SERVER_RCON_END
 * fell through into ADMIN_PACKET_SERVER_CONSOLE) fails here.
 */
public class GameEventMappingTest {

    private AdminHarness harness;
    private EventRecorder recorder;

    @BeforeEach
    public void setUp() throws IOException {
        this.harness = new AdminHarness();
        this.recorder = this.harness.getRecorder();
    }

    @AfterEach
    public void tearDown() {
        this.harness.close();
    }

    /** Packets that end the session are covered by NetworkAdminServerPacketsTest. */
    private static final PacketServerType[] SESSION_ENDING = {
            PacketServerType.ADMIN_PACKET_SERVER_FULL,
            PacketServerType.ADMIN_PACKET_SERVER_BANNED,
            PacketServerType.ADMIN_PACKET_SERVER_SHUTDOWN,
            PacketServerType.ADMIN_PACKET_SERVER_AUTH_REQUEST,
            PacketServerType.ADMIN_PACKET_SERVER_ENABLE_ENCRYPTION,
            PacketServerType.INVALID_ADMIN_PACKET,
    };

    @Test
    @DisplayName("every server packet is accounted for in the mapping table")
    public void testTableCoversEveryPacketType() {
        Map<PacketServerType, String> table = expectedEvents();
        for (PacketServerType type : PacketServerType.values()) {
            assertEquals(true, table.containsKey(type) || isSessionEnding(type),
                    type + " is neither mapped nor session ending");
        }
    }

    /**
     * The events each packet type must produce, as "EventClass x count".
     */
    private static Map<PacketServerType, String> expectedEvents() {
        Map<PacketServerType, String> table = new EnumMap<>(PacketServerType.class);
        table.put(PacketServerType.ADMIN_PACKET_SERVER_ERROR, "none");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_PROTOCOL, "none");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_WELCOME, "none");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_NEWGAME, "none");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_DATE, DateEvent.class.getSimpleName() + " x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_CLIENT_JOIN, ClientEvent.class.getSimpleName() + " CREATE x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_CLIENT_INFO, ClientEvent.class.getSimpleName() + " UPDATE x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_CLIENT_UPDATE, ClientEvent.class.getSimpleName() + " UPDATE x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_CLIENT_QUIT, ClientEvent.class.getSimpleName() + " DELETE x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_CLIENT_ERROR, ClientEvent.class.getSimpleName() + " DELETE x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_COMPANY_NEW, CompanyEvent.class.getSimpleName() + " CREATE x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_COMPANY_INFO, CompanyEvent.class.getSimpleName() + " UPDATE x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_COMPANY_UPDATE, CompanyEvent.class.getSimpleName() + " UPDATE x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_COMPANY_REMOVE, CompanyEvent.class.getSimpleName() + " DELETE x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_COMPANY_ECONOMY, CompanyEvent.class.getSimpleName() + " UPDATE x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_COMPANY_STATS, CompanyEvent.class.getSimpleName() + " UPDATE x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_CHAT, ChatEvent.class.getSimpleName() + " x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_RCON, RConEvent.class.getSimpleName() + " x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_CONSOLE, ConsoleEvent.class.getSimpleName() + " x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_CMD_NAMES, "none");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_CMD_LOGGING_OLD, "none");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_CMD_LOGGING, "none");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_GAMESCRIPT, GameScriptEvent.class.getSimpleName() + " x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_RCON_END, RConEndEvent.class.getSimpleName() + " x1");
        table.put(PacketServerType.ADMIN_PACKET_SERVER_PONG, PongEvent.class.getSimpleName() + " x1");
        return table;
    }

    private static boolean isSessionEnding(PacketServerType type) {
        for (PacketServerType ending : SESSION_ENDING) {
            if (type == ending) {
                return true;
            }
        }
        return false;
    }

    @Test
    @DisplayName("each packet produces exactly the events the table promises")
    public void testEventMapping() throws IOException {
        Map<PacketServerType, String> table = expectedEvents();
        for (Map.Entry<PacketServerType, String> entry : table.entrySet()) {
            recorder.clear();
            harness.getServer().send(minimalPacket(entry.getKey()));
            assertEvents(entry.getKey(), entry.getValue());
        }
    }

    private static final String[] EVENT_CLASSES = {
            ChatEvent.class.getSimpleName(),
            ClientEvent.class.getSimpleName(),
            CompanyEvent.class.getSimpleName(),
            ConsoleEvent.class.getSimpleName(),
            DateEvent.class.getSimpleName(),
            GameScriptEvent.class.getSimpleName(),
            PongEvent.class.getSimpleName(),
            RConEvent.class.getSimpleName(),
            RConEndEvent.class.getSimpleName(),
    };

    private void assertEvents(PacketServerType type, String expected) {
        String eventClass = expected.equals("none") ? null : expected.split(" ")[0];
        int count = expected.equals("none") ? 0 : Integer.parseInt(expected.split(" x")[1]);
        if (count > 0) {
            TestUtils.await(() -> countOf(eventClass) >= count, 5, type + " -> " + expected);
        }

        // Exactly one event class may fire, and exactly `count` times: a packet
        // that dispatches an extra event (RCON_END falling through into CONSOLE)
        // or none (COMPANY_STATS having no case) fails here.
        for (String klass : EVENT_CLASSES) {
            assertEquals(klass.equals(eventClass) ? count : 0, countOf(klass), type + " " + klass + " events");
        }
    }

    private int countOf(String eventClass) {
        switch (eventClass) {
            case "ChatEvent":
                return recorder.getChatEvents().size();
            case "ClientEvent":
                return recorder.getClientEvents().size();
            case "CompanyEvent":
                return recorder.getCompanyEvents().size();
            case "ConsoleEvent":
                return recorder.getConsoleEvents().size();
            case "DateEvent":
                return recorder.getDateEvents().size();
            case "GameScriptEvent":
                return recorder.getGameScriptEvents().size();
            case "PongEvent":
                return recorder.getPongEvents().size();
            case "RConEvent":
                return recorder.getRConEvents().size();
            case "RConEndEvent":
                return recorder.getRConEndEvents().size();
            default:
                throw new IllegalArgumentException(eventClass);
        }
    }

    /** Smallest valid payload for a packet type, per tcp_admin.h. */
    private static PacketWriter minimalPacket(PacketServerType type) {
        PacketWriter packet = new PacketWriter(type);
        switch (type) {
            case ADMIN_PACKET_SERVER_ERROR:
                return packet.writeUint8(0);
            case ADMIN_PACKET_SERVER_NEWGAME:
                return packet; // No payload
            case ADMIN_PACKET_SERVER_PROTOCOL:
                return com.openttd.server.UpstreamAdminProtocol.protocolPacket();
            case ADMIN_PACKET_SERVER_WELCOME:
                return com.openttd.server.UpstreamAdminProtocol.welcomePacket();
            case ADMIN_PACKET_SERVER_DATE:
                return packet.writeUint32(1950 * 365 + 1);
            case ADMIN_PACKET_SERVER_CLIENT_JOIN:
            case ADMIN_PACKET_SERVER_CLIENT_QUIT:
                return packet.writeUint32(7L);
            case ADMIN_PACKET_SERVER_CLIENT_INFO:
                return packet.writeUint32(7L).writeString("::1").writeString("Client").writeUint8(0).writeUint32(1L).writeUint8(255);
            case ADMIN_PACKET_SERVER_CLIENT_UPDATE:
                return packet.writeUint32(7L).writeString("Client").writeUint8(255);
            case ADMIN_PACKET_SERVER_CLIENT_ERROR:
                return packet.writeUint32(7L).writeUint8(1);
            case ADMIN_PACKET_SERVER_COMPANY_NEW:
                return packet.writeUint8(6);
            case ADMIN_PACKET_SERVER_COMPANY_INFO:
                return packet.writeUint8(6).writeString("Company").writeString("Manager").writeUint8(1)
                        .writeBool8(false).writeUint32(1950L).writeBool8(false).writeUint8(0)
                        .writeUint8(255).writeUint8(255).writeUint8(255).writeUint8(255);
            case ADMIN_PACKET_SERVER_COMPANY_UPDATE:
                return packet.writeUint8(6).writeString("Company").writeString("Manager").writeUint8(1)
                        .writeBool8(false).writeUint8(0)
                        .writeUint8(255).writeUint8(255).writeUint8(255).writeUint8(255);
            case ADMIN_PACKET_SERVER_COMPANY_REMOVE:
                return packet.writeUint8(6).writeUint8(0);
            case ADMIN_PACKET_SERVER_COMPANY_ECONOMY:
                return packet.writeUint8(6).writeInt64(0).writeUint64(BigInteger.ZERO).writeInt64(0).writeUint16(0)
                        .writeUint64(BigInteger.ZERO).writeUint16(0).writeUint16(0)
                        .writeUint64(BigInteger.ZERO).writeUint16(0).writeUint16(0);
            case ADMIN_PACKET_SERVER_COMPANY_STATS: {
                packet.writeUint8(6);
                for (int i = 0; i < NetworkVehicleType.length(); i++) {
                    packet.writeUint16(0);
                }
                for (int i = 0; i < NetworkVehicleType.length(); i++) {
                    packet.writeUint16(0);
                }
                return packet;
            }
            case ADMIN_PACKET_SERVER_CHAT:
                return packet.writeUint8(3).writeUint8(0).writeUint32(1L).writeString("chat").writeUint64(BigInteger.ZERO);
            case ADMIN_PACKET_SERVER_RCON:
                return packet.writeUint16(2).writeString("output");
            case ADMIN_PACKET_SERVER_CONSOLE:
                return packet.writeString("console").writeString("text");
            case ADMIN_PACKET_SERVER_CMD_NAMES:
                return packet.writeBool8(false);
            case ADMIN_PACKET_SERVER_CMD_LOGGING_OLD:
                return packet.writeUint32(1L).writeUint8(0).writeUint16(1).writeUint32(0).writeUint32(0)
                        .writeUint32(0).writeString("text").writeUint32(0);
            case ADMIN_PACKET_SERVER_CMD_LOGGING:
                return packet.writeUint32(1L).writeUint8(0).writeUint16(1).writeBytes(new byte[]{1}).writeUint32(0);
            case ADMIN_PACKET_SERVER_GAMESCRIPT:
                return packet.writeString("{\"id\":\"0\",\"data\":\"1\"}");
            case ADMIN_PACKET_SERVER_RCON_END:
                return packet.writeString("getdate");
            case ADMIN_PACKET_SERVER_PONG:
                return packet.writeUint32(1L);
            default:
                throw new IllegalArgumentException(type + " is session ending or has no payload");
        }
    }

    /**
     * A packet that closes the session must not dispatch a functional event.
     */
    @Test
    @DisplayName("session-ending packets dispatch nothing")
    public void testSessionEndingPacketsDispatchNothing() throws IOException {
        for (PacketServerType type : SESSION_ENDING) {
            if (type == PacketServerType.INVALID_ADMIN_PACKET) {
                continue; // Not a packet id a server sends
            }
            recorder.clear();
            harness.getServer().send(new PacketWriter(type));
            harness.getServer().awaitClientDisconnect(5);
            assertEquals(0, totalEvents(), type + " must dispatch no event");
        }
    }

    private int totalEvents() {
        return recorder.getChatEvents().size() + recorder.getClientEvents().size() + recorder.getCompanyEvents().size()
                + recorder.getConsoleEvents().size() + recorder.getDateEvents().size()
                + recorder.getGameScriptEvents().size() + recorder.getPongEvents().size()
                + recorder.getRConEvents().size() + recorder.getRConEndEvents().size();
    }

    /** WELCOME only registers updates, it is not a game event. */
    @Test
    @DisplayName("WELCOME dispatches no event")
    public void testWelcomeIsNotAnEvent() throws IOException {
        recorder.clear();
        harness.getServer().send(com.openttd.server.UpstreamAdminProtocol.welcomePacket());
        TestUtils.wait(1);
        assertEquals(0, totalEvents(), "WELCOME is protocol handshake, not game state");
    }
}
