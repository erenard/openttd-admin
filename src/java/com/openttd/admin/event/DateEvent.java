package com.openttd.admin.event;

import com.openttd.admin.model.Game;
import java.util.Calendar;

public class DateEvent implements Event<DateEventListener> {

	private final Game openttd;

	public DateEvent(Game openttd) {
		this.openttd = openttd;
	}

	public Game getOpenttd() {
		return openttd;
	}

	@Override
	public void notify(DateEventListener listener) {
		listener.onDateEvent(this);
	}

	@Override
	public String toString() {
		Calendar date = openttd.getDate();
		StringBuilder sb = new StringBuilder("DateEvent ");
		sb.append(date.get(Calendar.DAY_OF_MONTH)).append("-");
		sb.append(date.get(Calendar.MONTH)).append("-");
		sb.append(date.get(Calendar.YEAR));
		return sb.toString();
	}
}
