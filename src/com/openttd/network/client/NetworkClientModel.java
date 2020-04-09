package com.openttd.network.client;

/**
 * Network client's state
 */
public class NetworkClientModel {
	public static final short DAY_TICKS = 74;
	public long frame_counter_server;
	public short token;
	public long last_ack_frame;
	public String passwordServerId;
	public long passwordGameSeed;
	public long clientId;
}
