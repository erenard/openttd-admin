package com.openttd.admin.event;

import com.openttd.admin.model.Game;

public class GameScriptEvent implements Event<GameScriptEventListener> {

    private final Game openttd;
    private final String message;

    public GameScriptEvent(Game openttd, String message) {
        this.openttd = openttd;
        this.message = message;
    }

    public String getMessage() {
        return message;
    }

    public Game getOpenttd() {
        return openttd;
    }

    @Override
    public String toString() {
        return "GameScript " + message;
    }

    @Override
    public void notify(GameScriptEventListener listener) {
        listener.onGameScriptEvent(this);
    }
}
