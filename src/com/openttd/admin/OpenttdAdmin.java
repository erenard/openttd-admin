package com.openttd.admin;

import com.openttd.admin.event.EventDispatcher;
import com.openttd.admin.event.EventListener;
import com.openttd.admin.event.GameScriptEvent;
import com.openttd.admin.model.Game;
import com.openttd.gamescript.GSExecutor;
import com.openttd.network.admin.NetworkAdmin;
import com.openttd.network.admin.NetworkAdminSender;
import com.openttd.network.core.Configuration;

public class OpenttdAdmin {
	private final Configuration configuration;
	private NetworkAdmin networkAdmin;
	private GSExecutor executor;
	private EventDispatcher eventDispatcher;
	private Game game;

	public OpenttdAdmin(Configuration configuration) {
		this.configuration = configuration;
		// New client layer
		eventDispatcher = new EventDispatcher();
		executor = new GSExecutor(this);
		eventDispatcher.addListener(GameScriptEvent.class, executor);
	}

	// Client state
	public boolean isConnected() {
		return networkAdmin != null && networkAdmin.isAlive();
	}

	/**
	 * Start the client and the networking threads
	 */
	public void startup() {
		// Safety shutdown
		shutdown();
		// New client model
		game = new Game(eventDispatcher);
		// New network layer
		networkAdmin = new NetworkAdmin(configuration);
		networkAdmin.setGame(game);
		// Startup
		eventDispatcher.startup();
		networkAdmin.start();
		executor.clear();
	}

	/**
	 * Stop the client threads, disconnect it from the game's server.
	 */
	public void shutdown() {
		if (networkAdmin != null && networkAdmin.isAlive()) networkAdmin.shutdown();
		if (eventDispatcher != null && eventDispatcher.isAlive()) eventDispatcher.shutdown();
	}

	/**
	 * Obtain the thread sending messages to the server.
	 *
	 * @return message sender
	 */
	public NetworkAdminSender getSend() {
		return networkAdmin.getSend();
	}

	/**
	 * Obtain the thread calling gamescript on the server.
	 *
	 * @return gamescript executor
	 */
	public GSExecutor getGSExecutor() {
		return executor;
	}

	/* Event Dispatcher delegation */
	public void addListener(Class eventClass, EventListener listener) {
		eventDispatcher.addListener(eventClass, listener);
	}

	public void removeListener(Class eventClass, EventListener listener) {
		eventDispatcher.removeListener(eventClass, listener);
	}
}
