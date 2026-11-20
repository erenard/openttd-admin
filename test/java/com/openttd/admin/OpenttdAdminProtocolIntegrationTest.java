package com.openttd.admin;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertNotNull;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.IOException;
import java.util.List;
import java.util.regex.Pattern;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import com.google.gson.JsonArray;
import com.google.gson.JsonParser;
import com.google.gson.JsonObject;
import com.openttd.TestUtils;
import com.openttd.admin.event.GameScriptEvent;
import com.openttd.admin.event.RConEvent;
import com.openttd.constant.OTTD;
import com.openttd.gamescript.GSJsonRequest;
import com.openttd.gamescript.GSNewsPaper;
import com.openttd.network.admin.GameInfo;
import com.openttd.network.admin.NetworkAdminSender;
import com.openttd.network.constant.GameScript.GSCommand;
import com.openttd.network.constant.GameVersion;
import com.openttd.network.constant.TcpAdmin.AdminUpdateFrequency;
import com.openttd.network.constant.TcpAdmin.AdminUpdateType;
import com.openttd.network.core.Configuration;
import com.openttd.server.EventRecorder;
import com.openttd.server.OpenttdServerWrapper;
import com.openttd.server.UpstreamAdminProtocol;
import com.openttd.server.UpstreamCalendar;

/**
 * The admin protocol (port 3977) against a real OpenTTD server.
 *
 * The offline suite (FakeAdminServer) checks what NetworkAdmin decodes and what
 * NetworkAdminSender encodes; this suite checks that the real server accepts the
 * encoding and that the packets it sends back end up in the game model and as
 * functional events. Expected behaviour is taken from
 * `git show 15.3:src/network/network_admin.cpp`, `network_server.cpp` and
 * `console_cmds.cpp`.
 *
 * The dedicated server has no clients and no companies (no game-port client
 * joins here, see the integrationTest task in build.gradle), so client and
 * company assertions stay modest: the server's own client (#0) is always there,
 * and company polls must simply be accepted.
 */
public class OpenttdAdminProtocolIntegrationTest extends OpenttdServerWrapper {

    /** git show 15.3:src/console_type.h, CC_DEFAULT = TC_SILVER. */
    private static final char CC_DEFAULT = 0x01;
    /** CC_ERROR = TC_RED. */
    private static final char CC_ERROR = 0x03;
    /** CC_INFO = TC_YELLOW, also a colour an external chat may use. */
    private static final char CC_INFO = 0x08;
    /** TC_INVALID, rejected by IsValidConsoleColour. */
    private static final char TC_INVALID = 0xFF;

    private OpenttdAdmin admin;
    private EventRecorder recorder;
    private long nextPingId;

    @BeforeEach
    public void setUp() throws IOException {
        Configuration configuration = new Configuration();
        configuration.password = TestUtils.readSecrets().getProperty("admin_password");
        configuration.name = "openttd-admin protocol test";

        this.admin = new OpenttdAdmin(configuration);
        this.recorder = new EventRecorder();
        this.recorder.registerWith(this.admin);
        this.admin.startup();

        // startup() only connects; the WELCOME packet is what proves the join.
        TestUtils.await(() -> {
            try {
                GameInfo gameInfo = this.admin.getGame().getGameInfo();
                return gameInfo != null && gameInfo.getServerName() != null;
            } catch (NullPointerException notConnectedYet) {
                return false;
            }
        }, 20, "the WELCOME packet");
        this.recorder.clear();
    }

    @AfterEach
    public void tearDown() {
        this.admin.shutdown();
    }

    /** ADMIN_PACKET_ADMIN_JOIN -> PROTOCOL + WELCOME, decoded into GameInfo. */
    @Test
    @DisplayName("joining authenticates and fills the game model")
    public void testJoinAndWelcome() {
        GameInfo gameInfo = this.admin.getGame().getGameInfo();
        assertTrue(this.admin.isConnected(), "the admin connection is open");
        assertNotNull(gameInfo.getServerName(), "server name");
        assertFalse(gameInfo.getServerName().isEmpty(), "server name is not empty");
        assertTrue(gameInfo.getServerRevision().contains(GameVersion.OPENTTD),
                "revision " + gameInfo.getServerRevision() + " must match the advertised version");
        assertTrue(gameInfo.isServerDedicated(), "the test server is dedicated");
        assertEquals(UpstreamAdminProtocol.GAME_ADMIN_VERSION, gameInfo.getVersion(), "protocol version");

        // openttd/openttd.cfg: map_x = map_y = 8, i.e. 64x64 tiles.
        assertEquals(gameInfo.getMapWidth(), gameInfo.getMapHeight(), "square map");
        assertTrue(gameInfo.getMapWidth() > 0, "map size");
        assertTrue(gameInfo.getMapSeed() > 0, "generation seed");

        // [game_creation] starting_year = 1950.
        assertEquals(1950, UpstreamCalendar.yearOf((int) gameInfo.getStartDate()), "start year");
        // WELCOME carries only the map's start date; the current date comes with
        // the ADMIN_PACKET_SERVER_DATE the WELCOME handler polls for. Game.updateAdmin
        // swaps the network model on every packet, so the game info has to be read
        // again rather than kept in a reference.
        TestUtils.await(() -> {
            GameInfo latest = this.admin.getGame().getGameInfo();
            return latest.getCurrentDate() >= latest.getStartDate();
        }, 10, "the current game date");
    }

    /** Receive_ADMIN_JOIN compares the password and answers NETWORK_ERROR_WRONG_PASSWORD. */
    @Test
    @DisplayName("a wrong password never produces a WELCOME")
    public void testWrongPassword() throws IOException {
        Configuration configuration = new Configuration();
        configuration.password = "not-the-admin-password";
        configuration.name = "openttd-admin wrong password test";
        OpenttdAdmin rejected = new OpenttdAdmin(configuration);
        EventRecorder rejectedEvents = new EventRecorder();
        rejectedEvents.registerWith(rejected);
        rejected.startup();
        try {
            // Ping through the reconnect window: every connection is rejected,
            // so neither a WELCOME nor a PONG can ever arrive.
            for (int attempt = 0; attempt < 6; attempt++) {
                NetworkAdminSender send = rejected.getSend();
                if (send != null) {
                    send.ping(9001L);
                }
                TestUtils.wait(1);
            }
            assertTrue(rejectedEvents.getPongEvents().isEmpty(), "no PONG on a rejected connection");
            assertTrue(rejectedEvents.getGameScriptEvents().isEmpty(), "no gamescript traffic");
            assertFalse(hasGameInfo(rejected), "no WELCOME on a rejected connection");
        } finally {
            rejected.shutdown();
        }
    }

    private static boolean hasGameInfo(OpenttdAdmin admin) {
        try {
            GameInfo gameInfo = admin.getGame().getGameInfo();
            return gameInfo != null && gameInfo.getServerName() != null;
        } catch (NullPointerException notConnectedYet) {
            return false;
        }
    }

    /**
     * Receive_ADMIN_UPDATE_FREQUENCY closes the connection on a frequency the
     * server does not advertise, so every combination sent must be one the
     * PROTOCOL packet promised - including the ones Game.updateAdmin registers.
     */
    @Test
    @DisplayName("every advertised update type/frequency pair is accepted")
    public void testUpdateFrequencyCombinations() {
        NetworkAdminSender send = this.admin.getSend();
        assertNotNull(send, "a sender once connected");

        for (AdminUpdateType type : UpstreamAdminProtocol.FREQUENCIES.keySet()) {
            int advertised = UpstreamAdminProtocol.FREQUENCIES.get(type);
            for (AdminUpdateFrequency frequency : AdminUpdateFrequency.values()) {
                if ((advertised & frequency.mask) == 0) {
                    continue;
                }
                send.updateFrequency(type, frequency);
            }
        }
        assertConnectionUsable();
    }

    /**
     * ADMIN_UPDATE_COMPANY_INFO only advertises Poll|Automatic. The sender falls
     * back to an advertised frequency instead of letting the server drop the
     * connection, which is what keeps strategyboard's objectives running.
     */
    @Test
    @DisplayName("an unadvertised frequency falls back, the connection survives")
    public void testUpdateFrequencyFallback() {
        this.admin.getSend().updateFrequency(AdminUpdateType.ADMIN_UPDATE_COMPANY_INFO,
                AdminUpdateFrequency.ADMIN_FREQUENCY_ANUALLY);
        assertConnectionUsable();
    }

    /** Receive_ADMIN_POLL answers each pollable type. */
    @Test
    @DisplayName("all six pollable update types answer")
    public void testPolls() {
        NetworkAdminSender send = this.admin.getSend();
        assertNotNull(send, "a sender once connected");
        this.recorder.clear();

        send.pollDate();
        TestUtils.awaitSize(this.recorder.getDateEvents(), 1, 10, "a DateEvent from pollDate()");

        send.pollClientInfo(-1); // UINT32_MAX: every client, including the server's own
        // git show 15.3:src/network/network_type.h: CLIENT_ID_SERVER = 1, so a
        // dedicated server has exactly one client info, and id 0 is INVALID_CLIENT_ID.
        TestUtils.await(() -> this.admin.getGame().getClient(1) != null
                && this.admin.getGame().getClient(1).getName() != null, 10, "the server client #1");
        assertEquals(1, this.admin.getGame().getClient(1).getId(), "client #1");
        assertNotNull(this.admin.getGame().getClient(1).getName(), "client name");
        assertEquals(255, this.admin.getGame().getClient(1).getCompanyId(), "the server plays as spectator");

        send.pollClientInfo(1);
        TestUtils.awaitSize(this.recorder.getClientEvents(), 2, 10, "a ClientEvent for client #1");

        // No companies on a dedicated server with no clients: the polls are
        // accepted, and nothing is added to the model.
        send.pollCompanyInfo((short) -1);
        send.pollCompanyEconomy((short) -1);
        send.pollCompanyStats((short) -1);
        assertConnectionUsable();
        assertTrue(this.admin.getGame().getCompanies().isEmpty(), "no companies on the dedicated server");

        send.pollCmdNames();
        TestUtils.await(() -> !this.admin.getGame().getCommandNames().isEmpty(), 10, "command names");
        assertTrue(this.admin.getGame().getCommandNames().size() > 10, "a command name list");
    }

    /**
     * Receive_ADMIN_CHAT calls NetworkServerSendChat(..., from_admin = true), and
     * NetworkAdminChat returns immediately for admin messages: the admin gets no
     * SERVER_CHAT echo of its own message, only the console relay.
     */
    @Test
    @DisplayName("chat is accepted but never echoed back to the sender")
    public void testChatBroadcastAndClient() {
        String broadcast = "integration broadcast " + System.nanoTime();
        String toServerClient = "integration client chat " + System.nanoTime();

        this.admin.getSend().chatBroadcast(broadcast);
        awaitConsoleMessage(broadcast);

        // DESTTYPE_CLIENT with dest == CLIENT_ID_SERVER (1) prints the message on
        // the server's console; dest 0 is INVALID_CLIENT_ID and prints nothing.
        this.admin.getSend().chatClient(1, toServerClient);
        awaitConsoleMessage(toServerClient);

        assertTrue(this.recorder.getChatEvents().isEmpty(),
                "an admin's own chat must not come back as a ChatEvent: " + this.recorder.getChatEvents());

        this.admin.getSend().chatClient(0, "chat to an invalid client id " + System.nanoTime());
        assertConnectionUsable();
    }

    /** Team chat with no members: nothing to print, but the packet is legal. */
    @Test
    @DisplayName("team chat to an empty company is accepted")
    public void testChatCompany() {
        String team = "integration team chat " + System.nanoTime();
        this.admin.getSend().chatCompany((short) 0, team);
        assertConnectionUsable();
        assertTrue(this.recorder.getChatEvents().isEmpty(), "no ChatEvent for admin chat");
        assertFalse(consoleContains(team), "no client in company 0, so nothing is printed");
    }

    /** Chat text is relayed verbatim, including non-ASCII bytes. */
    @Test
    @DisplayName("chat survives the server's string validation unchanged")
    public void testChatUtf8() {
        String message = "unicode: héllo ütf8 — 100 €";
        this.admin.getSend().chatBroadcast(message);
        awaitConsoleMessage(message);
    }

    /** Receive_ADMIN_EXTERNAL_CHAT requires IsValidConsoleColour. */
    @Test
    @DisplayName("external chat with a console colour is printed")
    public void testExternalChatValidColour() {
        String message = "external chat " + System.nanoTime();
        this.admin.getSend().externalChat("strategyboard", CC_INFO, "bot", message);
        awaitConsoleMessage(message);
        assertConnectionUsable();
    }

    /** An invalid colour makes the server send NETWORK_ERROR_ILLEGAL_PACKET and close. */
    @Test
    @DisplayName("external chat with an invalid colour drops the connection, then reconnects")
    public void testExternalChatInvalidColour() {
        String message = "invalid colour chat " + System.nanoTime();
        this.admin.getSend().externalChat("strategyboard", TC_INVALID, "bot", message);

        assertFalse(consoleContains(message), "an illegal packet is not printed");
        // NetworkAdmin reconnects every 5 s, so the admin becomes usable again.
        assertConnectionUsable();
    }

    /** Receive_ADMIN_RCON redirects the console to ADMIN_PACKET_SERVER_RCON. */
    @Test
    @DisplayName("rcon replies carry colour and text, and RCON_END echoes the command")
    public void testRConCommands() {
        assertRCon("getdate", "Date: 19", CC_DEFAULT);
        assertRCon("getseed", "Generation Seed:", CC_DEFAULT);
        // NetworkPrintClients prints the server's own client #1, IP "server".
        assertRCon("clients", "Client #1", CC_INFO);
        assertRCon("clients", "IP: server", CC_INFO);

        // list_cmds prints one line per command; getdate is one of them.
        assertRCon("list_cmds", "getdate", CC_DEFAULT);
    }

    /** `companies` lists nothing when there are no companies, but still ends. */
    @Test
    @DisplayName("rcon of an empty command produces only RCON_END")
    public void testRConWithoutOutput() {
        NetworkAdminSender send = this.admin.getSend();
        assertNotNull(send, "a sender once connected");
        this.recorder.clear();

        send.rcon("companies");
        TestUtils.await(() -> !this.recorder.getRConEndEvents().isEmpty(), 10, "an RConEndEvent");
        assertEquals("companies", this.recorder.getRConEndEvents().get(0).getCommand(), "command echo");
        assertTrue(this.recorder.getRConEvents().isEmpty(), "no companies, so no output lines");
    }

    /** IConsoleCmdExec ends with "Command '{}' not found." in CC_ERROR. */
    @Test
    @DisplayName("an unknown rcon command is reported in the error colour")
    public void testRConUnknownCommand() {
        assertRCon("not-a-command", "Command 'not-a-command' not found.", CC_ERROR);
    }

    /** RCON_END must not fall through into a console event (see Game.updateAdmin). */
    @Test
    @DisplayName("rcon output is not relayed as console output")
    public void testRConIsNotRelayedAsConsole() {
        this.recorder.clear();
        this.admin.getSend().rcon("getdate");
        TestUtils.await(() -> !this.recorder.getRConEndEvents().isEmpty() && !this.recorder.getRConEvents().isEmpty(),
                10, "an RConEvent and its RConEndEvent");

        // The redirected console goes as ADMIN_PACKET_SERVER_RCON. ADMIN_PACKET_SERVER_RCON_END
        // dispatches an RConEndEvent only; unrelated Debug(net) lines still arrive
        // as console events, so assert on the command's output, not on emptiness.
        String reply = this.recorder.getRConEvents().get(0).getMessage();
        assertFalse(consoleContains(reply), "rcon output stays out of the console events: " + reply);
        assertTrue(this.recorder.getConsoleEvents().stream()
                .noneMatch(e -> e.getMessage().contains("Date: 19")), "no rcon output in console events");
    }

    /**
     * ADMIN_UPDATE_CONSOLE Automatic makes DebugReconsiderSendRemoteMessages
     * queue Debug messages, and DebugSendRemoteMessages passes them to
     * ADMIN_PACKET_SERVER_CONSOLE. AdminCmd's GSLog.Info is such a message.
     */
    @Test
    @DisplayName("gamescript log lines are relayed as console events")
    public void testConsoleRelayFromGamescript() {
        this.recorder.clear();
        GSJsonRequest request = new GSJsonRequest(GSCommand.countIndustry, new JsonArray());
        this.admin.getGSExecutor().send(request);
        TestUtils.await(request::hasResponse, 15, "the gamescript reply");

        // main.nut: GSLog.Info("Command: " + id + " " + cmd + " " + args). Squirrel
        // prints an array as "(array : 0x...)", not "[]". debug.cpp passes the debug
        // category as the console origin, so the origin is "script".
        Pattern commandLog = Pattern.compile("Command: \\d+ " + GSCommand.countIndustry.ordinal() + " \\(array");
        TestUtils.await(() -> this.recorder.getConsoleEvents().stream()
                .anyMatch(e -> "script".equals(e.getOrigin()) && commandLog.matcher(e.getMessage()).find()),
                10, "the gamescript's console log of the command");
    }

    /** GSExecutor correlates the reply by request id, against the real gamescript. */
    @Test
    @DisplayName("countIndustry answers with a number")
    public void testGameScriptCountIndustry() {
        GSJsonRequest request = new GSJsonRequest(GSCommand.countIndustry, new JsonArray());
        this.admin.getGSExecutor().send(request);
        TestUtils.await(request::hasResponse, 15, "the countIndustry reply");
        assertTrue(request.getResult().matches("\\d+"), "industry count '" + request.getResult() + "'");
    }

    /** GSNewsPaper to all companies and to one company (company 0 does not exist). */
    @Test
    @DisplayName("createNews answers for a broadcast and a company")
    public void testGameScriptNews() {
        GSNewsPaper broadcast = new GSNewsPaper(GSNewsPaper.NewsType.NT_GENERAL, "Protocol broadcast");
        this.admin.getGSExecutor().send(broadcast);
        TestUtils.await(broadcast::hasResponse, 15, "the broadcast news reply");
        assertEquals(Boolean.TRUE, broadcast.getResult(), "GSNewsPaper.New() returns true");

        GSNewsPaper forCompany = new GSNewsPaper((short) 0, GSNewsPaper.NewsType.NT_GENERAL, "Protocol company news");
        this.admin.getGSExecutor().send(forCompany);
        TestUtils.await(forCompany::hasResponse, 15, "the company news reply");
        // No company 0 on a dedicated server, so ScriptNews refuses it; the reply
        // still arrives, which is what GSExecutor correlates.
        assertEquals(Boolean.FALSE, forCompany.getResult(), "company 0 does not exist");
        assertConnectionUsable();
    }

    /**
     * The goal helpers go through GSExecutor, so AdminCmd's reply is correlated.
     * removeGoal used to pass the whole arguments array to GSGoal.Remove, which
     * made the gamescript fail; the reply must be a removal, not an exception.
     */
    @Test
    @DisplayName("goal commands add, remove and clear through the gamescript")
    public void testGameScriptGoals() {
        NetworkAdminSender send = this.admin.getSend();
        assertNotNull(send, "a sender once connected");
        this.recorder.clear();

        send.addGlobalGoal("protocol goal", OTTD.GoalType.GT_NONE, 0);
        String added = awaitGameScriptData(1);
        assertTrue(added.matches("\\d+"), "goal id '" + added + "'");

        send.removeGoal(Integer.parseInt(added));
        String removed = awaitGameScriptData(2);
        assertEquals("true", removed, "removing a valid goal returns true");

        send.removeAllGoal();
        // Nothing left to remove, so AdminCmd answers with false rather than an exception.
        awaitGameScriptData(3);
        assertConnectionUsable();
    }

    /** AdminCmd validates town and cargo effect itself, so a bad id is not an exception. */
    @Test
    @DisplayName("setTownCargoGoal validates an unknown town instead of throwing")
    public void testGameScriptTownCargoGoal() {
        this.recorder.clear();
        this.admin.getSend().setTownCargoGoal(10000, OTTD.TownEffect.TE_MAIL, 100);
        assertEquals("false", awaitGameScriptData(1), "an invalid town is refused");
        assertConnectionUsable();
    }

    /**
     * A command AdminCmd does not know is answered with {id, exception}; the raw
     * JSON path (NetworkAdminSender.gameScript) is what robot rules use.
     */
    @Test
    @DisplayName("an unknown gamescript command answers with an exception")
    public void testGameScriptUnknownCommand() {
        this.recorder.clear();
        assertTrue(this.admin.getSend().gameScript("{\"i\":\"4242\",\"c\":99,\"a\":[]}"), "raw json is sent");

        TestUtils.await(() -> {
            for (GameScriptEvent event : this.recorder.getGameScriptEvents()) {
                JsonObject reply = new JsonParser().parse(event.getMessage()).getAsJsonObject();
                if (reply.has("exception") && reply.get("exception").getAsString().equals("Unknown command")) {
                    return true;
                }
            }
            return false;
        }, 15, "the 'Unknown command' reply");
        assertConnectionUsable();
    }

    /** A reply for a request that was never sent must not disturb the executor. */
    @Test
    @DisplayName("a gamescript reply without a pending request is ignored")
    public void testGameScriptReplyWithoutRequest() {
        this.recorder.clear();
        assertTrue(this.admin.getSend().gameScript("{\"i\":\"99999\",\"c\":0,\"a\":[]}"), "request sent");

        // GSExecutor has no request with id 99999; it logs and ignores the reply,
        // so the listener thread must survive and the connection stay open.
        TestUtils.await(() -> {
            for (GameScriptEvent event : this.recorder.getGameScriptEvents()) {
                JsonObject reply = new JsonParser().parse(event.getMessage()).getAsJsonObject();
                if (reply.has("id") && reply.get("id").getAsString().equals("99999")) {
                    return true;
                }
            }
            return false;
        }, 15, "the reply with id 99999");
        assertConnectionUsable();
    }

    /** startup() shuts the previous client down first, shutdown() works before startup(). */
    @Test
    @DisplayName("startup twice and shutdown before startup")
    public void testLifecycle() throws IOException {
        Configuration configuration = new Configuration();
        configuration.password = TestUtils.readSecrets().getProperty("admin_password");
        configuration.name = "openttd-admin lifecycle test";
        OpenttdAdmin lifecycle = new OpenttdAdmin(configuration);
        EventRecorder events = new EventRecorder();
        events.registerWith(lifecycle);

        lifecycle.shutdown(); // never started
        lifecycle.startup();
        lifecycle.startup(); // the second startup() must not leave a dangling thread

        try {
            assertUsable(lifecycle, events, 4242L, 20);
            assertFalse(this.admin.getGame().getGameInfo().getServerName().isEmpty(), "still connected");
        } finally {
            lifecycle.shutdown();
        }
        assertFalse(lifecycle.isConnected(), "shutdown() disconnects");
    }

    /* Helpers */

    /** Ping until the server answers: also covers the 5 s reconnect window. */
    private void assertConnectionUsable() {
        assertUsable(this.admin, this.recorder, this.nextPingId++, 20);
    }

    /**
     * Ping until a PongEvent with `pingId` arrives. The sender is recreated on
     * every reconnect, so a ping sent while the connection is down is dropped -
     * hence pinging inside the awaited condition.
     */
    private static void assertUsable(OpenttdAdmin admin, EventRecorder recorder, long pingId, int timeoutSeconds) {
        TestUtils.await(() -> {
            NetworkAdminSender send = admin.getSend();
            if (send == null) {
                return false;
            }
            send.ping(pingId);
            return recorder.getPongEvents().stream().anyMatch(e -> e.getPingId() == pingId);
        }, timeoutSeconds, "a PONG for ping " + pingId);
    }

    private void assertRCon(String command, String expectedText, char expectedColour) {
        NetworkAdminSender send = this.admin.getSend();
        assertNotNull(send, "a sender once connected");
        this.recorder.clear();

        send.rcon(command);
        TestUtils.await(() -> !this.recorder.getRConEndEvents().isEmpty(), 10, "an RConEndEvent for " + command);
        assertEquals(command, this.recorder.getRConEndEvents().get(0).getCommand(), "command echo");

        List<RConEvent> replies = this.recorder.getRConEvents();
        assertTrue(replies.stream().anyMatch(e -> e.getMessage().contains(expectedText)),
                command + " must reply '" + expectedText + "', got " + replies);
        assertTrue(replies.stream().anyMatch(e -> e.getColor() == expectedColour),
                command + " must reply in colour " + expectedColour + ", got " + replies);
        assertConnectionUsable();
    }

    private void awaitConsoleMessage(String text) {
        TestUtils.await(() -> consoleContains(text), 10, "a ConsoleEvent containing '" + text + "'");
    }

    private boolean consoleContains(String text) {
        return this.recorder.getConsoleEvents().stream().anyMatch(e -> e.getMessage().contains(text));
    }

    /**
     * Reply of a gamescript command issued after the recorder was cleared.
     *
     * @param expectedReplyNumber which reply (1 = the first one after clear())
     * @return the reply's "data" field, or its "exception" text
     */
    private String awaitGameScriptData(int expectedReplyNumber) {
        TestUtils.awaitSize(this.recorder.getGameScriptEvents(), expectedReplyNumber, 15,
                "gamescript reply #" + expectedReplyNumber);
        JsonObject reply = new JsonParser()
                .parse(this.recorder.getGameScriptEvents().get(expectedReplyNumber - 1).getMessage())
                .getAsJsonObject();
        if (reply.has("data")) {
            return reply.get("data").getAsString();
        }
        if (reply.has("exception")) {
            throw new AssertionError("gamescript replied an exception: " + reply.get("exception").getAsString());
        }
        throw new AssertionError("no data or exception in " + reply);
    }
}
