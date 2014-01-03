/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */
package com.openttd.client;

import com.openttd.network.client.NetworkClient;
import com.openttd.network.core.Configuration;

/**
 *
 * @author erenard
 */
public class OpenttdClient {
	private final Configuration configuration;
	private NetworkClient networkClient;

	public OpenttdClient(Configuration configuration) {
		this.configuration = configuration;
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
		// New network layer
		networkClient = new NetworkClient(configuration);
		// Startup
		networkClient.start();
	}

	/**
	 * Stop the client threads, disconnect it from the game's server.
	 */
	public void shutdown() {
		if (networkClient != null && networkClient.isAlive()) networkClient.shutdown();
	}

}
