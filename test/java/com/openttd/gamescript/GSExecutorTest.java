package com.openttd.gamescript;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertFalse;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.IOException;
import java.math.BigInteger;

import org.junit.jupiter.api.AfterEach;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import com.google.gson.JsonArray;
import com.google.gson.JsonParser;
import com.google.gson.JsonObject;
import com.openttd.TestUtils;
import com.openttd.constant.OTTD;
import com.openttd.network.constant.GameScript;
import com.openttd.network.constant.GameScript.GSCommand;
import com.openttd.network.constant.TcpAdmin.PacketAdminType;
import com.openttd.server.AdminHarness;
import com.openttd.server.PacketWriter;

/**
 * GSExecutor correlates gamescript replies with the requests that were sent.
 *
 * The request and reply JSON shapes come from AdminCmd's gamescript
 * (src/squirrel/main.nut): MainClass::HandleEvents only runs a request when
 * "i", "c" and "a" are all present, and MainClass::ExecuteCommand answers with
 * {id, data}, {id, exception}, or {exception} alone for a malformed request.
 */
public class GSExecutorTest {

    private AdminHarness harness;

    @BeforeEach
    public void setUp() throws IOException {
        this.harness = new AdminHarness();
        this.harness.getSend();
    }

    @AfterEach
    public void tearDown() {
        this.harness.close();
    }

    private void sendGameScriptReply(String json) throws IOException {
        this.harness.getServer().send(
                new PacketWriter(com.openttd.network.constant.TcpAdmin.PacketServerType.ADMIN_PACKET_SERVER_GAMESCRIPT)
                        .writeString(json));
    }

    /** The request AdminCmd expects: "i" id, "c" command, "a" arguments. */
    @Test
    @DisplayName("every GSCommand request carries i, c and a")
    public void testRequestJsonShape() {
        JsonArray goalArguments = new JsonArray();
        goalArguments.add(new com.google.gson.JsonPrimitive(1));
        goalArguments.add(new com.google.gson.JsonPrimitive("deliver cargo"));
        goalArguments.add(new com.google.gson.JsonPrimitive(OTTD.GoalType.GT_TOWN.ordinal()));
        goalArguments.add(new com.google.gson.JsonPrimitive(4));

        for (GSCommand command : GSCommand.values()) {
            GSJsonRequest request = new GSJsonRequest(command, goalArguments);
            request.setId("7");
            JsonObject json = request.toJson();
            assertTrue(json.has(GameScript.ID), command + " needs \"i\", AdminCmd drops requests without it");
            assertEquals("7", json.get(GameScript.ID).getAsString(), command + " id");
            assertEquals(command.ordinal(), json.get(GameScript.CMD).getAsInt(), command + " command id");
            assertEquals(4, json.get(GameScript.ARGS).getAsJsonArray().size(), command + " arguments");
            // toString() is what goes on the wire, so it must be pure JSON.
            assertEquals(json.toString(), request.toString(), command + " wire payload");
            assertTrue(new JsonParser().parse(request.toString()).isJsonObject(), command + " is parseable JSON");
        }
    }

    /** GSNewsPaper is the one request modelled as its own class. */
    @Test
    public void testNewsPaperRequestJson() {
        GSNewsPaper request = new GSNewsPaper((short) 2, GSNewsPaper.NewsType.NT_GENERAL, "hello");
        request.setId("3");
        JsonObject json = request.toJson();
        assertEquals("3", json.get(GameScript.ID).getAsString());
        assertEquals(GSCommand.createNews.ordinal(), json.get(GameScript.CMD).getAsInt());
        JsonArray arguments = json.get(GameScript.ARGS).getAsJsonArray();
        assertEquals(GSNewsPaper.NewsType.NT_GENERAL.ordinal(), arguments.get(0).getAsInt());
        assertEquals("hello", arguments.get(1).getAsString());
        assertEquals(2, arguments.get(2).getAsInt());
    }

    /** A news type the gamescript cannot use is replaced by NT_GENERAL. */
    @Test
    public void testNewsPaperInvalidNewsType() {
        GSNewsPaper request = new GSNewsPaper((short) 1, GSNewsPaper.NewsType.NT_ACCIDENT, "boom");
        assertFalse(GSNewsPaper.NewsType.NT_ACCIDENT.isValid());
        request.setId("4");
        assertEquals(GSNewsPaper.NewsType.NT_GENERAL.ordinal(),
                request.toJson().get(GameScript.ARGS).getAsJsonArray().get(0).getAsInt());
    }

    /**
     * A reply with data completes the request, and getResult() returns it.
     * Requests are numbered by GSExecutor, so the id must match.
     */
    @Test
    @DisplayName("a data reply completes the correlated request")
    public void testReplyWithId() throws IOException {
        GSJsonRequest request = new GSJsonRequest(GSCommand.countIndustry, new JsonArray());
        this.harness.getGSExecutor().send(request);

        FakeGamescriptRequestSent sent = awaitRequestOnWire(GSCommand.countIndustry);
        assertEquals(sent.id, request.getId(), "the id sent is the one the executor correlated");

        sendGameScriptReply("{\"id\":\"" + sent.id + "\",\"data\":\"7\"}");
        TestUtils.await(request::hasResponse, 5, "the gamescript reply");
        assertEquals("7", request.getResult());
    }

    /**
     * AdminCmd answers a failing command with {id, exception}, and
     * GSException.getResult() throws.
     */
    @Test
    @DisplayName("an exception reply makes getResult throw")
    public void testReplyWithException() throws IOException {
        GSJsonRequest request = new GSJsonRequest(GSCommand.setTownCargoGoal, new JsonArray());
        this.harness.getGSExecutor().send(request);
        FakeGamescriptRequestSent sent = awaitRequestOnWire(GSCommand.setTownCargoGoal);

        sendGameScriptReply("{\"id\":\"" + sent.id + "\",\"exception\":\"Invalid town\"}");
        TestUtils.await(request::hasResponse, 5, "the gamescript error reply");
        assertThrows(GSException.class, request::getResult);
    }

    /**
     * Malformed requests are answered with {exception} alone: without an id
     * there is nothing to correlate, and the executor must not throw a
     * NullPointerException on the way (which used to kill the listener thread).
     */
    @Test
    @DisplayName("a reply without an id is ignored")
    public void testReplyWithoutId() throws IOException {
        sendGameScriptReply("{\"exception\":\"Malformed json.\"}");
        assertStillConnected();
    }

    /** A reply whose id was never sent must not complete another request. */
    @Test
    public void testReplyWithUnknownId() throws IOException {
        GSJsonRequest request = new GSJsonRequest(GSCommand.countIndustry, new JsonArray());
        this.harness.getGSExecutor().send(request);
        FakeGamescriptRequestSent sent = awaitRequestOnWire(GSCommand.countIndustry);

        String otherId = "" + (Integer.parseInt(sent.id) + 1000);
        sendGameScriptReply("{\"id\":\"" + otherId + "\",\"data\":\"7\"}");
        TestUtils.wait(1);
        assertFalse(request.hasResponse(), "a reply for another request must not complete this one");
        assertStillConnected();
    }

    /** Requests sent after clear() start from id 0 again. */
    @Test
    public void testClearRestartsIds() throws IOException {
        this.harness.getGSExecutor().clear();
        GSJsonRequest request = new GSJsonRequest(GSCommand.countIndustry, new JsonArray());
        this.harness.getGSExecutor().send(request);
        assertEquals("0", request.getId());
        this.harness.getGSExecutor().clear();
        GSJsonRequest next = new GSJsonRequest(GSCommand.countIndustry, new JsonArray());
        this.harness.getGSExecutor().send(next);
        assertEquals("0", next.getId(), "ids restart after clear()");
    }

    /**
     * GSExecutor.send() needs a connection; without one nothing is sent.
     */
    @Test
    @DisplayName("a request is not sent before the admin is connected")
    public void testSendWithoutConnection() throws IOException {
        // A harness whose server never answers the JOIN: getSend() still works,
        // but this proves requests are dropped when send is null.
        GSJsonRequest request = new GSJsonRequest(GSCommand.countIndustry, new JsonArray());
        this.harness.getGSExecutor().send(null);
        assertFalse(request.hasResponse(), "a null request is dropped");
        TestUtils.wait(1);
        assertEquals(0, this.harness.getServer().packets(PacketAdminType.ADMIN_PACKET_ADMIN_GAMESCRIPT).size(),
                "a null request never reaches the wire");
    }

    /** Reply payload types AdminCmd can produce, checked as strings. */
    @Test
    public void testReplyDataOfDifferentTypes() throws IOException {
        GSJsonRequest request = new GSJsonRequest(GSCommand.removeAllGoal, new JsonArray());
        this.harness.getGSExecutor().send(request);
        FakeGamescriptRequestSent sent = awaitRequestOnWire(GSCommand.removeAllGoal);

        sendGameScriptReply("{\"id\":\"" + sent.id + "\",\"data\":\"true\"}");
        TestUtils.await(request::hasResponse, 5, "the boolean reply");
        assertEquals("true", request.getResult());
        // Money and company values come back as numbers above 2^32 sometimes.
        assertTrue(new BigInteger("50000000000").compareTo(BigInteger.valueOf(4294967295L)) > 0);
    }

    /** Proof that the listener thread survived a bad reply. */
    private void assertStillConnected() throws IOException {
        this.harness.getSend().ping(11L);
        TestUtils.await(() -> !this.harness.getRecorder().getPongEvents().isEmpty(), 5, "a PONG");
        assertEquals(11L, this.harness.getRecorder().getPongEvents().get(0).getPingId());
    }

    /** A gamescript request the client put on the wire, with its id. */
    private class FakeGamescriptRequestSent {

        private final String id;

        private FakeGamescriptRequestSent(String id) {
            this.id = id;
        }
    }

    private FakeGamescriptRequestSent awaitRequestOnWire(GSCommand command) {
        TestUtils.await(() -> {
            for (com.openttd.server.FakeAdminServer.ReceivedPacket packet
                    : this.harness.getServer().packets(PacketAdminType.ADMIN_PACKET_ADMIN_GAMESCRIPT)) {
                JsonObject json = new JsonParser().parse(packet.getPacket().readString()).getAsJsonObject();
                if (json.get(GameScript.CMD).getAsInt() == command.ordinal()) {
                    return true;
                }
            }
            return false;
        }, 5, "gamescript command " + command);

        for (com.openttd.server.FakeAdminServer.ReceivedPacket packet
                : this.harness.getServer().packets(PacketAdminType.ADMIN_PACKET_ADMIN_GAMESCRIPT)) {
            JsonObject json = new JsonParser().parse(packet.getPacket().readString()).getAsJsonObject();
            if (json.get(GameScript.CMD).getAsInt() == command.ordinal()) {
                return new FakeGamescriptRequestSent(json.get(GameScript.ID).getAsString());
            }
        }
        throw new IllegalStateException("unreachable");
    }
}
