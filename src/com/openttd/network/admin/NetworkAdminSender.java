package com.openttd.network.admin;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.google.gson.Gson;
import com.google.gson.JsonArray;
import com.google.gson.JsonObject;
import com.google.gson.JsonPrimitive;
import com.openttd.constant.OTTD;
import com.openttd.network.constant.GameScript;
import com.openttd.network.constant.GameScript.GSCommand;
import com.openttd.network.constant.NetworkType.DestType;
import com.openttd.network.constant.NetworkType.NetworkAction;
import com.openttd.network.constant.TcpAdmin.AdminUpdateFrequency;
import com.openttd.network.constant.TcpAdmin.AdminUpdateType;
import com.openttd.network.constant.TcpAdmin.PacketAdminType;
import com.openttd.network.core.Configuration;
import com.openttd.network.core.NetworkSender;
import com.openttd.network.core.Packet;
import com.openttd.network.core.Socket;

/**
 * See network_admin.cpp for updates, search "Receive_ADMIN_"
 */
public class NetworkAdminSender extends NetworkSender {
	private static final Logger log = LoggerFactory.getLogger(NetworkAdminSender.class);

	private Protocol protocol;
	
	public NetworkAdminSender(Socket socket) {
		super(socket);
	}
	
	public void setProtocol(Protocol protocol) {
		this.protocol = protocol;
	}
	
	/**
	 * Join the admin network:
	 * string Password the server is expecting for this network.
	 * string Name of the application being used to connect.
	 * string Version string of the application being used to connect.
	 */
	void join(Configuration configuration) {
		Packet toSend = Packet.packetToSend(PacketAdminType.ADMIN_PACKET_ADMIN_JOIN);
		toSend.writeString(configuration.password);
		toSend.writeString(configuration.name);
		toSend.writeString(configuration.openttdVersion);
		queue.offer(toSend);
	}

	/**
	 * Notification to the server that this admin is quitting.
	 */
	void quit() {
		queue.offer(Packet.packetToSend(PacketAdminType.ADMIN_PACKET_ADMIN_QUIT));
	}

	/**
	 * Register updates to be sent at certain frequencies (as announced in the PROTOCOL packet):
	 * uint16 Update type (see #AdminUpdateType).
	 * uint16 Update frequency (see #AdminUpdateFrequency), setting #ADMIN_FREQUENCY_POLL is always ignored.
	 */
	public void updateFrequency(AdminUpdateType adminUpdateType, AdminUpdateFrequency adminUpdateFrequency) {
		if (!protocol.hasProtocol(adminUpdateType, adminUpdateFrequency)) {
			log.error("The server does not support " + adminUpdateFrequency + " for " + adminUpdateType);
			log.error(AdminUpdateFrequency.ADMIN_FREQUENCY_POLL + " used instead.");
			adminUpdateFrequency = AdminUpdateFrequency.ADMIN_FREQUENCY_POLL;
		}

		Packet packet = Packet.packetToSend(PacketAdminType.ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY);
		packet.writeUint16((char) adminUpdateType.ordinal());
		packet.writeUint16((char) adminUpdateFrequency.mask);
		queue.offer(packet);
	}

	/**
	 * Poll the server for certain updates, an invalid poll (e.g. not existent id) gets silently dropped:
	 * uint8 #AdminUpdateType the server should answer for, only if #AdminUpdateFrequency #ADMIN_FREQUENCY_POLL is advertised in the
	 * PROTOCOL packet.
	 * uint32 ID relevant to the packet type, e.g.
	 * - the client ID for #ADMIN_UPDATE_CLIENT_INFO. Use UINT32_MAX to show all clients.
	 * - the company ID for #ADMIN_UPDATE_COMPANY_INFO. Use UINT32_MAX to show all companies.
	 */
	public void poll(AdminUpdateType adminUpdateType, long data) {
		if (!protocol.hasProtocol(adminUpdateType, AdminUpdateFrequency.ADMIN_FREQUENCY_POLL)) {
			log.error("The server does not support " + PacketAdminType.ADMIN_PACKET_ADMIN_POLL + " for " + adminUpdateType);
			return;
		}

		Packet packet = Packet.packetToSend(PacketAdminType.ADMIN_PACKET_ADMIN_POLL);
		packet.writeUint8((short) adminUpdateType.ordinal());
		packet.writeUint32(data);
		queue.offer(packet);
	}

	/**
	 * Send chat as the server:
	 * uint8 Action such as NETWORK_ACTION_CHAT_CLIENT (see #NetworkAction).
	 * uint8 Destination type such as DESTTYPE_BROADCAST (see #DestType).
	 * uint32 ID of the destination such as company or client id.
	 * string Message.
	 */
	public void chat(NetworkAction action, DestType type, long dest, String message) {
		Packet packet = Packet.packetToSend(PacketAdminType.ADMIN_PACKET_ADMIN_CHAT);
		packet.writeUint8((short) action.ordinal());
		packet.writeUint8((short) type.ordinal());
		packet.writeUint32(dest);
		packet.writeString(message);
		queue.offer(packet);
	}

	/**
	 * Execute a command on the servers console:
	 * string Command to be executed.
	 */
	public void rcon(String command) {
		Packet packet = Packet.packetToSend(PacketAdminType.ADMIN_PACKET_ADMIN_RCON);
		packet.writeString(command);
		queue.offer(packet);
	}

	/* Utility methods */
	public void pollDate() {
		poll(AdminUpdateType.ADMIN_UPDATE_DATE, 0l);
	}

	public void pollCmdNames() {
		poll(AdminUpdateType.ADMIN_UPDATE_CMD_NAMES, 0l);
	}

	public void pollClientInfo(long clientId) {
		if (clientId < 0) clientId = Long.MAX_VALUE;
		poll(AdminUpdateType.ADMIN_UPDATE_CLIENT_INFO, clientId);
	}

	public void pollCompanyInfo(short companyId) {
		long longId = (long) companyId;
		if (longId < 0) longId = Long.MAX_VALUE;
		poll(AdminUpdateType.ADMIN_UPDATE_COMPANY_INFO, longId);
	}

	public void pollCompanyEconomy(short companyId) {
		long longId = (long) companyId;
		if (longId < 0) longId = Long.MAX_VALUE;
		poll(AdminUpdateType.ADMIN_UPDATE_COMPANY_ECONOMY, longId);
	}

	public void pollCompanyStats(short companyId) {
		long longId = (long) companyId;
		if (longId < 0) longId = Long.MAX_VALUE;
		poll(AdminUpdateType.ADMIN_UPDATE_COMPANY_STATS, longId);
	}

	public void chatClient(long clientId, String message) {
		if (clientId < 0) {
			this.chatBroadcast(message);
		} else {
			this.chat(NetworkAction.NETWORK_ACTION_SERVER_MESSAGE, DestType.DESTTYPE_CLIENT, clientId, message);
		}
	}

	/**
	 * Chat to a team
	 *
	 * @param companyId
	 * @param message
	 */
	public void chatCompany(short companyId, String message) {
		this.chat(NetworkAction.NETWORK_ACTION_SERVER_MESSAGE, DestType.DESTTYPE_TEAM, companyId, message);
	}

	/**
	 * Chat to all
	 * @param message
	 */
	public void chatBroadcast(String message) {
		this.chat(NetworkAction.NETWORK_ACTION_SERVER_MESSAGE, DestType.DESTTYPE_BROADCAST, 0, message);
	}

	/**
	 * Send a gamescript
	 * @return false if packet was too long
	 */
	public boolean gameScript(String json) {
		if(json == null || json.length() > Packet.MTU - 3) {
			log.error(json + " is " + json.length() + " long, max: " + (Packet.MTU - 3));
			return false;
		}
		Packet packet = Packet.packetToSend(PacketAdminType.ADMIN_PACKET_ADMIN_GAMESCRIPT);
		packet.writeString(json);
		queue.offer(packet);
		return true;
	}

	/**
	 * Ping the server
	 * @param pingId
	 */
	public void ping(long pingId) {
		Packet packet = Packet.packetToSend(PacketAdminType.ADMIN_PACKET_ADMIN_PING);
		packet.writeUint32(pingId);
		queue.offer(packet);
	}

	/**
	 * Add a goal to a company.
	 * @param companyId
	 * @param text Text displayed for this goal
	 * @param goalType Type of the goal's destination
	 * @param destinationId Id of the goal's destination
	 */
	public void addCompanyGoal(short companyId, String text, OTTD.GoalType goalType, int destinationId) {
		JsonArray arguments = new JsonArray();
		arguments.add(new JsonPrimitive(companyId));
		arguments.add(new JsonPrimitive(text));
		arguments.add(new JsonPrimitive(goalType.ordinal()));
		arguments.add(new JsonPrimitive(destinationId));
		JsonObject json = new JsonObject();
		json.add(GameScript.CMD, new JsonPrimitive(GSCommand.addGoal.ordinal()));
		json.add(GameScript.ARGS, arguments);
		Gson gson = new Gson();
		gameScript(gson.toJson(json));
	}

	/**
	 * Add a global goal to the game.
	 * @param text Text displayed for this goal
	 * @param goalType Type of the goal's destination
	 * @param destinationId Id of the goal's destination
	 */
	public void addGlobalGoal(String text, OTTD.GoalType goalType, int destinationId) {
		addCompanyGoal((short) -1, text, goalType, destinationId);
	}

	/**
	 * Remove a goal
	 * @param goalId
	 */
	public void removeGoal(int goalId) {
		JsonArray arguments = new JsonArray();
		arguments.add(new JsonPrimitive(goalId));
		JsonObject json = new JsonObject();
		json.add(GameScript.CMD, new JsonPrimitive(GSCommand.removeGoal.ordinal()));
		json.add(GameScript.ARGS, arguments);
		Gson gson = new Gson();
		gameScript(gson.toJson(json));
	}

	/**
	 * Clear the goal list.
	 */
	public void removeAllGoal() {
		JsonObject json = new JsonObject();
		json.add(GameScript.CMD, new JsonPrimitive(GSCommand.removeAllGoal.ordinal()));
		json.add(GameScript.ARGS, new JsonPrimitive(0));
		Gson gson = new Gson();
		gameScript(gson.toJson(json));
	}

	/**
	 * Set a cargo goal for a town.
	 */
	public void setTownCargoGoal(int townId, OTTD.TownEffect townEffect, int goal) {
		JsonArray arguments = new JsonArray();
		arguments.add(new JsonPrimitive(townId));
		arguments.add(new JsonPrimitive(townEffect.ordinal()));
		arguments.add(new JsonPrimitive(goal));
		JsonObject json = new JsonObject();
		json.add(GameScript.CMD, new JsonPrimitive(GSCommand.setTownCargoGoal.ordinal()));
		json.add(GameScript.ARGS, arguments);
		Gson gson = new Gson();
		gameScript(gson.toJson(json));
	}
	@Override
	protected boolean isQuitPacket(Packet packet) {
		return packet.getPacketTypeId() == PacketAdminType.ADMIN_PACKET_ADMIN_QUIT.ordinal();
	}

	@Override
	protected void debugPacket(Packet packet) {
		log.debug("Snd " + PacketAdminType.valueOf(packet.getPacketTypeId()));
	}
}
