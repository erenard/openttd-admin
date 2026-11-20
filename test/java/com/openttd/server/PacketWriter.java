package com.openttd.server;

import java.io.ByteArrayOutputStream;
import java.math.BigInteger;
import java.nio.charset.StandardCharsets;

import com.openttd.network.constant.PacketType;
import com.openttd.network.constant.TcpAdmin.PacketAdminType;
import com.openttd.network.constant.TcpAdmin.PacketServerType;

/**
 * Builds admin packets the way the C++ server does, i.e. exactly as
 * Packet::Send_uint8 / Send_uint16 / Send_uint32 / Send_int64 / Send_uint64 /
 * Send_string / Send_bool do in OpenTTD's src/network/core/packet.cpp: little
 * endian, strings terminated by a NUL byte, and a header holding the total
 * length of the packet (header included).
 *
 * Used by FakeAdminServer to script server->admin packets, and by the sender
 * tests to compare the bytes NetworkAdminSender produces.
 */
public class PacketWriter {

    private final PacketType type;
    private final ByteArrayOutputStream payload = new ByteArrayOutputStream();

    public PacketWriter(PacketType type) {
        this.type = type;
    }

    /**
     * Wire packet id: admin packets use their ordinal, server packets are
     * offset by 100 (PacketServerTypeStartValue).
     *
     * @return the id written in the packet header
     */
    public int getWireTypeId() {
        // PacketType has no ordinal(): the wire id comes from the enum that
        // implements it, exactly as Packet.packetToSend() does.
        if (type instanceof PacketServerType) {
            return ((PacketServerType) type).ordinal() + 100;
        }
        if (type instanceof PacketAdminType) {
            return ((PacketAdminType) type).ordinal();
        }
        throw new IllegalArgumentException("Not an admin packet: " + type);
    }

    public PacketWriter writeUint8(int value) {
        payload.write(value & 0xFF);
        return this;
    }

    public PacketWriter writeBool8(boolean value) {
        payload.write(value ? 1 : 0);
        return this;
    }

    public PacketWriter writeUint16(int value) {
        payload.write(value & 0xFF);
        payload.write((value >> 8) & 0xFF);
        return this;
    }

    public PacketWriter writeUint32(long value) {
        for (int shift = 0; shift < 32; shift += 8) {
            payload.write((int) ((value >> shift) & 0xFF));
        }
        return this;
    }

    public PacketWriter writeInt64(long value) {
        for (int shift = 0; shift < 64; shift += 8) {
            payload.write((int) ((value >> shift) & 0xFF));
        }
        return this;
    }

    public PacketWriter writeUint64(BigInteger value) {
        // toByteArray() is big endian and may carry a leading sign byte: always
        // write exactly 8 little endian bytes.
        byte[] big = value.toByteArray();
        byte[] little = new byte[8];
        for (int i = big.length - 1, j = 0; j < little.length && i >= 0; i--, j++) {
            little[j] = (byte) (big[i] & 0xFF);
        }
        payload.writeBytes(little);
        return this;
    }

    public PacketWriter writeString(String value) {
        payload.writeBytes(value.getBytes(StandardCharsets.UTF_8));
        payload.write(0);
        return this;
    }

    public PacketWriter writeBytes(byte[] value) {
        payload.writeBytes(value);
        return this;
    }

    public byte[] getPayload() {
        return payload.toByteArray();
    }

    /**
     * @return the whole packet, header included
     */
    public byte[] toBytes() {
        byte[] body = getPayload();
        int length = 3 + body.length;
        byte[] packet = new byte[length];
        packet[0] = (byte) (length & 0xFF);
        packet[1] = (byte) ((length >> 8) & 0xFF);
        packet[2] = (byte) getWireTypeId();
        System.arraycopy(body, 0, packet, 3, body.length);
        return packet;
    }

    @Override
    public String toString() {
        return type + " (" + getWireTypeId() + ", " + payload.size() + " bytes of payload)";
    }
}
