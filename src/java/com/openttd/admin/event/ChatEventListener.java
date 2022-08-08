package com.openttd.admin.event;

public interface ChatEventListener extends EventListener {

    void onChatEvent(ChatEvent chatEvent);
}
