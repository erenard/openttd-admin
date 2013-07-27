package com.openttd.network.client;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.openttd.network.constant.NetworkInternal.NetworkLanguage;
import com.openttd.network.constant.NetworkType.DestType;
import com.openttd.network.constant.NetworkType.NetworkAction;
import com.openttd.network.constant.TcpGame.PacketGameType;
import com.openttd.network.core.Configuration;
import com.openttd.network.core.NetworkSender;
import com.openttd.network.core.Packet;
import com.openttd.network.core.Socket;

/**
 * See network_client.cpp for updates, search "::Send"
 */
public class NetworkClientSender extends NetworkSender {
	private static final Logger log = LoggerFactory.getLogger(NetworkClientSender.class);

	public NetworkClientSender(Socket socket) {
		super(socket);
	}

	/** Query the server for company information. */
	public void companyInformationQuery() {
		queue.offer(Packet.packetToSend(PacketGameType.PACKET_CLIENT_COMPANY_INFO));
	}

	/** Tell the server we would like to join. */
	void join(Configuration configuration) {
		Packet toSend = Packet.packetToSend(PacketGameType.PACKET_CLIENT_JOIN);
		toSend.writeString(configuration.openttdVersion);
		toSend.writeUint32(configuration.openttdNewgrfVersion);
		toSend.writeString(configuration.name);
		toSend.writeUint8((short) 255); // PlayAs
		toSend.writeUint8((short) NetworkLanguage.NETLANG_ANY.ordinal()); // Language
		queue.offer(toSend);
	}

	/** Tell the server we got all the NewGRFs. */
	public void newGRFsOk() {
		queue.offer(Packet.packetToSend(PacketGameType.PACKET_CLIENT_NEWGRFS_CHECKED));
	}

	public void gamePassword(Configuration configuration) {
		Packet toSend = Packet.packetToSend(PacketGameType.PACKET_CLIENT_GAME_PASSWORD);
		toSend.writeString(configuration.password);
		queue.offer(toSend);
	}

	public void companyPassword(String password) {
		Packet toSend = Packet.packetToSend(PacketGameType.PACKET_CLIENT_COMPANY_PASSWORD);
		toSend.writeString(password);
		queue.offer(toSend);
	}

	public void getMap() {
		queue.offer(Packet.packetToSend(PacketGameType.PACKET_CLIENT_GETMAP));
	}

	public void mapOk() {
		queue.offer(Packet.packetToSend(PacketGameType.PACKET_CLIENT_MAP_OK));
	}

	public void ack(long frame, short token) {
		Packet toSend = Packet.packetToSend(PacketGameType.PACKET_CLIENT_ACK);
		toSend.writeUint32(frame);
		toSend.writeUint8(token);
		queue.offer(toSend);
	}

	/**
	 * Send chat as the server:
	 * uint8 Action such as NETWORK_ACTION_CHAT_CLIENT (see #NetworkAction).
	 * uint8 Destination type such as DESTTYPE_BROADCAST (see #DestType).
	 * uint32 ID of the destination such as company or client id.
	 * string Message.
	 */
	public void chat(NetworkAction action, DestType type, long dest, String message) {
		Packet packet = Packet.packetToSend(PacketGameType.PACKET_CLIENT_CHAT);
		packet.writeUint8((short) action.ordinal());
		packet.writeUint8((short) type.ordinal());
		packet.writeUint32(dest);
		packet.writeString(message);
		//packet.writeUint64(data);
		queue.offer(packet);
	}

	void quit() {
		queue.offer(Packet.packetToSend(PacketGameType.PACKET_CLIENT_QUIT));
	}

	/**
	 * Execute a command on the servers console:
	 * string Command to be executed.
	 */
	public void rcon(String command) {
		Packet packet = Packet.packetToSend(PacketGameType.PACKET_CLIENT_RCON);
		packet.writeString(command);
		queue.offer(packet);
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

	@Override
	protected boolean isQuitPacket(Packet packet) {
		return packet.getPacketTypeId() == PacketGameType.PACKET_CLIENT_QUIT.ordinal();
	}

	@Override
	protected void debugPacket(Packet packet) {
		log.debug("Snd " + PacketGameType.valueOf(packet.getPacketTypeId()));
	}
}
