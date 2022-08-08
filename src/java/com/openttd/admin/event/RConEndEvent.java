package com.openttd.admin.event;

import com.openttd.admin.model.Game;

public class RConEndEvent implements Event<RConEndEventListener> {

    private final Game openttd;
    private final String command;

    public RConEndEvent(Game openttd, String command) {
        this.openttd = openttd;
        this.command = command;
    }

    public String getCommand() {
        return command;
    }

    public Game getOpenttd() {
        return openttd;
    }

    @Override
    public String toString() {
        return "RCon ended :" + command;
    }

    @Override
    public void notify(RConEndEventListener listener) {
        listener.onRConEndEvent(this);
    }
}
