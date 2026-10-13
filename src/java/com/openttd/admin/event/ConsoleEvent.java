package com.openttd.admin.event;

import com.openttd.admin.model.Game;

public class ConsoleEvent implements Event<ConsoleEventListener> {

	private final Game openttd;
	private final String origin;
	private final String message;

	public ConsoleEvent(Game openttd, String origin, String message) {
		this.openttd = openttd;
		this.origin = origin;
		this.message = message;
	}

	public String getMessage() {
		return message;
	}

	public String getOrigin() {
		return origin;
	}

	public Game getOpenttd() {
		return openttd;
	}

	@Override
	public String toString() {
		return "Console (" + origin + ") " + message;
	}

	@Override
	public void notify(ConsoleEventListener listener) {
		listener.onConsoleEvent(this);
	}
}
