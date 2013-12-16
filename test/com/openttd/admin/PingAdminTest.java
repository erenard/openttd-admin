package com.openttd.admin;

import com.openttd.admin.event.PongEvent;
import com.openttd.admin.event.PongEventListener;
import com.openttd.admin.event.RConEndEvent;
import com.openttd.admin.event.RConEndEventListener;
import com.openttd.network.core.Configuration;
import java.util.HashSet;
import java.util.Set;
import org.junit.After;
import org.junit.Before;
import org.junit.Test;
import org.slf4j.LoggerFactory;

/**
 * Test RCon and Ping/Pong
 * @author edlefou
 */
public class PingAdminTest {

	private static final org.slf4j.Logger log = LoggerFactory.getLogger(PingAdminTest.class);
	private final OpenttdPingAdmin openttdAdmin;
	private final Set<Long> pingIds = new HashSet<Long>();
	private final Set<String> rconCmds = new HashSet<String>();
	private class OpenttdPingAdmin extends OpenttdAdmin implements PongEventListener, RConEndEventListener {
		public OpenttdPingAdmin(Configuration configuration) {
			super(configuration);
		}

		@Override
		public void onPongEvent(PongEvent pongEvent) {
			log.info(pongEvent.toString());
			pingIds.remove(pongEvent.getPingId());
		}

		@Override
		public void onRConEndEvent(RConEndEvent rConEndEvent) {
			log.info(rConEndEvent.toString());
			rconCmds.remove(rConEndEvent.getCommand());
		}
	}
	
	@Test
	public void testPingPong() {
		for(Long pingId : pingIds) openttdAdmin.getSend().ping(pingId);
		int tries = 10;
		while(tries-- > 0 && !pingIds.isEmpty())
			wait(1);
		assert pingIds.isEmpty();
	}

	@Test
	public void testRCon() {
		for(String rconCmd : rconCmds) openttdAdmin.getSend().rcon(rconCmd);
		int tries = 10;
		while(tries-- > 0 && !rconCmds.isEmpty())
			wait(1);
		assert rconCmds.isEmpty();
	}

	public PingAdminTest() {
		Configuration configuration = new Configuration();
		configuration.password = "plop";
		
		openttdAdmin = new OpenttdPingAdmin(configuration);
		openttdAdmin.addListener(PongEvent.class, openttdAdmin);
		openttdAdmin.addListener(RConEndEvent.class, openttdAdmin);
		
		for(long i = 0; i < 50; i++) {
			pingIds.add(i);
			rconCmds.add("say message" + i);
		}
	}

	@Before
	public void setUp() {
		openttdAdmin.startup();
		wait(1);
	}
	
	@After
	public void tearDown() {
		openttdAdmin.shutdown();
		wait(1);
	}
	
	private void wait(int s) {
		try {
			Thread.sleep(s * 1000);
		} catch (InterruptedException ex) {
			log.error(ex.getLocalizedMessage(), ex);
		}
	}
}
