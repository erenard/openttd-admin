package com.openttd.network.client;

import java.io.IOException;
import java.math.BigInteger;
import java.net.ConnectException;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.openttd.network.admin.NetworkException;
import com.openttd.network.constant.TcpGame.PacketGameType;
import com.openttd.network.core.Configuration;
import com.openttd.network.core.Packet;
import com.openttd.network.core.Socket;

/**
 * Network listening thread, own a second speaking thread.
 * See network_client.cpp for updates
 */
public class NetworkClient extends Thread {

	private static final Logger log = LoggerFactory.getLogger(NetworkClient.class);

	private final Configuration configuration;

	public NetworkClient(Configuration configuration) {
		this.configuration = configuration;
	}

	// Network speak thread
	private NetworkClientSender send;

	public NetworkClientSender getSend() {
		return send;
	}

	/**
	 * Thread's state
	 */
	private boolean running;

	/**
	 * Stop this thread
	 */
	public void shutdown() {
		// Stop the client
		this.running = false;
		// Interrupt the networking threads
		send.quit();
		try {
			this.join(5000);
		} catch (InterruptedException e) {
			log.error(e.getMessage(), e);
		}
		log.debug("stopped");
	}

	/**
	 * Start this thread
	 */
	@Override
	public void run() {
		log.debug("started");
		this.running = true;
		// Auto-Reconnection
		while (this.running) {
			Socket socket = null;
			try {
				log.debug("connecting...");
				socket = Socket.newTcpSocket(configuration.host, configuration.clientPort);
				send = new NetworkClientSender(socket);
				send.startup();
				send.join(configuration);
				// Packets listening
				while (this.running) {
					List<Packet> packets = socket.receive();
					for (Packet packet : packets) {
						handlePacket(packet);
					}
				}
			} catch (ConnectException e) {
				log.error(e.getMessage(), e);
			} catch (IOException e) {
				// Do not output error in case of shutdown
				if (this.running) {
					log.error(e.getMessage(), e);
				}
			} catch (NetworkException e) {
				log.error(e.getMessage(), e);
			} catch (Exception e) {
				log.error(e.getMessage(), e);
			} finally {
				log.debug("disconneting...");
				if (send != null) send.shutdown();
				if (socket != null) socket.close();
			}
			try {
				Thread.sleep(5000);
			} catch (InterruptedException e) {
				log.error(e.getMessage(), e);
			}
		}
	}

	private NetworkClientModel client = new NetworkClientModel();
	
	/**
	 * Extract infos from the server's packets,
	 *
	 * See network_client.cpp for updates, search "Receive_SERVER_"
	 * @param packet
	 * @throws NetworkException
	 */
	@SuppressWarnings("unused")
	private void handlePacket(Packet packet) throws NetworkException {
		int packetTypeId = packet.readUint8();
		PacketGameType packetType = PacketGameType.valueOf(packetTypeId);
		if(log.isDebugEnabled()) {
			log.debug("Rcv " + packetType);
		}
		switch (packetType) {
		case PACKET_SERVER_FULL:
			break;
		case PACKET_SERVER_BANNED:
			break;
		case PACKET_SERVER_COMPANY_INFO: {
			short companyInfoVersion = packet.readUint8();
			if(packet.readBool8()) {
				short companyId = packet.readUint8();
				String name = packet.readString();
				long inauguratedYear = packet.readUint32();
				BigInteger value = packet.readUint64();
				BigInteger money = packet.readUint64();
				BigInteger income = packet.readUint64();
				int performance = packet.readUint16();
				boolean hasPassword = packet.readBool8();
				int[] vehicules = new int[5];
				vehicules[0] = packet.readUint16();
				vehicules[1] = packet.readUint16();
				vehicules[2] = packet.readUint16();
				vehicules[3] = packet.readUint16();
				vehicules[4] = packet.readUint16();
				int[] stations = new int[5];
				stations[0] = packet.readUint16();
				stations[1] = packet.readUint16();
				stations[2] = packet.readUint16();
				stations[3] = packet.readUint16();
				stations[4] = packet.readUint16();
				boolean isAi = packet.readBool8();
				String clients = packet.readString();
				break;
			}
		}
		case PACKET_SERVER_CLIENT_INFO: {
			long clientId = packet.readUint32();
			short companyId = packet.readUint8();
			String name = packet.readString();
			break;
		}
		case PACKET_SERVER_ERROR: {
			short errorId = packet.readUint8();
			break;
		}
		case PACKET_SERVER_CHECK_NEWGRFS: {
			short grfCount = packet.readUint8();
			for(int grfIndex = 0; grfIndex < grfCount; grfIndex++) {
				long grfId = packet.readUint32();
				//TODO incomplete
				long md5 = packet.readUint32();
			}
			send.newGRFsOk();
			break;
		}
		case PACKET_SERVER_NEED_GAME_PASSWORD: {
			send.gamePassword(configuration);
			break;
		}
		case PACKET_SERVER_NEED_COMPANY_PASSWORD: {
			long seed = packet.readUint32();
			String serverId = packet.readString();
			send.companyPassword("");
			break;
		}
		case PACKET_SERVER_WELCOME: {
			long clientId = packet.readUint32();
			long passwordSeed = packet.readUint32();
			String passwordServerId = packet.readString();
			send.getMap();
			break;
		}
		case PACKET_SERVER_WAIT: {
			short waiting = packet.readUint8();
			break;
		}
		case PACKET_SERVER_MAP_BEGIN: {
			long frame = packet.readUint32();
			break;
		}
		case PACKET_SERVER_MAP_SIZE: {
			long mapSize = packet.readUint32();
			break;
		}
		case PACKET_SERVER_MAP_DATA: {
			//TODO incomplete
			break;
		}
		case PACKET_SERVER_MAP_DONE: {
			send.mapOk();
			break;
		}
		case PACKET_SERVER_FRAME: {
			client.frame_counter_server = packet.readUint32();
			long frame_counter_max = packet.readUint32();
			if (packet.getBuffer().position() + 1 < packet.getLength()) {
				long sync_frame = client.frame_counter_server;
				long sync_seed1 = packet.readUint32();
				if (packet.getBuffer().position() + 1 < packet.getLength()) {
					long sync_seed2 = packet.readUint32();
				}
			}
			if(packet.getBuffer().position() != packet.getLength()) {
				client.token = packet.readUint8();
			}
			if(client.last_ack_frame < client.frame_counter_server) {
				client.last_ack_frame = client.frame_counter_server + NetworkClientModel.DAY_TICKS;
				send.ack(client.frame_counter_server, client.token);
			}
			break;
		}
		case PACKET_SERVER_SYNC: {
			long sync_frame = packet.readUint32();
			long sync_seed1 = packet.readUint32();
			if(packet.getBuffer().position() != packet.getLength()) {
				long sync_seed2 = packet.readUint32();
			}
			break;
		}
		case PACKET_SERVER_COMMAND: {
			long cmdFrame = packet.readUint32();
			boolean myCmd = packet.readBool8();
			break;
		}
		case PACKET_SERVER_CHAT: {
			short networkActionId = packet.readUint8();
			long client_id = packet.readUint32();
			boolean selfSend = packet.readBool8();
			String message = packet.readString();
			BigInteger data = packet.readUint64();
			break;
		}
		case PACKET_SERVER_ERROR_QUIT:
		case PACKET_SERVER_QUIT:
		case PACKET_SERVER_JOIN:
		{
			long clientId = packet.readUint32();
			break;
		}
		case PACKET_SERVER_SHUTDOWN:
		case PACKET_SERVER_NEWGAME:
			break;
		case PACKET_SERVER_RCON: {
			char color = packet.readUint16();
			String rcon = packet.readString();
			break;
		}
		case PACKET_SERVER_MOVE: {
			long clientId = packet.readUint32();
			short companyId = packet.readUint8();
			break;
		}
		case PACKET_SERVER_CONFIG_UPDATE: {
			short maxCompagnies = packet.readUint8();
			short maxSpectators = packet.readUint8();
			break;
		}
		case PACKET_SERVER_COMPANY_UPDATE: {
			char passworded = packet.readUint16();
			break;
		}
		}
	}
}
