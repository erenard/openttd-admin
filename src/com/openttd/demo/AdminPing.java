package com.openttd.demo;

import com.openttd.admin.OpenttdAdmin;
import com.openttd.admin.event.EventDispatcher;
import com.openttd.admin.event.GameScriptEvent;
import com.openttd.admin.model.Game;
import com.openttd.gamescript.GSExecutor;
import com.openttd.network.admin.NetworkAdmin;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.openttd.network.core.Configuration;

public class AdminPing {

	private static final Logger log = LoggerFactory.getLogger(AdminPing.class);
	private final Configuration configuration;
	private NetworkAdmin networkAdmin;
	private EventDispatcher eventDispatcher;

	public AdminPing(Configuration configuration) {
        this.configuration = configuration;
		eventDispatcher = new EventDispatcher();
		// New network layer
		networkAdmin = new NetworkAdmin(configuration);
    }
    
    public void start() {
		// Startup
		eventDispatcher.startup();
		networkAdmin.start();
    }
    
    public void ping() {
        networkAdmin.getSend().ping(System.currentTimeMillis());
    }

	/**
	 * Console output demo
	 * 
	 * @param args
	 */
	public static void main(String[] args) {
		Configuration configuration = new Configuration();
        configuration.password = "lepasswd";
		AdminPing simpleAdmin = new AdminPing(configuration);
		simpleAdmin.start();
        simpleAdmin.ping();
		log.info("Openttd admin started");
	}
}
