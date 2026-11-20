package com.openttd.server;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.channels.ServerSocketChannel;
import java.nio.channels.SocketChannel;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.LinkedBlockingQueue;

import com.openttd.TestUtils;
import com.openttd.network.constant.TcpAdmin.PacketAdminType;
import com.openttd.network.constant.TcpAdmin.PacketServerType;
import com.openttd.network.core.Packet;

/**
 * Minimal admin network server: accepts one connection, reads the packets the
 * client sends and scripts the packets a server would send. It frames packets
 * the way src/network/core/packet.cpp does (uint16 length including the header,
 * then uint8 packet type, then payload), so NetworkAdmin's Socket.receive() sees
 * exactly what OpenTTD would send.
 *
 * No OpenTTD process is needed, which makes the admin protocol decodable in a
 * unit test - including the packets a real server never sends to an
 * ADMIN_PACKET_ADMIN_JOIN client (FULL, BANNED, CMD_LOGGING, AUTH_REQUEST, ...).
 */
public class FakeAdminServer implements AutoCloseable {

    /** A packet sent by the admin client, framed and ready to decode. */
    public class ReceivedPacket {

        private final int wireTypeId;
        private final Packet packet;

        private ReceivedPacket(int wireTypeId, Packet packet) {
            this.wireTypeId = wireTypeId;
            this.packet = packet;
        }

        public int getWireTypeId() {
            return wireTypeId;
        }

        public PacketAdminType getType() {
            return PacketAdminType.valueOf(wireTypeId);
        }

        /**
         * @return packet positioned after the header, i.e. the payload can be
         * read with Packet's read* methods
         */
        public Packet getPacket() {
            packet.prepareToRead();
            packet.readUint8();
            return packet;
        }

        @Override
        public String toString() {
            return String.valueOf(getType()) + " (" + wireTypeId + ")";
        }
    }

    private final ServerSocketChannel listener;
    private final LinkedBlockingQueue<ReceivedPacket> received = new LinkedBlockingQueue<>();
    private final Object connectionLock = new Object();
    private SocketChannel connection;
    private int connectionCount;
    private volatile boolean clientDisconnected;
    private volatile boolean running = true;
    private final Thread acceptor;

    public FakeAdminServer() throws IOException {
        this.listener = ServerSocketChannel.open();
        this.listener.bind(new InetSocketAddress("localhost", 0));
        this.acceptor = new Thread(this::acceptLoop, "fake-admin-server");
        this.acceptor.setDaemon(true);
        this.acceptor.start();
    }

    public int getPort() {
        return this.listener.socket().getLocalPort();
    }

    /** How many admin connections have been accepted so far. */
    public int getConnectionCount() {
        synchronized (this.connectionLock) {
            return this.connectionCount;
        }
    }

    /** Whether the admin client closed its connection (or hit EOF). */
    public boolean isClientDisconnected() {
        return this.clientDisconnected;
    }

    private void acceptLoop() {
        try {
            while (this.running) {
                SocketChannel channel = this.listener.accept();
                synchronized (this.connectionLock) {
                    this.connection = channel;
                    this.connectionCount++;
                    this.clientDisconnected = false;
                    this.connectionLock.notifyAll();
                }
                try {
                    while (this.running) {
                        ReceivedPacket packet = readPacket(channel);
                        if (packet == null) {
                            break;
                        }
                        this.received.add(packet);
                        answer(packet);
                    }
                } catch (IOException ignored) {
                    // The client's connection broke: accept the next one, the
                    // client reconnects every 5 s.
                }
                synchronized (this.connectionLock) {
                    // Never write to a closed connection: send() waits for the
                    // client's next one instead.
                    this.connection = null;
                }
                this.clientDisconnected = true;
            }
        } catch (IOException ignored) {
            // close() closed the listener.
        }
    }

    /**
     * Answers that a real server gives automatically: ADMIN_PACKET_ADMIN_PING is
     * answered with ADMIN_PACKET_SERVER_PONG echoing the identifier
     * (Receive_ADMIN_PING in network_admin.cpp).
     */
    private void answer(ReceivedPacket packet) throws IOException {
        if (packet.getWireTypeId() != PacketAdminType.ADMIN_PACKET_ADMIN_PING.ordinal()) {
            return;
        }
        long pingId = packet.getPacket().readUint32();
        send(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_PONG).writeUint32(pingId));
    }

    private ReceivedPacket readPacket(SocketChannel channel) throws IOException {
        byte[] header = readExactly(channel, 2);
        if (header == null) {
            return null;
        }
        int length = (header[0] & 0xFF) | ((header[1] & 0xFF) << 8);
        byte[] body = readExactly(channel, length - 2);
        if (body == null) {
            return null;
        }
        ByteBuffer buffer = ByteBuffer.allocate(Packet.MTU).order(ByteOrder.LITTLE_ENDIAN);
        buffer.put(header);
        buffer.put(body);
        int wireTypeId = body[0] & 0xFF;
        return new ReceivedPacket(wireTypeId, Packet.packetToReceive(buffer));
    }

    private byte[] readExactly(SocketChannel channel, int count) throws IOException {
        byte[] bytes = new byte[count];
        int read = 0;
        while (read < count) {
            int readNow = channel.read(ByteBuffer.wrap(bytes, read, count - read));
            if (readNow < 0) {
                return null;
            }
            read += readNow;
        }
        return bytes;
    }

    /**
     * Wait until the client has connected, then answer its ADMIN_PACKET_ADMIN_JOIN
     * the way SendProtocol() + SendWelcome() do.
     */
    public void handshake() throws IOException {
        awaitConnection(10);
        send(UpstreamAdminProtocol.protocolPacket());
        send(UpstreamAdminProtocol.welcomePacket());
    }

    public void awaitConnection(int timeoutSeconds) {
        TestUtils.await(() -> {
            synchronized (this.connectionLock) {
                return this.connection != null;
            }
        }, timeoutSeconds, "the admin client to connect");
    }

    public void awaitConnectionCount(int count, int timeoutSeconds) {
        TestUtils.await(() -> getConnectionCount() >= count, timeoutSeconds,
                "the admin client to reconnect (" + count + " connections)");
    }

    /**
     * Wait until the admin client closed its connection, which it does when it
     * receives a packet that ends the session (FULL, BANNED, SHUTDOWN, ...).
     */
    public void awaitClientDisconnect(int timeoutSeconds) {
        TestUtils.await(this::isClientDisconnected, timeoutSeconds, "the admin client to disconnect");
    }

    /**
     * Close the connection, as SendError() / a full or banned server does.
     */
    public void disconnect() {
        synchronized (this.connectionLock) {
            try {
                if (this.connection != null) {
                    this.connection.close();
                    this.connection = null;
                }
            } catch (IOException ignored) {
            }
        }
    }

    public void send(PacketWriter packet) throws IOException {
        awaitConnection(10);
        byte[] bytes = packet.toBytes();
        synchronized (this.connectionLock) {
            this.connection.write(ByteBuffer.wrap(bytes));
        }
    }

    public void sendAll(List<PacketWriter> packets) throws IOException {
        for (PacketWriter packet : packets) {
            send(packet);
        }
    }

    /**
     * @return every packet the client sent so far
     */
    public List<ReceivedPacket> received() {
        return new ArrayList<>(this.received);
    }

    /**
     * Wait for a specific command from the client.
     *
     * @param type admin command
     * @param timeoutSeconds
     * @return the packet
     */
    public ReceivedPacket awaitPacket(PacketAdminType type, int timeoutSeconds) {
        TestUtils.await(() -> {
            for (ReceivedPacket packet : this.received) {
                if (packet.getWireTypeId() == type.ordinal()) {
                    return true;
                }
            }
            return false;
        }, timeoutSeconds, "ADMIN_PACKET_" + type);
        for (ReceivedPacket packet : this.received) {
            if (packet.getWireTypeId() == type.ordinal()) {
                return packet;
            }
        }
        throw new IllegalStateException("unreachable");
    }

    /**
     * All packets of one command type, in the order they were sent.
     */
    public List<ReceivedPacket> packets(PacketAdminType type) {
        List<ReceivedPacket> matching = new ArrayList<>();
        for (ReceivedPacket packet : this.received) {
            if (packet.getWireTypeId() == type.ordinal()) {
                matching.add(packet);
            }
        }
        return matching;
    }

    @Override
    public void close() {
        this.running = false;
        try {
            synchronized (this.connectionLock) {
                if (this.connection != null) {
                    this.connection.close();
                }
            }
        } catch (IOException ignored) {
        }
        try {
            this.listener.close();
        } catch (IOException ignored) {
        }
        this.acceptor.interrupt();
    }
}
