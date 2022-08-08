package com.openttd.admin.event;

import com.openttd.admin.model.Game;

public class ChatEvent implements Event<ChatEventListener> {

    private final Game openttd;
    private final int clientId;
    private final String message;

    public ChatEvent(Game openttd, long clientId, String message) {
        this.openttd = openttd;
        this.clientId = (int) clientId;
        this.message = message;
    }

    public String getMessage() {
        return message;
    }

    public int getClientId() {
        return clientId;
    }

    public Game getOpenttd() {
        return openttd;
    }

    @Override
    public String toString() {
        return clientId + " says " + message;
    }

    @Override
    public void notify(ChatEventListener listener) {
        listener.onChatEvent(this);
    }
}
