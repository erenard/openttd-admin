package com.openttd.admin;

import com.openttd.admin.event.EventDispatcher;
import com.openttd.admin.event.EventListener;
import com.openttd.admin.model.Game;
import com.openttd.network.admin.NetworkClient;
import com.openttd.network.admin.NetworkClient.Send;
import com.openttd.network.core.Configuration;

public class OpenttdAdmin {
	private final Configuration configuration;
	private NetworkClient networkClient;
	private EventDispatcher eventDispatcher;
	private Game game;

	public OpenttdAdmin(Configuration configuration) {
		this.configuration = configuration;
		// New client layer
		eventDispatcher = new EventDispatcher();
	}

	// Client state
	public boolean isConnected() {
		return networkClient != null && networkClient.isAlive();
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
		networkClient = new NetworkClient(configuration);
		networkClient.setGame(game);
		// Startup
		eventDispatcher.startup();
		networkClient.start();
	}

	/**
	 * Stop the client threads, disconnect it from the game's server.
	 */
	public void shutdown() {
		if (networkClient != null && networkClient.isAlive()) networkClient.shutdown();
		if (eventDispatcher != null && eventDispatcher.isAlive()) eventDispatcher.shutdown();
	}

	/**
	 * Obtain the thread sending messages to the server.
	 * 
	 * @return message sender
	 */
	public Send getSend() {
		return networkClient.getSend();
	}

	/* Event Dispatcher delegation */
	public void addListener(Class eventClass, EventListener listener) {
		eventDispatcher.addListener(eventClass, listener);
	}

	public void removeListener(Class eventClass, EventListener listener) {
		eventDispatcher.removeListener(eventClass, listener);
	}
}
