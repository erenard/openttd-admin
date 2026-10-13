package com.openttd.network.constant;

public interface PacketType {

	enum SubProtocol {
		UDP,
		TCP_ADMIN,
		TCP_GAME; // TODO Rename TCP_CLIENT
	}

	SubProtocol getType();
}
