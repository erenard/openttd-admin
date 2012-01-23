package com.openttd.admin.event;

import java.util.ArrayList;
import java.util.Collection;
import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.TimeUnit;
import java.util.concurrent.locks.ReentrantReadWriteLock;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class EventDispatcher implements Runnable {

	private static final Logger log = LoggerFactory.getLogger(EventDispatcher.class);
	// Events queue
	private BlockingQueue<Event> events = new LinkedBlockingQueue<Event>();
	// Listeners
	private Collection<ChatEventListener> chatEventListeners = new ArrayList<ChatEventListener>();
	private Collection<CompanyEventListener> companyEventListeners = new ArrayList<CompanyEventListener>();
	private Collection<ClientEventListener> clientEventListeners = new ArrayList<ClientEventListener>();
	private Collection<DateEventListener> dateEventListeners = new ArrayList<DateEventListener>();
	//
	private ReentrantReadWriteLock eventListenersLock = new ReentrantReadWriteLock();
	// Simple thread managing
	private boolean running;

	public void addListener(EventListener listener) {
		try {
			eventListenersLock.writeLock().lock();
			if (listener instanceof ChatEventListener) {
				chatEventListeners.add((ChatEventListener) listener);
			}
			if (listener instanceof CompanyEventListener) {
				companyEventListeners.add((CompanyEventListener) listener);
			}
			if (listener instanceof ClientEventListener) {
				clientEventListeners.add((ClientEventListener) listener);
			}
			if (listener instanceof DateEventListener) {
				dateEventListeners.add((DateEventListener) listener);
			}
		} finally {
			eventListenersLock.writeLock().unlock();
		}
	}

	public void removeListener(EventListener listener) {
		try {
			eventListenersLock.writeLock().lock();
			if (listener instanceof ChatEventListener) {
				chatEventListeners.add((ChatEventListener) listener);
			} else if (listener instanceof CompanyEventListener) {
				companyEventListeners.add((CompanyEventListener) listener);
			} else if (listener instanceof ClientEventListener) {
				clientEventListeners.add((ClientEventListener) listener);
			} else {
				log.error("Unknown EventListener.");
			}
		} finally {
			eventListenersLock.writeLock().unlock();
		}
	}

	public void dispatch(Event event) {
		events.offer(event);
	}

	@Override
	public void run() {
		running = true;
		while (this.running) {
			try {
				Event event = events.poll(5, TimeUnit.SECONDS);
				if (event != null) {
					try {
						eventListenersLock.readLock().lock();
						if (event instanceof ChatEvent) {
							ChatEvent chatEvent = (ChatEvent) event;
							for (ChatEventListener chatEventListener : chatEventListeners) {
								chatEventListener.onChatEvent(chatEvent);
							}
						} else if (event instanceof CompanyEvent) {
							CompanyEvent companyEvent = (CompanyEvent) event;
							for (CompanyEventListener companyEventListener : companyEventListeners) {
								companyEventListener.onCompanyEvent(companyEvent);
							}
						} else if (event instanceof DateEvent) {
							DateEvent dateEvent = (DateEvent) event;
							for (DateEventListener dateEventListener : dateEventListeners) {
								dateEventListener.onDateEvent(dateEvent);
							}
						} else if (event instanceof ClientEvent) {
							ClientEvent clientEvent = (ClientEvent) event;
							for (ClientEventListener clientEventListener : clientEventListeners) {
								clientEventListener.onClientEvent(clientEvent);
							}
						}
					} catch (Exception e) {
						log.error(e.getMessage(), e);
					} finally {
						eventListenersLock.readLock().unlock();
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
		events.clear();
		thread = new Thread(this);
		thread.start();
		log.debug("started");
	}

	public void shutdown() {
		log.debug("shutdown...");
		running = false;
		events.clear();
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
