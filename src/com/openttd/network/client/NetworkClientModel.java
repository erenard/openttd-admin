package com.openttd.network.client;

import java.io.PrintWriter;
import java.io.StringWriter;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;

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
