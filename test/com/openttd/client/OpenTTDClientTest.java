package com.openttd.client;

import com.openttd.network.client.NetworkClient;
import com.openttd.network.core.Configuration;
import org.junit.Test;
import org.slf4j.LoggerFactory;

/**
 *
 * @author edlefou
 */
public class OpenTTDClientTest {

	private static final org.slf4j.Logger log = LoggerFactory.getLogger(OpenTTDClientTest.class);
	private final Configuration configuration;
	
	public OpenTTDClientTest() {
		configuration = new Configuration();
		configuration.name = "test-client";
		configuration.password = "plop";
	}
	
	@Test
	public void testSimpleAdmin() {
		NetworkClient networkClient = new NetworkClient(configuration);
		networkClient.start();
		try {
			Thread.sleep(60000);
		} catch (InterruptedException ex) {
			log.error(ex.getLocalizedMessage(), ex);
		}
		networkClient.shutdown();
	}

}
