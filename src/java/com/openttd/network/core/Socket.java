package com.openttd.network.core;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.channels.ByteChannel;
import java.nio.channels.DatagramChannel;
import java.nio.channels.SocketChannel;
import java.util.ArrayList;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.openttd.network.constant.PacketType;

public class Socket {

    private static final Logger log = LoggerFactory.getLogger(Socket.class);

    private final ByteChannel channel;

    private Socket(ByteChannel channel) {
        this.channel = channel;
    }

    public static Socket newTcpSocket(String host, Integer port) throws IOException {
        InetSocketAddress adminAddress = new InetSocketAddress(host, port);
        SocketChannel tcp = SocketChannel.open(adminAddress);
        return new Socket(tcp);
    }

    public static Socket newUdpSocket(String host, Integer port) throws IOException {
        DatagramChannel udp = DatagramChannel.open();
        InetSocketAddress datagramAddress = new InetSocketAddress(host, port);
        udp.connect(datagramAddress);
        return new Socket(udp);
    }

    public void sendEmptyPacket(PacketType packet) {
        Packet toSend = Packet.packetToSend(packet);
        send(toSend);
    }

    public void send(Packet packet) {
        try {
            packet.prepareToSend();
            channel.write(packet.getBuffer());
            if (log.isTraceEnabled()) {
                log.trace("Send " + packet.hashCode());
            }
        } catch (IOException e) {
            log.error(e.getMessage());
        }
    }

    ByteBuffer data = null;

    /**
     * Packet-ize the network byte stream
     *
     * @return packets
     * @throws IOException
     */
    public List<Packet> receive() throws IOException {
        List<Packet> packets = new ArrayList<>();
        if (data == null) {
            data = ByteBuffer.allocate(Packet.MTU);
            data.order(ByteOrder.LITTLE_ENDIAN);
        }
        data = receiveMore(channel, data, packets);
        return packets;
    }

    private ByteBuffer receiveMore(ByteChannel channel, ByteBuffer data, List<Packet> packets) throws IOException {
        channel.read(data);
        int dataLength = data.position();
        if (dataLength <= 0) {
            throw new IOException("Connection lost, received data length: " + dataLength);
        }

        Packet dataReader = Packet.packetToReceive(data);
        data.rewind();

        // Read packet size (16bit)
        int packetLength = dataReader.readUint16();
        while (dataLength >= packetLength && dataLength > 0) {
            // packet's buffer
            ByteBuffer buffer = ByteBuffer.allocate(Packet.MTU);
            buffer.order(ByteOrder.LITTLE_ENDIAN);
            // Rewind the packet size (2*8bit)
            data.position(data.position() - 2);
            // Copy the data to the buffer's packet
            for (int i = 0; i < packetLength; i++) {
                buffer.put(data.get());
            }
            // Instantiate the Packet
            Packet packet = Packet.packetToReceive(buffer);
            packet.prepareToRead();
            packets.add(packet);
            // Decrease the data length of the readed packet length
            dataLength -= packetLength;
            // if their is still data
            if (dataLength > 0) {
                // Read packet size (16bit) for the next loop
                packetLength = dataReader.readUint16();
            } else {
                // Useful for the coming if (the packet is finished)
                packetLength = 0;
            }
        }
        // if the packet is unfinished
        if (packetLength > dataLength) {
            // remainingData contains the beggining of the unfinished packet.
            ByteBuffer buffer = ByteBuffer.allocate(Packet.MTU);
            buffer.order(ByteOrder.LITTLE_ENDIAN);
            // Rewind the packet size (2*8bit)
            data.position(data.position() - 2);
            // Copy the data to the buffer's packet
            for (int i = 0; i < dataLength; i++) {
                buffer.put(data.get());
            }
            return buffer;
        }
        return null;
    }

    public boolean isOpen() {
        return channel != null && channel.isOpen();
    }

    public void close() {
        try {
            if (channel != null) {
                channel.close();
            }
        } catch (IOException e) {
        }
    }
}
