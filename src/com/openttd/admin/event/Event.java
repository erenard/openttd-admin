package com.openttd.admin.event;

public interface Event<L> {
	void notify(final L listener);
}
