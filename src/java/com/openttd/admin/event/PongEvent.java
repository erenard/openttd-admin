package com.openttd.admin.event;

import com.openttd.admin.model.Game;

public class PongEvent implements Event<PongEventListener> {

	private final Game openttd;
	private final long pingId;

	public PongEvent(Game openttd, long pingId) {
		this.openttd = openttd;
		this.pingId = pingId;
	}

	public long getPingId() {
		return pingId;
	}

	public Game getOpenttd() {
		return openttd;
	}

	@Override
	public String toString() {
		return "Pong (" + pingId + ")";
	}

	@Override
	public void notify(PongEventListener listener) {
		listener.onPongEvent(this);
	}
}
