package com.openttd.network.constant;

public interface PacketType {
	public enum SubProtocol {
		UDP, TCP_ADMIN, TCP_GAME;
	}

	SubProtocol getType();
}
