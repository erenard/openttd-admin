package com.openttd.server;

import java.util.ArrayList;
import java.util.List;

import com.openttd.admin.OpenttdAdmin;
import com.openttd.admin.event.ChatEvent;
import com.openttd.admin.event.ChatEventListener;
import com.openttd.admin.event.ClientEvent;
import com.openttd.admin.event.ClientEventListener;
import com.openttd.admin.event.CompanyEvent;
import com.openttd.admin.event.CompanyEventListener;
import com.openttd.admin.event.ConsoleEvent;
import com.openttd.admin.event.ConsoleEventListener;
import com.openttd.admin.event.DateEvent;
import com.openttd.admin.event.DateEventListener;
import com.openttd.admin.event.GameScriptEvent;
import com.openttd.admin.event.GameScriptEventListener;
import com.openttd.admin.event.PongEvent;
import com.openttd.admin.event.PongEventListener;
import com.openttd.admin.event.RConEndEvent;
import com.openttd.admin.event.RConEndEventListener;
import com.openttd.admin.event.RConEvent;
import com.openttd.admin.event.RConEventListener;

/**
 * Collects every functional event the admin facade dispatches, so tests can
 * assert on all of them instead of only the three that were covered before.
 */
public class EventRecorder implements ChatEventListener, ClientEventListener, CompanyEventListener,
        ConsoleEventListener, DateEventListener, GameScriptEventListener, PongEventListener,
        RConEventListener, RConEndEventListener {

    private final List<ChatEvent> chatEvents = new ArrayList<>();
    private final List<ClientEvent> clientEvents = new ArrayList<>();
    private final List<CompanyEvent> companyEvents = new ArrayList<>();
    private final List<ConsoleEvent> consoleEvents = new ArrayList<>();
    private final List<DateEvent> dateEvents = new ArrayList<>();
    private final List<GameScriptEvent> gameScriptEvents = new ArrayList<>();
    private final List<PongEvent> pongEvents = new ArrayList<>();
    private final List<RConEvent> rConEvents = new ArrayList<>();
    private final List<RConEndEvent> rConEndEvents = new ArrayList<>();

    /** Drop everything collected so far, so a test asserts on its own packets only. */
    public void clear() {
        chatEvents.clear();
        clientEvents.clear();
        companyEvents.clear();
        consoleEvents.clear();
        dateEvents.clear();
        gameScriptEvents.clear();
        pongEvents.clear();
        rConEvents.clear();
        rConEndEvents.clear();
    }

    public void registerWith(OpenttdAdmin admin) {
        admin.addListener(ChatEvent.class, this);
        admin.addListener(ClientEvent.class, this);
        admin.addListener(CompanyEvent.class, this);
        admin.addListener(ConsoleEvent.class, this);
        admin.addListener(DateEvent.class, this);
        admin.addListener(GameScriptEvent.class, this);
        admin.addListener(PongEvent.class, this);
        admin.addListener(RConEvent.class, this);
        admin.addListener(RConEndEvent.class, this);
    }

    public List<ChatEvent> getChatEvents() {
        return chatEvents;
    }

    public List<ClientEvent> getClientEvents() {
        return clientEvents;
    }

    public List<CompanyEvent> getCompanyEvents() {
        return companyEvents;
    }

    public List<ConsoleEvent> getConsoleEvents() {
        return consoleEvents;
    }

    public List<DateEvent> getDateEvents() {
        return dateEvents;
    }

    public List<GameScriptEvent> getGameScriptEvents() {
        return gameScriptEvents;
    }

    public List<PongEvent> getPongEvents() {
        return pongEvents;
    }

    public List<RConEvent> getRConEvents() {
        return rConEvents;
    }

    public List<RConEndEvent> getRConEndEvents() {
        return rConEndEvents;
    }

    @Override
    public void onChatEvent(ChatEvent chatEvent) {
        chatEvents.add(chatEvent);
    }

    @Override
    public void onClientEvent(ClientEvent clientEvent) {
        clientEvents.add(clientEvent);
    }

    @Override
    public void onCompanyEvent(CompanyEvent companyEvent) {
        companyEvents.add(companyEvent);
    }

    @Override
    public void onConsoleEvent(ConsoleEvent consoleEvent) {
        consoleEvents.add(consoleEvent);
    }

    @Override
    public void onDateEvent(DateEvent dateEvent) {
        dateEvents.add(dateEvent);
    }

    @Override
    public void onGameScriptEvent(GameScriptEvent gameScriptEvent) {
        gameScriptEvents.add(gameScriptEvent);
    }

    @Override
    public void onPongEvent(PongEvent pongEvent) {
        pongEvents.add(pongEvent);
    }

    @Override
    public void onRConEvent(RConEvent rConEvent) {
        rConEvents.add(rConEvent);
    }

    @Override
    public void onRConEndEvent(RConEndEvent rConEndEvent) {
        rConEndEvents.add(rConEndEvent);
    }
}
