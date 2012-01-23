package com.openttd.network.admin;

import com.openttd.network.constant.TcpAdmin.PacketServerType;

public class NetworkEvent {

	private final PacketServerType packetServerType;
	private long clientId;
	private short companyId;
	private short extraId;
	private String message;
	private long date;

	public NetworkEvent(PacketServerType packetServerType) {
		this.packetServerType = packetServerType;
	}

	public static NetworkEvent newClientEvent(PacketServerType packetServerType, long clientId) {
		NetworkEvent networkEvent = new NetworkEvent(packetServerType);
		networkEvent.clientId = clientId;
		return networkEvent;
	}

	public static NetworkEvent newCompanyEvent(PacketServerType packetServerType, short companyId) {
		NetworkEvent networkEvent = new NetworkEvent(packetServerType);
		networkEvent.companyId = companyId;
		return networkEvent;
	}

	public static NetworkEvent newRConEvent(PacketServerType packetServerType, char color, String message) {
		NetworkEvent networkEvent = new NetworkEvent(packetServerType);
		networkEvent.message = message;
		return networkEvent;
	}

	public static NetworkEvent newDateEvent(PacketServerType packetServerType, long date) {
		NetworkEvent networkEvent = new NetworkEvent(packetServerType);
		networkEvent.date = date;
		return networkEvent;
	}

	public static NetworkEvent newChatEvent(PacketServerType packetServerType, long clientId, String message) {
		NetworkEvent networkEvent = new NetworkEvent(packetServerType);
		networkEvent.clientId = clientId;
		networkEvent.message = message;
		return networkEvent;
	}

	public long getClientId() {
		return clientId;
	}

	public void setClientId(long clientId) {
		this.clientId = clientId;
	}

	public short getCompanyId() {
		return companyId;
	}

	public void setCompanyId(short companyId) {
		this.companyId = companyId;
	}

	public short getExtraId() {
		return extraId;
	}

	public void setExtraId(short extraId) {
		this.extraId = extraId;
	}

	public String getMessage() {
		return message;
	}

	public void setMessage(String message) {
		this.message = message;
	}

	public long getDate() {
		return date;
	}

	public void setDate(long date) {
		this.date = date;
	}

	public PacketServerType getPacketServerType() {
		return packetServerType;
	}
}
