package com.openttd.network.admin;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.IOException;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import com.google.gson.JsonParser;
import com.google.gson.JsonObject;
import com.openttd.TestUtils;
import com.openttd.constant.OTTD;
import com.openttd.network.constant.GameScript;
import com.openttd.network.constant.GameScript.GSCommand;
import com.openttd.network.constant.GameVersion;
import com.openttd.network.constant.NetworkType.DestType;
import com.openttd.network.constant.NetworkType.NetworkAction;
import com.openttd.network.constant.TcpAdmin.AdminUpdateFrequency;
import com.openttd.network.constant.TcpAdmin.AdminUpdateType;
import com.openttd.network.constant.TcpAdmin.PacketAdminType;
import com.openttd.network.core.Packet;
import com.openttd.server.AdminHarness;
import com.openttd.server.FakeAdminServer;

/**
 * Every ADMIN_PACKET_ADMIN_* command, byte for byte. Payload layouts are copied
 * from git show 15.3:src/network/core/tcp_admin.h ("Receive_ADMIN_*" doc
 * comments); the field length limits come from
 * git show 15.3:src/network/core/config.h.
 */
public class NetworkAdminSenderTest {

    private AdminHarness harness;
    private FakeAdminServer server;
    private NetworkAdminSender send;

    @BeforeEach
    public void setUp() throws IOException {
        this.harness = new AdminHarness();
        this.server = this.harness.getServer();
        this.send = this.harness.getSend();
    }

    @AfterEach
    public void tearDown() {
        this.harness.close();
    }

    /** ADMIN_PACKET_ADMIN_JOIN: string password, string name, string version. */
    @Test
    public void testJoin() {
        FakeAdminServer.ReceivedPacket join = this.server.awaitPacket(PacketAdminType.ADMIN_PACKET_ADMIN_JOIN, 5);
        Packet payload = join.getPacket();
        assertEquals(AdminHarness.PASSWORD, payload.readString());
        assertEquals(AdminHarness.ADMIN_NAME, payload.readString());
        assertEquals(GameVersion.OPENTTD, payload.readString());
        // NETWORK_PASSWORD_LENGTH / NETWORK_CLIENT_NAME_LENGTH / NETWORK_REVISION_LENGTH
        // are 33, 25 and 33 bytes including the terminator.
        assertTrue(join.getPacket().getLength() <= 3 + 33 + 25 + 33, "join packet fits the field limits");
    }

    /** ADMIN_PACKET_ADMIN_QUIT: no payload. */
    @Test
    @DisplayName("quit is sent and has no payload")
    public void testQuit() {
        this.send.quit();
        FakeAdminServer.ReceivedPacket quit = this.server.awaitPacket(PacketAdminType.ADMIN_PACKET_ADMIN_QUIT, 5);
        assertEquals(3, quit.getPacket().getLength(), "header only");
    }

    /** ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY: uint16 update type, uint16 frequency mask. */
    @Test
    public void testUpdateFrequency() {
        this.send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_COMPANY_ECONOMY, AdminUpdateFrequency.ADMIN_FREQUENCY_WEEKLY);
        // COMPANY_ECONOMY advertises Poll|Weekly|..., so Weekly is legal here.
        assertSentFrequency(AdminUpdateType.ADMIN_UPDATE_COMPANY_ECONOMY, AdminUpdateFrequency.ADMIN_FREQUENCY_WEEKLY);

        FakeAdminServer.ReceivedPacket packet = this.server.awaitPacket(PacketAdminType.ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY, 5);
        assertEquals(PacketAdminType.ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY.ordinal(), packet.getWireTypeId(), "command id");
        assertEquals(7, packet.getPacket().getLength(), "header + uint16 type + uint16 frequency");
    }

    /**
     * The server closes the connection on a frequency it does not advertise, so
     * an unsupported request must fall back to an advertised one. POLL is not
     * advertised for ADMIN_UPDATE_CHAT, which used to be the fallback.
     */
    @Test
    @DisplayName("an unsupported frequency falls back to an advertised one, never Poll")
    public void testUpdateFrequencyFallback() {
        this.send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_CHAT, AdminUpdateFrequency.ADMIN_FREQUENCY_DAILY);
        assertSentFrequency(AdminUpdateType.ADMIN_UPDATE_CHAT, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC);

        this.send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_GAMESCRIPT, AdminUpdateFrequency.ADMIN_FREQUENCY_ANUALLY);
        assertSentFrequency(AdminUpdateType.ADMIN_UPDATE_GAMESCRIPT, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC);

        this.send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_DATE, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC);
        assertSentFrequency(AdminUpdateType.ADMIN_UPDATE_DATE, AdminUpdateFrequency.ADMIN_FREQUENCY_POLL);

        this.send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_CMD_NAMES, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC);
        assertSentFrequency(AdminUpdateType.ADMIN_UPDATE_CMD_NAMES, AdminUpdateFrequency.ADMIN_FREQUENCY_POLL);
    }

    private void assertSentFrequency(AdminUpdateType type, AdminUpdateFrequency expected) {
        TestUtils.await(() -> {
            for (FakeAdminServer.ReceivedPacket packet : this.server.packets(PacketAdminType.ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY)) {
                Packet payload = packet.getPacket();
                if (payload.readUint16() == type.ordinal() && payload.readUint16() == expected.mask) {
                    return true;
                }
            }
            return false;
        }, 5, type + " sent with " + expected);
    }

    /** ADMIN_PACKET_ADMIN_POLL: uint8 update type, uint32 id. */
    @Test
    public void testPollDate() {
        this.send.pollDate();
        assertSentPoll(AdminUpdateType.ADMIN_UPDATE_DATE, 0L);
    }

    /** UINT32_MAX asks for every client. */
    @Test
    public void testPollAllClients() {
        this.send.pollClientInfo(-1);
        assertSentPoll(AdminUpdateType.ADMIN_UPDATE_CLIENT_INFO, 4294967295L);
    }

    @Test
    public void testPollOneClient() {
        this.send.pollClientInfo(2);
        assertSentPoll(AdminUpdateType.ADMIN_UPDATE_CLIENT_INFO, 2L);
    }

    /** UINT32_MAX asks for every company. */
    @Test
    public void testPollAllCompanies() {
        this.send.pollCompanyInfo((short) -1);
        assertSentPoll(AdminUpdateType.ADMIN_UPDATE_COMPANY_INFO, 4294967295L);
    }

    @Test
    public void testPollCompanyEconomy() {
        this.send.pollCompanyEconomy((short) 1);
        assertSentPoll(AdminUpdateType.ADMIN_UPDATE_COMPANY_ECONOMY, 1L);
    }

    @Test
    @DisplayName("company stats are pollable, and the helper sends them")
    public void testPollCompanyStats() {
        this.send.pollCompanyStats((short) 0);
        assertSentPoll(AdminUpdateType.ADMIN_UPDATE_COMPANY_STATS, 0L);
    }

    @Test
    @DisplayName("command names are pollable, and the helper sends them")
    public void testPollCmdNames() {
        this.send.pollCmdNames();
        assertSentPoll(AdminUpdateType.ADMIN_UPDATE_CMD_NAMES, 0L);
    }

    /**
     * CHAT, CONSOLE, CMD_LOGGING and GAMESCRIPT have no Poll bit: polling them
     * makes the server answer NETWORK_ERROR_ILLEGAL_PACKET and close the
     * connection, so nothing must be sent.
     */
    @Test
    @DisplayName("polling an unpollable update type sends nothing")
    public void testPollUnsupportedType() {
        for (AdminUpdateType type : new AdminUpdateType[]{
                AdminUpdateType.ADMIN_UPDATE_CHAT,
                AdminUpdateType.ADMIN_UPDATE_CONSOLE,
                AdminUpdateType.ADMIN_UPDATE_CMD_LOGGING,
                AdminUpdateType.ADMIN_UPDATE_GAMESCRIPT,
        }) {
            this.send.poll(type, 0L);
        }
        TestUtils.wait(1);
        for (FakeAdminServer.ReceivedPacket packet : this.server.packets(PacketAdminType.ADMIN_PACKET_ADMIN_POLL)) {
            Packet payload = packet.getPacket();
            short sentType = payload.readUint8();
            assertTrue(sentType < AdminUpdateType.ADMIN_UPDATE_END.ordinal(), "poll type " + sentType);
            assertTrue(sentType != AdminUpdateType.ADMIN_UPDATE_CHAT.ordinal(), "chat is not pollable");
            assertTrue(sentType != AdminUpdateType.ADMIN_UPDATE_CONSOLE.ordinal(), "console is not pollable");
            assertTrue(sentType != AdminUpdateType.ADMIN_UPDATE_CMD_LOGGING.ordinal(), "command logging is not pollable");
            assertTrue(sentType != AdminUpdateType.ADMIN_UPDATE_GAMESCRIPT.ordinal(), "gamescript is not pollable");
        }
    }

    private void assertSentPoll(AdminUpdateType type, long data) {
        TestUtils.await(() -> {
            for (FakeAdminServer.ReceivedPacket packet : this.server.packets(PacketAdminType.ADMIN_PACKET_ADMIN_POLL)) {
                Packet payload = packet.getPacket();
                if (payload.readUint8() == type.ordinal() && payload.readUint32() == data) {
                    return true;
                }
            }
            return false;
        }, 5, "poll " + type + " " + data);
    }

    /** ADMIN_PACKET_ADMIN_CHAT: uint8 action, uint8 desttype, uint32 dest, string. */
    @Test
    public void testChatBroadcast() {
        this.send.chatBroadcast("hello all");
        assertSentChat(NetworkAction.NETWORK_ACTION_SERVER_MESSAGE, DestType.DESTTYPE_BROADCAST, 0L, "hello all");
    }

    @Test
    public void testChatCompany() {
        this.send.chatCompany((short) 3, "team message");
        assertSentChat(NetworkAction.NETWORK_ACTION_SERVER_MESSAGE, DestType.DESTTYPE_TEAM, 3L, "team message");
    }

    @Test
    public void testChatClient() {
        this.send.chatClient(5L, "private message");
        assertSentChat(NetworkAction.NETWORK_ACTION_SERVER_MESSAGE, DestType.DESTTYPE_CLIENT, 5L, "private message");
    }

    /** A negative client id falls back to a broadcast. */
    @Test
    public void testChatClientNegativeFallsBackToBroadcast() {
        this.send.chatClient(-1L, "to everyone");
        assertSentChat(NetworkAction.NETWORK_ACTION_SERVER_MESSAGE, DestType.DESTTYPE_BROADCAST, 0L, "to everyone");
    }

    /** Chat with UTF-8, matching NETWORK_CHAT_LENGTH (900 bytes including '\0'). */
    @Test
    public void testChatUtf8() {
        String message = "こんにちは ☕";
        this.send.chatBroadcast(message);
        assertSentChat(NetworkAction.NETWORK_ACTION_SERVER_MESSAGE, DestType.DESTTYPE_BROADCAST, 0L, message);
    }

    private void assertSentChat(NetworkAction action, DestType destType, long dest, String message) {
        TestUtils.await(() -> {
            for (FakeAdminServer.ReceivedPacket packet : this.server.packets(PacketAdminType.ADMIN_PACKET_ADMIN_CHAT)) {
                Packet payload = packet.getPacket();
                if (payload.readUint8() == action.ordinal() && payload.readUint8() == destType.ordinal()
                        && payload.readUint32() == dest && message.equals(payload.readString())) {
                    return true;
                }
            }
            return false;
        }, 5, "chat " + action + " " + destType + " " + dest + " \"" + message + "\"");
    }

    /**
     * ADMIN_PACKET_ADMIN_EXTERNAL_CHAT: string source, uint16 colour, string
     * user, string message.
     */
    @Test
    public void testExternalChat() {
        this.send.externalChat("strategyboard", (char) 2, "Alice", "a message from outside");
        FakeAdminServer.ReceivedPacket packet = this.server.awaitPacket(PacketAdminType.ADMIN_PACKET_ADMIN_EXTERNAL_CHAT, 5);
        Packet payload = packet.getPacket();
        assertEquals("strategyboard", payload.readString());
        assertEquals((char) 2, payload.readUint16());
        assertEquals("Alice", payload.readString());
        assertEquals("a message from outside", payload.readString());
    }

    /** ADMIN_PACKET_ADMIN_RCON: string command. */
    @Test
    public void testRcon() {
        this.send.rcon("getdate");
        FakeAdminServer.ReceivedPacket packet = this.server.awaitPacket(PacketAdminType.ADMIN_PACKET_ADMIN_RCON, 5);
        assertEquals("getdate", packet.getPacket().readString());
    }

    /**
     * NETWORK_RCONCOMMAND_LENGTH is 500 bytes including '\0': a longer command
     * makes the server reject the packet, so callers must keep commands short.
     */
    @Test
    public void testRconCommandLength() {
        StringBuilder longCommand = new StringBuilder("say ");
        while (longCommand.length() < 500) {
            longCommand.append('x');
        }
        this.send.rcon(longCommand.toString());
        FakeAdminServer.ReceivedPacket packet = this.server.awaitPacket(PacketAdminType.ADMIN_PACKET_ADMIN_RCON, 5);
        assertEquals(longCommand.toString(), packet.getPacket().readString());
        assertTrue(packet.getPacket().getLength() > 3 + 500,
                "the sender does not cap commands at NETWORK_RCONCOMMAND_LENGTH");
    }

    /** ADMIN_PACKET_ADMIN_GAMESCRIPT: string json. */
    @Test
    public void testGameScript() {
        String json = "{\"c\":1,\"a\":[14,\"hello\",\"-1\"],\"i\":\"0\"}";
        assertTrue(this.send.gameScript(json));
        FakeAdminServer.ReceivedPacket packet = this.server.awaitPacket(PacketAdminType.ADMIN_PACKET_ADMIN_GAMESCRIPT, 5);
        assertEquals(json, packet.getPacket().readString());
    }

    @Test
    @DisplayName("gamescript json over COMPAT_MTU is refused and not sent")
    public void testGameScriptTooLong() {
        assertFalse(this.send.gameScript(null));
        StringBuilder json = new StringBuilder("{\"a\":\"");
        while (json.length() <= Packet.MTU - 3) {
            json.append('x');
        }
        json.append("\"}");
        assertFalse(this.send.gameScript(json.toString()));
        TestUtils.wait(1);
        assertTrue(this.server.packets(PacketAdminType.ADMIN_PACKET_ADMIN_GAMESCRIPT).isEmpty());
    }

    /** ADMIN_PACKET_ADMIN_PING: uint32 identifier. */
    @Test
    public void testPing() {
        this.send.ping(4294967295L); // uint32 max
        FakeAdminServer.ReceivedPacket packet = this.server.awaitPacket(PacketAdminType.ADMIN_PACKET_ADMIN_PING, 5);
        assertEquals(4294967295L, packet.getPacket().readUint32());
    }

    /**
     * Goal commands go through GSExecutor, so they carry a request id: AdminCmd
     * only runs requests with "i", "c" and "a".
     */
    @Test
    @DisplayName("goal commands carry a request id")
    public void testAddCompanyGoal() {
        this.send.addCompanyGoal((short) 1, "deliver cargo", OTTD.GoalType.GT_TOWN, 4);
        JsonObject json = awaitGameScriptJson(GSCommand.addGoal);
        assertEquals(GSCommand.addGoal.ordinal(), json.get(GameScript.CMD).getAsInt());
        assertTrue(json.has(GameScript.ID), "AdminCmd requires \"i\"");
        assertEquals("0", json.get(GameScript.ID).getAsString());
        assertEquals(1, json.get(GameScript.ARGS).getAsJsonArray().get(0).getAsInt());
        assertEquals("deliver cargo", json.get(GameScript.ARGS).getAsJsonArray().get(1).getAsString());
        assertEquals(OTTD.GoalType.GT_TOWN.ordinal(), json.get(GameScript.ARGS).getAsJsonArray().get(2).getAsInt());
        assertEquals(4, json.get(GameScript.ARGS).getAsJsonArray().get(3).getAsInt());
    }

    @Test
    public void testAddGlobalGoal() {
        this.send.addGlobalGoal("global goal", OTTD.GoalType.GT_INDUSTRY, 7);
        JsonObject json = awaitGameScriptJson(GSCommand.addGoal);
        assertEquals(-1, json.get(GameScript.ARGS).getAsJsonArray().get(0).getAsInt(), "a global goal is company -1");
        assertEquals(OTTD.GoalType.GT_INDUSTRY.ordinal(), json.get(GameScript.ARGS).getAsJsonArray().get(2).getAsInt());
    }

    @Test
    public void testRemoveGoal() {
        this.send.removeGoal(12);
        JsonObject json = awaitGameScriptJson(GSCommand.removeGoal);
        assertEquals(12, json.get(GameScript.ARGS).getAsJsonArray().get(0).getAsInt());
        assertTrue(json.has(GameScript.ID));
    }

    @Test
    public void testRemoveAllGoal() {
        this.send.removeAllGoal();
        JsonObject json = awaitGameScriptJson(GSCommand.removeAllGoal);
        assertEquals(0, json.get(GameScript.ARGS).getAsJsonArray().size(), "removeAllGoal takes no arguments");
        assertTrue(json.has(GameScript.ID));
    }

    @Test
    public void testSetTownCargoGoal() {
        this.send.setTownCargoGoal(3, OTTD.TownEffect.TE_MAIL, 500);
        JsonObject json = awaitGameScriptJson(GSCommand.setTownCargoGoal);
        assertEquals(3, json.get(GameScript.ARGS).getAsJsonArray().get(0).getAsInt());
        assertEquals(OTTD.TownEffect.TE_MAIL.ordinal(), json.get(GameScript.ARGS).getAsJsonArray().get(1).getAsInt());
        assertEquals(500, json.get(GameScript.ARGS).getAsJsonArray().get(2).getAsInt());
        assertTrue(json.has(GameScript.ID));
    }

    private JsonObject awaitGameScriptJson(GSCommand command) {
        TestUtils.await(() -> {
            for (FakeAdminServer.ReceivedPacket packet : this.server.packets(PacketAdminType.ADMIN_PACKET_ADMIN_GAMESCRIPT)) {
                JsonObject json = parseJson(packet.getPacket().readString());
                if (json.get(GameScript.CMD).getAsInt() == command.ordinal()) {
                    return true;
                }
            }
            return false;
        }, 5, "gamescript command " + command);

        for (FakeAdminServer.ReceivedPacket packet : this.server.packets(PacketAdminType.ADMIN_PACKET_ADMIN_GAMESCRIPT)) {
            JsonObject json = parseJson(packet.getPacket().readString());
            if (json.get(GameScript.CMD).getAsInt() == command.ordinal()) {
                assertNotNull(json.get(GameScript.ARGS));
                return json;
            }
        }
        throw new IllegalStateException("unreachable");
    }

    /**
     * gson 2.2.2 only fills a JsonObject through JsonParser; Gson.fromJson(text,
     * JsonObject.class) returns an empty object (GSExecutor deserialises its
     * response into a POJO, which does work).
     */
    private static JsonObject parseJson(String json) {
        return new JsonParser().parse(json).getAsJsonObject();
    }
}
