package com.openttd.admin;

import com.openttd.admin.OpenttdAdmin;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.openttd.admin.event.DateEvent;
import com.openttd.admin.event.DateEventListener;
import com.openttd.admin.event.GameScriptEvent;
import com.openttd.admin.event.GameScriptEventListener;
import com.openttd.network.constant.GameScript.GoalType;
import com.openttd.network.core.Configuration;

public class GameScript extends OpenttdAdmin implements DateEventListener, GameScriptEventListener {

	private static final Logger log = LoggerFactory.getLogger(GameScript.class);

	public GameScript(Configuration configuration) {
		super(configuration);
	}

	@Override
	public void onGameScriptEvent(GameScriptEvent gameScriptEvent) {
		log.info(gameScriptEvent.toString());
	}

	private int i = 0;
	
	@Override
	public void onDateEvent(DateEvent dateEvent) {
		if(i < 3) {
			getSend().addGlobalGoal("Global goal " + i, GoalType.GT_COMPANY, 0);
		} else if(i == 4) {
			getSend().removeAllGoal();
		} else if(i == 5) {
                        getSend().newsBroadcast("News Broadcasting !");
		} else if(i > 5) {
			this.shutdown();
		}
		i++;
	}

	/**
	 * Console output demo
	 * 
	 * @param args
	 */
	public static void main(String[] args) {
		Configuration configuration = new Configuration();
		
		try {
			String url = args[0];
			configuration.host = url.split(":")[0];
			configuration.adminPort = new Integer(url.split(":")[1]);
			configuration.password = args[1];
		} catch (Exception e) {
			log.error(e.getMessage(), e);
			log.error("Usage: java -jar openttd-admin.jar localhost:3977 admin_password");
			log.error("See openttd.cfg to set your server's admin_password first !");
			System.exit(0);
		}

		GameScript simpleAdmin = new GameScript(configuration);
		simpleAdmin.addListener(DateEvent.class, simpleAdmin);
		simpleAdmin.addListener(GameScriptEvent.class, simpleAdmin);
		simpleAdmin.startup();
		log.info("Openttd admin started");
	}
}
