package com.openttd.network.core;

import java.util.concurrent.BlockingQueue;
import java.util.concurrent.LinkedBlockingQueue;
import java.util.concurrent.TimeUnit;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public abstract class NetworkSender implements Runnable {

    private static final Logger log = LoggerFactory.getLogger(NetworkSender.class);
    /* Network */
    private final Socket socket;

    public NetworkSender(Socket socket) {
        this.socket = socket;
    }

    /* Thread */
    private boolean running;
    private Thread senderThread;

    public void startup() {
        this.running = true;
        this.senderThread = new Thread(this, "OpenttdPacketSender");
        this.senderThread.start();
    }

    public void shutdown() {
        this.running = false;
        if (senderThread != null && senderThread.isAlive()) {
            this.senderThread.interrupt();
        }
    }

    @Override
    public void run() {
        while (running && socket != null && socket.isOpen()) {
            try {
                Packet packet = queue.poll(5, TimeUnit.SECONDS);
                if (packet != null && socket != null) {
                    socket.send(packet);
                    if (log.isDebugEnabled()) {
                        this.debugPacket(packet);
                    }
                    // Clean shutdown
                    if (this.isQuitPacket(packet)) {
                        this.running = false;
                    } else {
                        Thread.sleep(100);
                    }
                }
            } catch (InterruptedException e) {
                log.error(e.getMessage(), e);
            }
        }
    }

    protected abstract boolean isQuitPacket(Packet packet);

    protected abstract void debugPacket(Packet packet);

    /* OpenTTD */
    protected BlockingQueue<Packet> queue = new LinkedBlockingQueue<>();
}
