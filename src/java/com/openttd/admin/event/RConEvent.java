package com.openttd.admin.event;

import com.openttd.admin.model.Game;

public class RConEvent implements Event<RConEventListener> {

    private final Game openttd;
    private final char color;
    private final String message;

    public RConEvent(Game openttd, char color, String message) {
        this.openttd = openttd;
        this.color = color;
        this.message = message;
    }

    public String getMessage() {
        return message;
    }

    public char getColor() {
        return color;
    }

    public Game getOpenttd() {
        return openttd;
    }

    @Override
    public String toString() {
        return "RCon (" + color + ") " + message;
    }

    @Override
    public void notify(RConEventListener listener) {
        listener.onRConEvent(this);
    }
}
