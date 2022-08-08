package com.openttd.admin;

import java.io.IOException;
import java.util.ArrayList;
import java.util.Collection;
import static org.junit.jupiter.api.Assertions.*;
import org.junit.jupiter.api.*;
import com.openttd.admin.event.PongEvent;
import com.openttd.admin.event.PongEventListener;
import com.openttd.admin.event.RConEndEvent;
import com.openttd.admin.event.RConEndEventListener;
import com.openttd.TestUtils;
import com.openttd.admin.event.GameScriptEvent;
import com.openttd.admin.event.GameScriptEventListener;
import com.openttd.gamescript.GSNewsPaper;
import com.openttd.network.core.Configuration;
import com.openttd.server.OpenttdServerWrapper;

/**
 * Test RCon and Ping/Pong
 *
 * @author erenard
 */
public class OpenttdAdminTest extends OpenttdServerWrapper {

    private final Collection<PongEvent> pongEvents = new ArrayList<>();
    private final Collection<RConEndEvent> rConEndEvents = new ArrayList<>();
    private final Collection<GameScriptEvent> gameScriptEvents = new ArrayList<>();

    private final OpenttdTestAdmin openttdAdmin;

    public OpenttdAdminTest() throws IOException {
        Configuration configuration = new Configuration();
        configuration.password = TestUtils.readSecrets().getProperty("admin_password");
        openttdAdmin = new OpenttdTestAdmin(configuration);
        openttdAdmin.addListener(PongEvent.class, openttdAdmin);
        openttdAdmin.addListener(RConEndEvent.class, openttdAdmin);
        openttdAdmin.addListener(GameScriptEvent.class, openttdAdmin);
    }

    private class OpenttdTestAdmin extends OpenttdAdmin implements PongEventListener, RConEndEventListener, GameScriptEventListener {

        public OpenttdTestAdmin(Configuration configuration) {
            super(configuration);
        }

        @Override
        public void onGameScriptEvent(GameScriptEvent gameScriptEvent) {
            gameScriptEvents.add(gameScriptEvent);
        }

        @Override
        public void onPongEvent(PongEvent pongEvent) {
            pongEvents.add(pongEvent);
        }

        @Override
        public void onRConEndEvent(RConEndEvent rConEndEvent) {
            rConEndEvents.add(rConEndEvent);
        }
    }

    @BeforeEach
    public void setUp() {
        openttdAdmin.startup();
        pongEvents.clear();
        rConEndEvents.clear();
        gameScriptEvents.clear();
        TestUtils.wait(1);
    }

    @AfterEach
    public void tearDown() {
        openttdAdmin.shutdown();
    }

    @Test
    public void testPingPong() {
        var send = openttdAdmin.getSend();
        assertNotNull(send);
        for (long i = 0; i < 5; i++) {
            send.ping(i);
        }
        TestUtils.wait(1);
        assertEquals(5, pongEvents.size());
    }

    @Test
    public void testRCon() {
        var send = openttdAdmin.getSend();
        assertNotNull(send);
        for (long i = 0; i < 5; i++) {
            send.rcon("say message " + i);
            TestUtils.wait(1);
        }
        TestUtils.wait(1);
        assertEquals(5, rConEndEvents.size());
    }

    @Test
    public void testGameScript() {
        var gs = openttdAdmin.getGSExecutor();
        assertNotNull(gs);
        GSNewsPaper broadcast = new GSNewsPaper(GSNewsPaper.NewsType.NT_GENERAL, "News Broadcasting");
        gs.send(broadcast);
        TestUtils.wait(1);
        GSNewsPaper newsPaperCompany0 = new GSNewsPaper((short) 0, GSNewsPaper.NewsType.NT_GENERAL, "News for company 0");
        gs.send(newsPaperCompany0);
        TestUtils.wait(1);
        assertEquals(2, gameScriptEvents.size());
    }
}
