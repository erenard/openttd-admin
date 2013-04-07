package com.openttd.admin.event;

import com.openttd.admin.model.Game;

public class ClientEvent implements Event<ClientEventListener> {

	public enum Action {
		CREATE,
		UPDATE,
		DELETE
	};

	private final Game openttd;
	private final int clientId;
	private final Action action;

	public ClientEvent(Game openttd, long clientId, Action action) {
		this.openttd = openttd;
		this.clientId = (int) clientId;
		this.action = action;
	}

	public int getClientId() {
		return clientId;
	}

	public Game getOpenttd() {
		return openttd;
	}

	public Action getAction() {
		return action;
	}

	@Override
	public String toString() {
		return action + ":" + clientId;
	}

	@Override
	public void notify(ClientEventListener listener) {
		listener.onClientEvent(this);
	}
}
