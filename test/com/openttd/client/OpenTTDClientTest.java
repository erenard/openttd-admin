package com.openttd.client;

import com.openttd.demo.CLIUtil;
import com.openttd.network.client.NetworkClient;
import com.openttd.network.core.Configuration;
import java.io.IOException;
import org.junit.Test;
import org.slf4j.LoggerFactory;

/**
 *
 * @author erenard
 */
public class OpenTTDClientTest {

	private static final org.slf4j.Logger log = LoggerFactory.getLogger(OpenTTDClientTest.class);
	private final Configuration configuration;
	
	public OpenTTDClientTest() throws IOException {
		configuration = new Configuration();
		configuration.name = CLIUtil.readTestProperties().getProperty("clientName");
	}
	
	@Test
	public void testSimpleAdmin() {
		NetworkClient networkClient = new NetworkClient(configuration);
		log.info("Starting client and joining");
		networkClient.start();
		log.info("Staying 1 minute");
		CLIUtil.wait(60);
		log.info("Leaving");
		networkClient.shutdown();
	}
}
