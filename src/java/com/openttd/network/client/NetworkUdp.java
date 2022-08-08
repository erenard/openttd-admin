package com.openttd.network.client;

import java.io.IOException;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.openttd.network.constant.PacketUdpType;
import com.openttd.network.core.Packet;
import com.openttd.network.core.Socket;

public class NetworkUdp {

    private static final Logger log = LoggerFactory.getLogger(NetworkUdp.class);

    private final Socket socket;
    private final Updater updater;

    public NetworkUdp(String host, Integer port) throws IOException {
        this.socket = Socket.newUdpSocket(host, port);
        this.updater = new Updater();
    }

    private boolean serverFound = false;

    public boolean isServerFound() {
        return serverFound;
    }

    public void start() {
        updater.start();
        try {
            updater.join();
        } catch (InterruptedException e) {
            log.error(e.getMessage());
        }
    }

    private class Updater extends Thread {

        @Override
        public void run() {
            try {
                socket.sendEmptyPacket(PacketUdpType.PACKET_UDP_CLIENT_FIND_SERVER);
                receive:
                while (true) {
                    List<Packet> packets = socket.receive();
                    for (Packet packet : packets) {
                        PacketUdpType packetUdpType = PacketUdpType.valueOf(packet.readUint8());
                        log.debug(packetUdpType.toString());
                        switch (packetUdpType) {
                            case PACKET_UDP_SERVER_RESPONSE: {
                                serverFound = true;
                                break receive;
                            }
                        }
                    }
                }
            } catch (IOException e) {
                log.error(e.getMessage(), e);
            }
        }
    };

    /**
     * main for test purpose
     *
     * @throws IOException
     */
    public static void main(String[] args) throws IOException {
        log.info("Testing UdpClient");
        NetworkUdp udpClient = new NetworkUdp("localhost", 3979);
        log.info("Finding server...");
        udpClient.start();
        log.info(udpClient.isServerFound() ? "Server found" : "Server not found");
    }
}
