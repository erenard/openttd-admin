package com.openttd.admin;

import com.openttd.demo.OpenttdAdminConsole;
import com.openttd.admin.event.ChatEvent;
import com.openttd.admin.event.ClientEvent;
import com.openttd.admin.event.CompanyEvent;
import com.openttd.admin.event.DateEvent;
import com.openttd.admin.event.GameScriptEvent;
import com.openttd.demo.CLIUtil;
import com.openttd.network.core.Configuration;
import java.io.IOException;
import org.junit.Test;
import org.slf4j.LoggerFactory;

/**
 * Unit Test
 * @author erenard
 */
public class OpenTTDAdminTest {

	private static final org.slf4j.Logger log = LoggerFactory.getLogger(OpenTTDAdminTest.class);
	private final Configuration configuration;
	
	public OpenTTDAdminTest() throws IOException {
		configuration = new Configuration();
		configuration.password = CLIUtil.readTestProperties().getProperty("password");
	}
	
	@Test
	public void testSimpleAdmin() {
		OpenttdAdminConsole simpleAdmin = new OpenttdAdminConsole(configuration);
		simpleAdmin.addListener(DateEvent.class, simpleAdmin);
		simpleAdmin.addListener(ChatEvent.class, simpleAdmin);
		simpleAdmin.addListener(ClientEvent.class, simpleAdmin);
		simpleAdmin.addListener(CompanyEvent.class, simpleAdmin);
		simpleAdmin.startup();
		try {
			Thread.sleep(10000);
		} catch (InterruptedException ex) {
			log.error(ex.getLocalizedMessage(), ex);
		}
		simpleAdmin.shutdown();
	}

	@Test
	public void testGameScript() {
		GameScript gameScript = new GameScript(configuration);
		gameScript.addListener(DateEvent.class, gameScript);
		gameScript.addListener(GameScriptEvent.class, gameScript);
		gameScript.startup();
		while(gameScript.isConnected()) {
			try {
				Thread.sleep(1000);
			} catch (InterruptedException ex) {
				log.error(ex.getLocalizedMessage(), ex);
			}
		}
	}
}
