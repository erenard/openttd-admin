package com.openttd.server;

import java.io.IOException;

import com.openttd.TestUtils;
import com.openttd.admin.OpenttdAdmin;
import com.openttd.admin.model.Game;
import com.openttd.gamescript.GSExecutor;
import com.openttd.network.admin.GameInfo;
import com.openttd.network.admin.NetworkAdminSender;
import com.openttd.network.core.Configuration;

/**
 * Admin facade wired to a FakeAdminServer: no OpenTTD process, so the whole
 * admin protocol is exercisable in a unit test.
 *
 * Game.updateAdmin registers update frequencies and polls when the WELCOME
 * packet arrives, exactly as it does against a real server.
 */
public class AdminHarness implements AutoCloseable {

    public static final String PASSWORD = "test-password";
    public static final String ADMIN_NAME = "openttd-admin-test";

    private final FakeAdminServer server;
    private final Configuration configuration;
    private final OpenttdAdmin admin;
    private final EventRecorder recorder;

    public AdminHarness() throws IOException {
        this(PASSWORD, ADMIN_NAME);
    }

    public AdminHarness(String password, String name) throws IOException {
        this.server = new FakeAdminServer();
        this.configuration = new Configuration();
        this.configuration.password = password;
        this.configuration.name = name;
        this.configuration.host = "localhost";
        this.configuration.adminPort = this.server.getPort();
        this.admin = new OpenttdAdmin(this.configuration);
        this.recorder = new EventRecorder();
        this.recorder.registerWith(this.admin);
        this.admin.startup();
        this.server.handshake();
        // The game model only exists once a packet has been decoded, so a
        // missing model means "not connected yet", not a protocol failure.
        TestUtils.await(() -> {
            try {
                GameInfo gameInfo = this.admin.getGame().getGameInfo();
                return gameInfo != null && gameInfo.getServerName() != null;
            } catch (NullPointerException notConnectedYet) {
                return false;
            }
        }, 10, "the WELCOME packet");
    }

    public FakeAdminServer getServer() {
        return server;
    }

    public Configuration getConfiguration() {
        return configuration;
    }

    public OpenttdAdmin getAdmin() {
        return admin;
    }

    public EventRecorder getRecorder() {
        return recorder;
    }

    public Game getGame() {
        return admin.getGame();
    }

    public NetworkAdminSender getSend() {
        TestUtils.await(() -> admin.getSend() != null, 10, "the admin sender thread");
        return admin.getSend();
    }

    public GSExecutor getGSExecutor() {
        return admin.getGSExecutor();
    }

    @Override
    public void close() {
        this.admin.shutdown();
        this.server.close();
    }
}
