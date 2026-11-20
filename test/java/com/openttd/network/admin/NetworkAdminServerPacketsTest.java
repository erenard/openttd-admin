package com.openttd.network.admin;

import static org.junit.jupiter.api.Assertions.assertArrayEquals;
import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.IOException;
import java.math.BigInteger;
import java.util.Calendar;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import com.openttd.TestUtils;
import com.openttd.admin.event.ClientEvent;
import com.openttd.admin.event.CompanyEvent;
import com.openttd.admin.event.ConsoleEvent;
import com.openttd.admin.event.GameScriptEvent;
import com.openttd.admin.event.PongEvent;
import com.openttd.admin.event.RConEndEvent;
import com.openttd.admin.event.RConEvent;
import com.openttd.network.constant.NetworkType.NetworkVehicleType;
import com.openttd.network.constant.TcpAdmin.PacketServerType;
import com.openttd.server.AdminHarness;
import com.openttd.server.EventRecorder;
import com.openttd.server.PacketWriter;
import com.openttd.server.UpstreamAdminProtocol;
import com.openttd.server.UpstreamCalendar;

/**
 * Every ADMIN_PACKET_SERVER_* packet the admin network can receive, decoded
 * against FakeAdminServer. Field layouts are copied from
 * git show 15.3:src/network/core/tcp_admin.h ("Receive_SERVER_*" doc comments)
 * and the Send_* functions in git show 15.3:src/network/network_admin.cpp.
 *
 * Packets a real server never sends to a plain ADMIN_PACKET_ADMIN_JOIN client
 * (FULL, BANNED, CMD_LOGGING, AUTH_REQUEST, ENABLE_ENCRYPTION) are covered here
 * too, which is why this suite is offline.
 */
public class NetworkAdminServerPacketsTest {

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

    private void send(PacketWriter packet) throws IOException {
        this.harness.getServer().send(packet);
    }

    /** Prove the connection survived, and that the reply round trips. */
    private void assertStillConnected() throws IOException {
        this.harness.getSend().ping(9L);
        TestUtils.awaitSize(this.recorder.getPongEvents(), 1, 5, "a PONG");
        assertEquals(9L, this.recorder.getPongEvents().get(0).getPingId());
    }

    /**
     * ADMIN_PACKET_SERVER_WELCOME: string server name, string revision, bool
     * dedicated, string (used to be map name), uint32 seed, uint8 landscape,
     * uint32 start date, uint16 width, uint16 height.
     */
    @Test
    public void testWelcome() {
        GameInfo gameInfo = this.harness.getGame().getGameInfo();
        // PROTOCOL comes before WELCOME; WELCOME must not lose its version.
        assertEquals(UpstreamAdminProtocol.GAME_ADMIN_VERSION, gameInfo.getVersion(), "protocol version");
        assertEquals("Strategyboard test server", gameInfo.getServerName());
        assertEquals("15.3", gameInfo.getServerRevision());
        assertTrue(gameInfo.isServerDedicated());
        assertEquals(123456789L, gameInfo.getMapSeed());
        assertEquals(1, gameInfo.getMapSet());
        assertEquals(1950 * 365 + 1, gameInfo.getStartDate());
        assertEquals(64, gameInfo.getMapWidth());
        assertEquals(64, gameInfo.getMapHeight());
    }

    /**
     * Welcome is what makes Game.updateAdmin register updates: eight update
     * frequencies and four polls, all of them advertised by the server.
     */
    @Test
    @DisplayName("WELCOME registers the update frequencies and polls the game asks for")
    public void testWelcomeRegistersUpdates() {
        TestUtils.await(() -> this.harness.getServer().packets(
                com.openttd.network.constant.TcpAdmin.PacketAdminType.ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY).size() >= 8,
                5, "the eight update frequency registrations");
        TestUtils.await(() -> this.harness.getServer().packets(
                com.openttd.network.constant.TcpAdmin.PacketAdminType.ADMIN_PACKET_ADMIN_POLL).size() >= 4,
                5, "the four polls");
    }

    /** ADMIN_PACKET_SERVER_DATE: uint32 current game date. */
    @Test
    public void testDate() throws IOException {
        long date = 1950 * 365 + 100;
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_DATE).writeUint32(date));
        TestUtils.awaitSize(this.recorder.getDateEvents(), 1, 5, "a DateEvent");

        assertEquals(date, this.harness.getGame().getGameInfo().getCurrentDate());
        // OpenTTD counts year 0 and its leap years, so date != year * 365.
        assertEquals(UpstreamCalendar.yearOf((int) date), this.harness.getGame().getDate().get(Calendar.YEAR));
    }

    /** ADMIN_PACKET_SERVER_CLIENT_JOIN: uint32 id of the new client. */
    @Test
    public void testClientJoin() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CLIENT_JOIN).writeUint32(2L));
        TestUtils.awaitSize(this.recorder.getClientEvents(), 1, 5, "a ClientEvent");
        assertEquals(ClientEvent.Action.CREATE, this.recorder.getClientEvents().get(0).getAction());
        assertEquals(2, this.recorder.getClientEvents().get(0).getClientId());
    }

    /**
     * ADMIN_PACKET_SERVER_CLIENT_INFO: uint32 id, string address, string name,
     * uint8 language, uint32 join date, uint8 company id (255 for spectators).
     */
    @Test
    public void testClientInfo() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CLIENT_INFO)
                .writeUint32(2L)
                .writeString("192.168.1.20")
                .writeString("Player Two")
                .writeUint8(0) // Language: always 0 upstream
                .writeUint32(1950 * 365 + 5)
                .writeUint8(1));
        TestUtils.awaitSize(this.recorder.getClientEvents(), 1, 5, "a ClientEvent");

        Client client = this.harness.getGame().getClient(2);
        assertEquals("192.168.1.20", client.getIp());
        assertEquals("Player Two", client.getName());
        assertEquals(1950 * 365 + 5, client.getJoinDate());
        assertEquals(1, client.getCompanyId());
        assertEquals(ClientEvent.Action.UPDATE, this.recorder.getClientEvents().get(0).getAction());
        assertEquals(1, this.harness.getGame().getGameInfo().clientsCount);
    }

    /** ADMIN_PACKET_SERVER_CLIENT_UPDATE: uint32 id, string name, uint8 company id. */
    @Test
    public void testClientUpdate() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CLIENT_INFO)
                .writeUint32(3L).writeString("::1").writeString("Old name").writeUint8(0).writeUint32(1L).writeUint8(0));
        TestUtils.awaitSize(this.recorder.getClientEvents(), 1, 5, "the client info");

        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CLIENT_UPDATE)
                .writeUint32(3L).writeString("Renamed").writeUint8(2));
        TestUtils.awaitSize(this.recorder.getClientEvents(), 2, 5, "the client update");

        Client client = this.harness.getGame().getClient(3);
        assertEquals("Renamed", client.getName());
        assertEquals(2, client.getCompanyId());
        assertEquals("::1", client.getIp(), "an update does not clear the address");
        assertEquals(ClientEvent.Action.UPDATE, this.recorder.getClientEvents().get(1).getAction());
    }

    /** ADMIN_PACKET_SERVER_CLIENT_QUIT: uint32 id of the client that left. */
    @Test
    public void testClientQuit() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CLIENT_INFO)
                .writeUint32(4L).writeString("::1").writeString("Leaver").writeUint8(0).writeUint32(1L).writeUint8(0));
        TestUtils.awaitSize(this.recorder.getClientEvents(), 1, 5, "the client info");

        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CLIENT_QUIT).writeUint32(4L));
        TestUtils.awaitSize(this.recorder.getClientEvents(), 2, 5, "the quit event");

        assertEquals(ClientEvent.Action.DELETE, this.recorder.getClientEvents().get(1).getAction());
        assertEquals(0, this.harness.getGame().getClients().size());
        assertEquals(0, this.harness.getGame().getGameInfo().clientsCount);
    }

    /**
     * ADMIN_PACKET_SERVER_CLIENT_ERROR: uint32 id, uint8 error
     * (NetworkErrorCode). The client is dropped like on a quit.
     */
    @Test
    public void testClientError() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CLIENT_INFO)
                .writeUint32(5L).writeString("::1").writeString("Cheater").writeUint8(0).writeUint32(1L).writeUint8(0));
        TestUtils.awaitSize(this.recorder.getClientEvents(), 1, 5, "the client info");

        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CLIENT_ERROR)
                .writeUint32(5L).writeUint8(13)); // NETWORK_ERROR_CHEATER
        TestUtils.awaitSize(this.recorder.getClientEvents(), 2, 5, "the error event");

        assertEquals(ClientEvent.Action.DELETE, this.recorder.getClientEvents().get(1).getAction());
        assertEquals(0, this.harness.getGame().getClients().size());
        assertStillConnected();
    }

    /** ADMIN_PACKET_SERVER_COMPANY_NEW: uint8 id of the new company. */
    @Test
    public void testCompanyNew() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_COMPANY_NEW).writeUint8(3));
        TestUtils.awaitSize(this.recorder.getCompanyEvents(), 1, 5, "a CompanyEvent");
        assertEquals(CompanyEvent.Action.CREATE, this.recorder.getCompanyEvents().get(0).getAction());
        assertEquals(3, this.recorder.getCompanyEvents().get(0).getCompanyId());
    }

    /**
     * ADMIN_PACKET_SERVER_COMPANY_INFO: uint8 id, string name, string manager,
     * uint8 colour, bool password, uint32 inaugural year, bool ai, uint8
     * bankruptcy quarters, 4 * uint8 share owners.
     */
    @Test
    public void testCompanyInfo() throws IOException {
        send(companyInfo(1));
        TestUtils.awaitSize(this.recorder.getCompanyEvents(), 1, 5, "a CompanyEvent");

        Company company = this.harness.getGame().getCompany((short) 1);
        assertEquals("Test Company", company.getName());
        assertEquals("Test Manager", company.getPresident());
        assertEquals(2, company.getColor());
        assertTrue(company.isUsePassword());
        assertEquals(1950, company.getInauguratedYear());
        assertFalse(company.isAi());
        assertEquals(0, company.getBankruptcy());
        assertArrayEquals(new short[]{0, 255, 255, 255}, company.getShareOwners());
        assertEquals(CompanyEvent.Action.UPDATE, this.recorder.getCompanyEvents().get(0).getAction());
        assertEquals(1, this.harness.getGame().getGameInfo().companiesCount);
    }

    /**
     * ADMIN_PACKET_SERVER_COMPANY_UPDATE: same as COMPANY_INFO without the
     * inaugural year and the ai flag.
     */
    @Test
    public void testCompanyUpdate() throws IOException {
        send(companyInfo(2));
        TestUtils.awaitSize(this.recorder.getCompanyEvents(), 1, 5, "the company info");

        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_COMPANY_UPDATE)
                .writeUint8(2)
                .writeString("Renamed Company")
                .writeString("New Manager")
                .writeUint8(7)
                .writeBool8(false)
                .writeUint8(3) // Quarters of bankruptcy
                .writeUint8(1).writeUint8(2).writeUint8(3).writeUint8(4));
        TestUtils.awaitSize(this.recorder.getCompanyEvents(), 2, 5, "the company update");

        Company company = this.harness.getGame().getCompany((short) 2);
        assertEquals("Renamed Company", company.getName());
        assertEquals("New Manager", company.getPresident());
        assertEquals(7, company.getColor());
        assertFalse(company.isUsePassword());
        assertEquals(3, company.getBankruptcy());
        assertArrayEquals(new short[]{1, 2, 3, 4}, company.getShareOwners());
        assertEquals(1950, company.getInauguratedYear(), "an update keeps the inaugural year");
        assertFalse(company.isAi(), "an update keeps the ai flag");
    }

    /**
     * ADMIN_PACKET_SERVER_COMPANY_REMOVE: uint8 id, uint8 reason
     * (AdminCompanyRemoveReason).
     */
    @Test
    @DisplayName("a removed company is dropped from the model, whatever the reason")
    public void testCompanyRemove() throws IOException {
        send(companyInfo(3));
        TestUtils.awaitSize(this.recorder.getCompanyEvents(), 1, 5, "the company info");

        for (short reason : new short[]{0, 1, 2}) { // MANUAL, AUTOCLEAN, BANKRUPT
            send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_COMPANY_REMOVE).writeUint8(3).writeUint8(reason));
        }
        TestUtils.awaitSize(this.recorder.getCompanyEvents(), 3, 5, "three remove events");

        assertEquals(CompanyEvent.Action.DELETE, this.recorder.getCompanyEvents().get(2).getAction());
        assertEquals(0, this.harness.getGame().getCompanies().size());
        assertStillConnected();
    }

    /**
     * ADMIN_PACKET_SERVER_COMPANY_ECONOMY: uint8 id, int64 money, uint64 loan,
     * int64 income, uint16 delivered cargo, uint64 value (last quarter), uint16
     * performance, uint16 delivered cargo, uint64 value (previous quarter),
     * uint16 performance, uint16 delivered cargo.
     */
    @Test
    @DisplayName("economy values survive the uint64 / int64 boundary")
    public void testCompanyEconomy() throws IOException {
        BigInteger loan = new BigInteger("50000000000"); // above 2^32
        BigInteger lastValue = new BigInteger("200000000000");
        BigInteger previousValue = BigInteger.valueOf(1_000_000L);

        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_COMPANY_ECONOMY)
                .writeUint8(4)
                .writeInt64(-123456L) // Money can be negative
                .writeUint64(loan)
                .writeInt64(987654321L)
                .writeUint16(1200)
                .writeUint64(lastValue)
                .writeUint16(10000)
                .writeUint16(500)
                .writeUint64(previousValue)
                .writeUint16(9000)
                .writeUint16(400));
        TestUtils.awaitSize(this.recorder.getCompanyEvents(), 1, 5, "a CompanyEvent");

        Company company = this.harness.getGame().getCompany((short) 4);
        assertEquals(-123456L, company.getMoney());
        assertEquals(loan, company.getLoan());
        assertEquals(987654321L, company.getIncome());
        assertEquals(1200, company.getDeliveredCargo());
        assertEquals(lastValue, company.getLastValue());
        assertEquals(10000, company.getLastPerformance());
        assertEquals(500, company.getLastDeliveredCargo());
        assertEquals(previousValue, company.getPreviousValue());
        assertEquals(9000, company.getPreviousPerformance());
        assertEquals(400, company.getPreviousDeliveredCargo());
    }

    /**
     * ADMIN_PACKET_SERVER_COMPANY_STATS: uint8 id, then 5 vehicle counts and 5
     * station counts in NetworkVehicleType order.
     */
    @Test
    @DisplayName("company stats are decoded and dispatched")
    public void testCompanyStats() throws IOException {
        PacketWriter packet = new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_COMPANY_STATS).writeUint8(5);
        for (int vehicleType = 0; vehicleType < NetworkVehicleType.length(); vehicleType++) {
            packet.writeUint16(vehicleType + 10);
        }
        for (int stationType = 0; stationType < NetworkVehicleType.length(); stationType++) {
            packet.writeUint16(stationType + 20);
        }
        send(packet);
        TestUtils.awaitSize(this.recorder.getCompanyEvents(), 1, 5, "a CompanyEvent");

        Company company = this.harness.getGame().getCompany((short) 5);
        assertArrayEquals(new char[]{10, 11, 12, 13, 14}, company.getVehicules());
        assertArrayEquals(new char[]{20, 21, 22, 23, 24}, company.getStations());
        assertEquals(CompanyEvent.Action.UPDATE, this.recorder.getCompanyEvents().get(0).getAction(),
                "the server gives the admin statistics, the game must report them");
    }

    /**
     * ADMIN_PACKET_SERVER_CHAT: uint8 action, uint8 desttype, uint32 client id,
     * string message, uint64 money (only for a 'give money' action).
     */
    @Test
    @DisplayName("chat is relayed, UTF-8 intact")
    public void testChat() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CHAT)
                .writeUint8(3) // NETWORK_ACTION_CHAT
                .writeUint8(0) // DESTTYPE_BROADCAST
                .writeUint32(2L)
                .writeString("Grüße, player two")
                .writeUint64(BigInteger.ZERO));
        TestUtils.awaitSize(this.recorder.getChatEvents(), 1, 5, "a ChatEvent");

        assertEquals(2, this.recorder.getChatEvents().get(0).getClientId());
        assertEquals("Grüße, player two", this.recorder.getChatEvents().get(0).getMessage());
    }

    /** ADMIN_PACKET_SERVER_CHAT with a 'give money' action carries a uint64 amount. */
    @Test
    public void testChatGiveMoney() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CHAT)
                .writeUint8(6) // NETWORK_ACTION_GIVE_MONEY
                .writeUint8(2) // DESTTYPE_CLIENT
                .writeUint32(1L)
                .writeString("here is money")
                .writeUint64(new BigInteger("30000000000")));
        TestUtils.awaitSize(this.recorder.getChatEvents(), 1, 5, "a ChatEvent");
        assertEquals("here is money", this.recorder.getChatEvents().get(0).getMessage());
    }

    /** ADMIN_PACKET_SERVER_RCON: uint16 colour, string output. */
    @Test
    public void testRCon() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_RCON)
                .writeUint16(2) // CC_GENERAL
                .writeString("Current date: 1950-01-01"));
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_RCON)
                .writeUint16(8) // CC_ERROR
                .writeString("Unknown command"));
        TestUtils.awaitSize(this.recorder.getRConEvents(), 2, 5, "two RConEvents");

        assertEquals(2, this.recorder.getRConEvents().get(0).getColor());
        assertEquals("Current date: 1950-01-01", this.recorder.getRConEvents().get(0).getMessage());
        assertEquals(8, this.recorder.getRConEvents().get(1).getColor());
        assertEquals("Unknown command", this.recorder.getRConEvents().get(1).getMessage());
    }

    /**
     * ADMIN_PACKET_SERVER_RCON_END: string command. It must not also produce a
     * ConsoleEvent: the switch used to fall through into the CONSOLE case.
     */
    @Test
    @DisplayName("RCON_END does not fall through into CONSOLE")
    public void testRConEnd() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_RCON).writeUint16(2).writeString("output"));
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_RCON_END).writeString("getdate"));
        TestUtils.awaitSize(this.recorder.getRConEndEvents(), 1, 5, "an RConEndEvent");

        assertEquals("getdate", this.recorder.getRConEndEvents().get(0).getCommand());
        assertEquals(1, this.recorder.getRConEvents().size());
        assertEquals(0, this.recorder.getConsoleEvents().size(),
                "RCON_END has no origin, so no ConsoleEvent may be dispatched");
    }

    /** ADMIN_PACKET_SERVER_CONSOLE: string origin, string text. */
    @Test
    public void testConsole() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CONSOLE)
                .writeString("net").writeString("Admin connection closed"));
        TestUtils.awaitSize(this.recorder.getConsoleEvents(), 1, 5, "a ConsoleEvent");

        ConsoleEvent event = this.recorder.getConsoleEvents().get(0);
        assertEquals("net", event.getOrigin());
        assertEquals("Admin connection closed", event.getMessage());
    }

    /** ADMIN_PACKET_SERVER_GAMESCRIPT: string json. */
    @Test
    public void testGameScript() throws IOException {
        String json = "{\"id\":\"0\",\"data\":\"7\"}";
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_GAMESCRIPT).writeString(json));
        TestUtils.awaitSize(this.recorder.getGameScriptEvents(), 1, 5, "a GameScriptEvent");

        GameScriptEvent event = this.recorder.getGameScriptEvents().get(0);
        assertEquals(json, event.getMessage());
    }

    /** ADMIN_PACKET_SERVER_PONG: uint32 identifier echoed from the ping. */
    @Test
    public void testPong() throws IOException {
        this.harness.getSend().ping(4294967295L); // uint32 max
        TestUtils.awaitSize(this.recorder.getPongEvents(), 1, 5, "a PongEvent");

        PongEvent event = this.recorder.getPongEvents().get(0);
        assertEquals(4294967295L, event.getPingId());
    }

    /**
     * ADMIN_PACKET_SERVER_CMD_NAMES: repeated bool/uint16/string, terminated by
     * a false bool. SendCmdNames splits into several packets when COMPAT_MTU is
     * reached, so the loop must handle consecutive packets.
     */
    @Test
    @DisplayName("command names arrive across packet boundaries")
    public void testCmdNames() throws IOException {
        for (int packetIndex = 0; packetIndex < 2; packetIndex++) {
            PacketWriter packet = new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CMD_NAMES);
            for (int commandId = packetIndex * 3; commandId < packetIndex * 3 + 3; commandId++) {
                packet.writeBool8(true).writeUint16(commandId).writeString("CMD_" + commandId);
            }
            packet.writeBool8(false);
            send(packet);
        }
        TestUtils.await(() -> this.harness.getGame().getCommandNames().size() == 6, 5, "six command names");

        assertEquals("CMD_0", this.harness.getGame().getCommandNames().get(0));
        assertEquals("CMD_5", this.harness.getGame().getCommandNames().get(5));
        assertStillConnected();
    }

    /**
     * ADMIN_PACKET_SERVER_CMD_LOGGING (id 127): uint32 client, uint8 company,
     * uint16 command, variable length command buffer, uint32 frame. Logging only.
     */
    @Test
    @DisplayName("CMD_LOGGING decodes the variable length command buffer")
    public void testCmdLogging() throws IOException {
        byte[] commandBuffer = {1, 2, 3, 4, 5, 6, 7, 8, 9};
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CMD_LOGGING)
                .writeUint32(2L).writeUint8(1).writeUint16(37).writeBytes(commandBuffer).writeUint32(12345L));
        assertStillConnected();
        assertEquals(0, this.recorder.getCompanyEvents().size(), "command logging is for logging purposes only");
    }

    /**
     * ADMIN_PACKET_SERVER_CMD_LOGGING_OLD (id 123): the admin protocol version 1
     * layout, uint32 client, uint8 company, uint16 command, uint32 p1, uint32
     * p2, uint32 tile, string text, uint32 frame.
     */
    @Test
    public void testCmdLoggingOld() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CMD_LOGGING_OLD)
                .writeUint32(3L).writeUint8(1).writeUint16(42)
                .writeUint32(1000L).writeUint32(2L).writeUint32(30000L)
                .writeString("Build station").writeUint32(999L));
        assertStillConnected();
    }

    /** ADMIN_PACKET_SERVER_NEWGAME carries no payload and no functional event. */
    @Test
    public void testNewGame() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_NEWGAME));
        assertStillConnected();
        assertEquals(0, this.recorder.getCompanyEvents().size());
        assertEquals(0, this.recorder.getClientEvents().size());
    }

    /** ADMIN_PACKET_SERVER_ERROR: uint8 NetworkErrorCode. The session survives. */
    @Test
    @DisplayName("an error packet is reported and does not drop the connection")
    public void testError() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_ERROR).writeUint8(10)); // WRONG_PASSWORD
        assertStillConnected();
        assertEquals(0, this.recorder.getChatEvents().size());
    }

    /** ADMIN_PACKET_SERVER_FULL: the server cannot accept this admin. */
    @Test
    public void testFull() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_FULL));
        this.harness.getServer().awaitClientDisconnect(5);
    }

    /** ADMIN_PACKET_SERVER_BANNED: the source address is banned. */
    @Test
    public void testBanned() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_BANNED));
        this.harness.getServer().awaitClientDisconnect(5);
    }

    /** ADMIN_PACKET_SERVER_SHUTDOWN: the server is going away. */
    @Test
    public void testShutdownPacket() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_SHUTDOWN));
        this.harness.getServer().awaitClientDisconnect(5);
    }

    /**
     * ADMIN_PACKET_SERVER_AUTH_REQUEST (id 128): uint8 method, then key
     * exchange material. Secure admin auth is not implemented, so an encrypted
     * session cannot continue.
     */
    @Test
    @DisplayName("AUTH_REQUEST is recognised and refused, not misread as an unknown packet")
    public void testAuthRequest() throws IOException {
        PacketWriter packet = new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_AUTH_REQUEST)
                .writeUint8(1) // NETWORK_AUTH_METHOD_X25519_PAKE
                .writeBytes(new byte[32]) // server public key
                .writeBytes(new byte[24]); // nonce
        send(packet);
        this.harness.getServer().awaitClientDisconnect(5);
    }

    /** ADMIN_PACKET_SERVER_ENABLE_ENCRYPTION (id 129): 24 bytes nonce. */
    @Test
    public void testEnableEncryption() throws IOException {
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_ENABLE_ENCRYPTION).writeBytes(new byte[24]));
        this.harness.getServer().awaitClientDisconnect(5);
    }

    /** A packet id the protocol does not define must not be read as a payload. */
    @Test
    public void testUnknownPacketId() throws IOException {
        PacketWriter packet = new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_FULL);
        byte[] bytes = packet.toBytes();
        bytes[2] = (byte) 250; // Unknown server packet id
        this.harness.getServer().send(new RawPacket(bytes));
        this.harness.getServer().awaitClientDisconnect(5);
    }

    private static PacketWriter companyInfo(int companyId) {
        return new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_COMPANY_INFO)
                .writeUint8(companyId)
                .writeString("Test Company")
                .writeString("Test Manager")
                .writeUint8(2)
                .writeBool8(true)
                .writeUint32(1950L)
                .writeBool8(false)
                .writeUint8(0)
                .writeUint8(0).writeUint8(255).writeUint8(255).writeUint8(255);
    }

    /** A packet whose header type byte was changed, for the invalid packet test. */
    private static class RawPacket extends PacketWriter {

        private final byte[] bytes;

        RawPacket(byte[] bytes) {
            super(PacketServerType.ADMIN_PACKET_SERVER_FULL);
            this.bytes = bytes;
        }

        @Override
        public byte[] toBytes() {
            return this.bytes;
        }
    }

    /** Documents the protocol version the fake server announces. */
    @Test
    public void testProtocolVersionAdvertised() {
        assertNotNull(this.harness.getGame().getGameInfo());
    }
}
