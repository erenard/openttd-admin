package com.openttd.admin;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.openttd.admin.event.DateEvent;
import com.openttd.admin.event.DateEventListener;
import com.openttd.admin.event.GameScriptEvent;
import com.openttd.admin.event.GameScriptEventListener;
import com.openttd.gamescript.GSNewsPaper;
import com.openttd.network.core.Configuration;
import java.util.ArrayList;
import java.util.Collection;

public class GameScript extends OpenttdAdmin implements DateEventListener, GameScriptEventListener {

	private static final Logger log = LoggerFactory.getLogger(GameScript.class);

	public GameScript(Configuration configuration) {
		super(configuration);
	}

	Collection<GSNewsPaper> newsPapers = new ArrayList<GSNewsPaper>();
	
	@Override
	public void onGameScriptEvent(GameScriptEvent gameScriptEvent) {
		log.info(gameScriptEvent.toString());
		for(GSNewsPaper newsPaper : newsPapers) {
			if(newsPaper.hasResponse()) {
				log.info(newsPaper.toString() + " " + newsPaper.getResult());
			} else {
				log.info(newsPaper.toString());
			}
		}
	}
	private int i = 0;

	@Override
	public void onDateEvent(DateEvent dateEvent) {
		if (i == 0) {
			GSNewsPaper newsPaper = new GSNewsPaper((short) -1, GSNewsPaper.NewsType.NT_GENERAL, "News Broadcasting !");
			getGSExecutor().send(newsPaper);
			newsPapers.add(newsPaper);
		} else if (i == 1) {
			GSNewsPaper newsPaper = new GSNewsPaper((short) 0, GSNewsPaper.NewsType.NT_GENERAL, "News for company 0");
			getGSExecutor().send(newsPaper);
			newsPapers.add(newsPaper);
		} else if (i > 1) {
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

		GameScript gameScript = new GameScript(configuration);
		gameScript.addListener(DateEvent.class, gameScript);
		gameScript.addListener(GameScriptEvent.class, gameScript);
		gameScript.startup();
		log.info("Openttd admin started");
	}
}
