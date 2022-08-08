package com.openttd.admin.event;

import java.util.ArrayList;
import java.util.Collection;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.locks.ReentrantReadWriteLock;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Thread doing the event dispatching.
 */
public class EventDispatcher implements Runnable {

    private static final Logger log = LoggerFactory.getLogger(EventDispatcher.class);

    private class EventMessage {

        Event event;
        Class eventClass;

        public EventMessage(Event event, Class eventClass) {
            this.event = event;
            this.eventClass = eventClass;
        }
    }
    // Events queue
    private BlockingQueue<EventMessage> eventMessages = new LinkedBlockingQueue<>();
    // Simple thread managing
    private boolean running;

    // Listeners
    private ReentrantReadWriteLock lock = new ReentrantReadWriteLock();
    private Map<Class, Collection<EventListener>> listenersByEventClass = new HashMap<>();

    public void addListener(Class eventClass, EventListener listener) {
        try {
            lock.writeLock().lock();
            if (!listenersByEventClass.containsKey(eventClass)) {
                listenersByEventClass.put(eventClass, new ArrayList<>());
            }
            Collection<EventListener> listeners = listenersByEventClass.get(eventClass);
            listeners.add(listener);
        } finally {
            lock.writeLock().unlock();
        }
    }

    public void removeListener(Class eventClass, EventListener listener) {
        try {
            lock.writeLock().lock();
            if (!listenersByEventClass.containsKey(eventClass)) {
                listenersByEventClass.put(eventClass, new ArrayList<>());
            }
            Collection<EventListener> listeners = listenersByEventClass.get(eventClass);
            listeners.remove(listener);
        } finally {
            lock.writeLock().unlock();
        }
    }

    public <L> void dispatch(Event<L> event) {
        @SuppressWarnings("unchecked")
        Class<Event<L>> evtClass = (Class<Event<L>>) event.getClass();
        eventMessages.offer(new EventMessage(event, evtClass));
    }

    @Override
    public void run() {
        running = true;
        while (this.running) {
            try {
                EventMessage eventClass = eventMessages.poll(5, TimeUnit.SECONDS);
                if (eventClass != null) {
                    try {
                        lock.readLock().lock();
                        Collection<EventListener> listeners = listenersByEventClass.get(eventClass.eventClass);
                        if (listeners != null) {
                            for (EventListener eventListener : listeners) {
                                eventClass.event.notify(eventListener);
                            }
                        }
                    } catch (Exception e) {
                        log.error(e.getMessage(), e);
                    } finally {
                        lock.readLock().unlock();
                    }
                }
            } catch (InterruptedException e) {
                log.error(e.getMessage(), e);
            }
        }
    }

    private Thread thread = null;

    public void startup() {
        log.debug("startup...");
        running = true;
        eventMessages.clear();
        thread = new Thread(this);
        thread.start();
        log.debug("started");
    }

    public void shutdown() {
        log.debug("shutdown...");
        running = false;
        eventMessages.clear();
        try {
            thread.join(5000);
        } catch (InterruptedException e) {
            log.error(e.getMessage(), e);
            thread.interrupt();
        }
        log.debug("stopped");
    }

    public final boolean isAlive() {
        return thread != null && thread.isAlive();
    }
}
