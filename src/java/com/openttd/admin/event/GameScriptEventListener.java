package com.openttd.admin.event;

public interface GameScriptEventListener extends EventListener {

    void onGameScriptEvent(GameScriptEvent gameScriptEvent);
}
