package com.openttd.client;

import com.openttd.TestUtils;
import com.openttd.network.client.NetworkClient;
import com.openttd.network.core.Configuration;
import java.io.IOException;
import org.junit.jupiter.api.Test;
import org.slf4j.LoggerFactory;

/**
 *
 * @author erenard
 */
public class OpenttdClientTest {

    private static final org.slf4j.Logger log = LoggerFactory.getLogger(OpenttdClientTest.class);
    private final Configuration configuration;

    public OpenttdClientTest() throws IOException {
        configuration = new Configuration();
        configuration.name = "Client test name";
    }

    @Test
    public void testSimpleClient() {
        NetworkClient networkClient = new NetworkClient(configuration);
        log.info("Starting client and joining");
        networkClient.start();
        log.info("Staying 5 seconds");
        TestUtils.wait(5);
        log.info("Leaving");
        networkClient.shutdown();
    }
}
