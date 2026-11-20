package com.openttd.network.core;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.io.IOException;
import java.math.BigInteger;
import java.net.InetSocketAddress;
import java.nio.ByteBuffer;
import java.nio.channels.ServerSocketChannel;
import java.nio.channels.SocketChannel;
import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import com.openttd.network.constant.TcpAdmin.PacketAdminType;
import com.openttd.network.constant.TcpAdmin.PacketServerType;
import com.openttd.server.PacketWriter;

/**
 * The wire codec, checked against OpenTTD 15.3's src/network/core/packet.cpp:
 * little endian integers, NUL terminated strings, and a uint16 header holding
 * the total packet length including those three header bytes.
 */
public class PacketCodecTest {

    @Test
    @DisplayName("the header carries the total length, header included")
    public void testHeaderLength() {
        Packet packet = Packet.packetToSend(PacketAdminType.ADMIN_PACKET_ADMIN_PING);
        packet.writeUint32(7L);
        packet.prepareToSend();

        byte[] bytes = new byte[3 + 4];
        packet.getBuffer().get(bytes, 0, bytes.length);
        assertEquals(7, bytes[0] & 0xFF | ((bytes[1] & 0xFF) << 8));
        assertEquals(PacketAdminType.ADMIN_PACKET_ADMIN_PING.ordinal(), bytes[2]);
    }

    @Test
    @DisplayName("server packet ids are written offset by 100")
    public void testServerPacketIdOffset() {
        byte[] bytes = new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_DATE).writeUint32(1L).toBytes();
        assertEquals(107, bytes[2] & 0xFF);
    }

    @Test
    public void testUint32LittleEndian() throws IOException {
        Packet packet = decode(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_PONG).writeUint32(0xDEADBEEFL));
        assertEquals(0xDEADBEEFL, packet.readUint32());
        // The four bytes are least significant first.
        byte[] payload = new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_PONG).writeUint32(0x04030201L).getPayload();
        assertEquals(0x01, payload[0]);
        assertEquals(0x02, payload[1]);
        assertEquals(0x03, payload[2]);
        assertEquals(0x04, payload[3]);
    }

    @Test
    public void testUint16LittleEndian() throws IOException {
        Packet packet = decode(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_RCON).writeUint16(0x1234));
        assertEquals((char) 0x1234, packet.readUint16());
    }

    @Test
    public void testUint8AndBool8() throws IOException {
        Packet packet = decode(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CMD_NAMES)
                .writeUint8(255).writeBool8(true).writeBool8(false));
        assertEquals(255, packet.readUint8());
        assertTrue(packet.readBool8());
        assertEquals(false, packet.readBool8());
    }

    /**
     * Send_uint64 / Send_int64 are 8 raw little endian bytes. readUint64 must
     * scale the high half by 2^32, not by 2^32 - 1.
     */
    @Test
    @DisplayName("uint64 round trips across the 2^32 boundary")
    public void testUint64() throws IOException {
        BigInteger[] values = {
                BigInteger.ZERO,
                BigInteger.ONE,
                BigInteger.valueOf(4294967295L), // 2^32 - 1
                BigInteger.valueOf(4294967296L), // 2^32
                BigInteger.valueOf(4294967297L), // 2^32 + 1
                BigInteger.valueOf(Long.MAX_VALUE),
                new BigInteger("9223372036854775807").add(BigInteger.ONE), // 2^63
                new BigInteger("18446744073709551615"), // 2^64 - 1
        };
        for (BigInteger value : values) {
            Packet packet = decode(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_COMPANY_ECONOMY).writeUint64(value));
            assertEquals(value, packet.readUint64(), "uint64 " + value);
        }
    }

    @Test
    @DisplayName("int64 round trips, including negatives")
    public void testInt64() throws IOException {
        long[] values = {0L, 1L, -1L, 12345678901234L, -12345678901234L, Long.MAX_VALUE, Long.MIN_VALUE};
        for (long value : values) {
            Packet packet = decode(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_COMPANY_ECONOMY).writeInt64(value));
            assertEquals(value, packet.readInt64(), "int64 " + value);
        }
    }

    /**
     * Send_string sends UTF-8 bytes; writeString encodes UTF-8, so readString
     * must decode UTF-8 rather than one byte per char.
     */
    @Test
    @DisplayName("strings round trip as UTF-8")
    public void testUtf8Strings() throws IOException {
        String[] values = {"", "hello", "Grüße", "こんにちは", "café ☕", "line\nbreak"};
        for (String value : values) {
            Packet packet = decode(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CHAT).writeString(value));
            assertEquals(value, packet.readString(), "string " + value);
        }
    }

    @Test
    @DisplayName("remaining() reports the bytes left in the packet")
    public void testRemaining() throws IOException {
        byte[] variableData = {1, 2, 3, 4, 5};
        Packet packet = decode(new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_CMD_LOGGING)
                .writeUint32(2L).writeUint8(1).writeUint16(37).writeBytes(variableData).writeUint32(99L));
        packet.readUint32();
        packet.readUint8();
        packet.readUint16();
        assertEquals(variableData.length + 4, packet.remaining());
        packet.readBytes(variableData.length);
        assertEquals(4, packet.remaining());
        assertEquals(99L, packet.readUint32());
        assertEquals(0, packet.remaining());
    }

    /**
     * COMPAT_MTU in src/network/core/config.h is 1460, and Packet.MTU is the
     * buffer both sides allocate.
     */
    @Test
    public void testMtu() {
        assertEquals(1460, Packet.MTU);
    }

    /**
     * Socket.receive() reassembles packets split across TCP reads, and unpacks
     * packets concatenated into one read.
     */
    @Test
    @DisplayName("packets split across TCP reads are reassembled")
    public void testPacketFraming() throws IOException {
        try (ServerSocketChannel listener = ServerSocketChannel.open()) {
            listener.bind(new InetSocketAddress("localhost", 0));
            Socket client = Socket.newTcpSocket("localhost", listener.socket().getLocalPort());
            SocketChannel serverSide = listener.accept();

            PacketWriter datePacket = new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_DATE).writeUint32(42L);
            PacketWriter pongPacket = new PacketWriter(PacketServerType.ADMIN_PACKET_SERVER_PONG).writeUint32(7L);
            byte[] first = datePacket.toBytes();
            byte[] second = pongPacket.toBytes();

            // Half of the first packet only: receive() must not report it yet.
            serverSide.write(ByteBuffer.wrap(first, 0, first.length / 2));
            List<Packet> partial = client.receive();
            assertEquals(0, partial.size());

            // The rest of the first packet, plus the whole second one.
            serverSide.write(ByteBuffer.wrap(first, first.length / 2, first.length - first.length / 2));
            serverSide.write(ByteBuffer.wrap(second));
            List<Packet> packets = client.receive();
            assertEquals(2, packets.size());

            assertEquals(PacketServerType.ADMIN_PACKET_SERVER_DATE.ordinal() + 100, packets.get(0).readUint8());
            assertEquals(42L, packets.get(0).readUint32());
            assertEquals(PacketServerType.ADMIN_PACKET_SERVER_PONG.ordinal() + 100, packets.get(1).readUint8());
            assertEquals(7L, packets.get(1).readUint32());

            client.close();
            serverSide.close();
        }
    }

    private static Packet decode(PacketWriter writer) {
        ByteBuffer buffer = ByteBuffer.wrap(writer.toBytes());
        Packet packet = Packet.packetToReceive(buffer);
        packet.prepareToRead();
        packet.readUint8(); // The packet type
        return packet;
    }
}
