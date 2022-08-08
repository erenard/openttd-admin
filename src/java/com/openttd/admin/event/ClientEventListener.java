package com.openttd.admin.event;

public interface ClientEventListener extends EventListener {

    void onClientEvent(ClientEvent clientEvent);
}
