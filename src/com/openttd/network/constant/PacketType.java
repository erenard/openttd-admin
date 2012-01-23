package com.openttd.network.constant;

public interface PacketType {
	static final int TCP = 0;
	static final int UDP = 1;

	int getType();
}
