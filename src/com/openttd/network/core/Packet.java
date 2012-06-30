package com.openttd.network.core;

import java.math.BigInteger;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.openttd.network.constant.PacketType;
import com.openttd.network.constant.PacketUdpType;
import com.openttd.network.constant.TcpAdmin.PacketAdminType;

public class Packet {
	private static final Logger log = LoggerFactory.getLogger(Packet.class);
	public static final int MTU = 1460;
	private ByteBuffer buffer;
	// Meta information
	private int packetTypeId;
	private int subProtocol;

	private Packet() {
	}

	public static Packet packetToSend(PacketType packetType) {
		int type = packetType.getType();
		Packet packet = null;
		if (packetType.getType() == PacketType.UDP) {
			PacketUdpType packetUdp = (PacketUdpType) packetType;
			packet = packetToSend(packetUdp.ordinal());
			if (log.isTraceEnabled()) log.trace(packet.hashCode() + " - UDP " + packetUdp);
		} else {
			PacketAdminType packetTcp = (PacketAdminType) packetType;
			packet = packetToSend(packetTcp.ordinal());
			if (log.isTraceEnabled()) log.trace(packet.hashCode() + " - TCP " + packetTcp);
		}
		packet.subProtocol = type;
		return packet;
	}

	private static Packet packetToSend(int packetId) {
		Packet packet = new Packet();
		packet.buffer = ByteBuffer.allocate(MTU);
		packet.buffer.order(ByteOrder.LITTLE_ENDIAN);
		packet.buffer.put((byte) 0);
		packet.buffer.put((byte) 0);
		packet.buffer.put((byte) packetId);
		packet.packetTypeId = packetId;
		return packet;
	}

	public static Packet packetToReceive(ByteBuffer buffer) {
		Packet packet = new Packet();
		packet.buffer = buffer;
		return packet;
	}

	public int getSubProtocol() {
		return subProtocol;
	}

	public int getPacketTypeId() {
		return packetTypeId;
	}

	public ByteBuffer getBuffer() {
		return buffer;
	}

	public void prepareToSend() {
		int length = buffer.position();
		// Taille (2o)
		buffer.put(0, (byte) ((length << 24) >> 24));
		buffer.put(1, (byte) ((length << 16) >> 24));
		buffer.limit(length);
		buffer.rewind();
	}

	public int prepareToRead() {
		buffer.rewind();
		int length = (int) readUint16();
		// buffer.limit(length);
		return length;
	}

	public byte[] readBytes(int count) {
		byte[] returned = new byte[count];
		for (int i = 0; i < count; i++) {
			returned[i] = (byte) ((buffer.get() << 24) >> 24);
		}
		return returned;
	}

	public long readInt64() {
		byte buf[] = readBytes(8);
		long byte1 = 0x000000FF & ((long) buf[7]);
		long byte2 = 0x000000FF & ((long) buf[6]);
		long byte3 = 0x000000FF & ((long) buf[5]);
		long byte4 = 0x000000FF & ((long) buf[4]);
		long byte5 = 0x000000FF & ((long) buf[3]);
		long byte6 = 0x000000FF & ((long) buf[2]);
		long byte7 = 0x000000FF & ((long) buf[1]);
		long byte8 = 0x000000FF & ((long) buf[0]);
		return ((long) (byte1 << 56 | byte2 << 48 | byte3 << 40 | byte4 << 32 | byte5 << 24 | byte6 << 16 | byte7 << 8 | byte8));
	}

	public BigInteger readUint64() {
		byte buf[] = readBytes(8);
		int byte1 = (0x000000FF & ((int) buf[7]));
		int byte2 = (0x000000FF & ((int) buf[6]));
		int byte3 = (0x000000FF & ((int) buf[5]));
		int byte4 = (0x000000FF & ((int) buf[4]));
		long firstInt32 = ((long) (byte1 << 24 | byte2 << 16 | byte3 << 8 | byte4)) & 0xFFFFFFFFL;
		int byte5 = (0x000000FF & ((int) buf[3]));
		int byte6 = (0x000000FF & ((int) buf[2]));
		int byte7 = (0x000000FF & ((int) buf[1]));
		int byte8 = (0x000000FF & ((int) buf[0]));
		long secondInt32 = ((long) (byte5 << 24 | byte6 << 16 | byte7 << 8 | byte8)) & 0xFFFFFFFFL;

		BigInteger i1 = new BigInteger("" + firstInt32);
		BigInteger i2 = new BigInteger("" + secondInt32);
		// 2^32 - 1 = 4294967295
		return i1.multiply(new BigInteger("4294967295")).add(i2);
	}

	public long readUint32() {
		byte buf[] = readBytes(4);
		int firstByte = (0x000000FF & ((int) buf[3]));
		int secondByte = (0x000000FF & ((int) buf[2]));
		int thirdByte = (0x000000FF & ((int) buf[1]));
		int fourthByte = (0x000000FF & ((int) buf[0]));
		return ((long) (firstByte << 24 | secondByte << 16 | thirdByte << 8 | fourthByte)) & 0xFFFFFFFFL;
	}

	public char readUint16() {
		byte buf[] = readBytes(2);
		int firstByte = (0x000000FF & ((int) buf[1]));
		int secondByte = (0x000000FF & ((int) buf[0]));
		return (char) (firstByte << 8 | secondByte);
	}

	public short readUint8() {
		byte buf[] = readBytes(1);
		int firstByte = (0x000000FF & ((int) buf[0]));
		return (short) firstByte;
	}

	public String readString() {
		StringBuffer newString = new StringBuffer();
		byte c = buffer.get();
		while (c != 0) {
			newString.append((char) c);
			c = buffer.get();
		}
		return newString.toString();
	}

	public boolean readBool8() {
		return readUint8() == (short) 1;
	}

	public void writeUint32(long v) {
		buffer.putInt((int) v);
	}

	public void writeUint16(char v) {
		buffer.putShort((short) v);
	}

	public void writeUint8(short v) {
		buffer.put((byte) v);
	}

	public void writeString(String s) {
		for (int i = 0; i < s.length(); i++) {
			buffer.put((byte) s.charAt(i));
		}
		buffer.put((byte) '\0');
	}

	public void writeBool8(boolean b) {
		if (b)
			buffer.put((byte) 1);
		else
			buffer.put((byte) 0);
	}
}
