package com.openttd.admin.event;

import com.openttd.admin.model.Game;

public class DateEvent implements Event {

	private final Game openttd;

	public DateEvent(Game openttd) {
		this.openttd = openttd;
	}

	public Game getOpenttd() {
		return openttd;
	}
}
