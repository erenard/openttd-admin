package com.openttd.network.constant;

/**
 * See udp.h for changes
 */
public enum PacketUdpType implements PacketType {
	PACKET_UDP_CLIENT_FIND_SERVER,   ///< Queries a game server for game information
	PACKET_UDP_SERVER_RESPONSE,      ///< Reply of the game server with game information
	PACKET_UDP_END;                  ///< Must ALWAYS be on the end of this list!! (period)

	public static PacketUdpType valueOf(int id) {
		for (PacketUdpType value : values()) {
			if (value.ordinal() == id) return value;
		}
		return null;
	}

	@Override
	public SubProtocol getType() {
		return SubProtocol.UDP;
	}
}
