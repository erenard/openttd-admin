package com.openttd.admin.event;

public interface PongEventListener extends EventListener {

    void onPongEvent(PongEvent pongEvent);
}
