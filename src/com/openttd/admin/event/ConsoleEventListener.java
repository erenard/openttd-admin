package com.openttd.admin.event;

public interface ConsoleEventListener extends EventListener {
	void onConsoleEvent(ConsoleEvent consoleEvent);
}
