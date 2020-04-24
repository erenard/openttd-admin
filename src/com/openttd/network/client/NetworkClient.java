package com.openttd.network.client;

import java.io.IOException;
import java.io.PrintWriter;
import java.io.StringWriter;
import java.math.BigInteger;
import java.net.ConnectException;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.openttd.network.admin.NetworkException;
import com.openttd.network.constant.NetworkType.NetworkErrorCode;
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
			log.error(NetworkErrorCode.valueOf(errorId).toString());
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
			send.gamePassword(configuration.password);
			break;
		}
		case PACKET_SERVER_NEED_COMPANY_PASSWORD: {
			client.passwordGameSeed = packet.readUint32();
			client.passwordServerId = packet.readString();
			send.companyPassword(hashPassword(configuration.password, client.passwordServerId, client.passwordGameSeed));
			break;
		}
		case PACKET_SERVER_WELCOME: {
			client.clientId = packet.readUint32();
			client.passwordGameSeed = packet.readUint32();
			client.passwordServerId = packet.readString();
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
		default:
			break;
		}
	}

	/**
	 * See network.cpp, method: GenerateCompanyPasswordHash
	 */
	public String hashPassword(String password, String passwordServerId, long passwordGameSeed) {
		final int NETWORK_SERVER_ID_LENGTH = passwordServerId.length();
		//if (StrEmpty(password)) return password;
		if (password == null || password.length() == 0) return password;

		//char salted_password[NETWORK_SERVER_ID_LENGTH];
		//memset(salted_password, 0, sizeof(salted_password));
		byte [] salted_password = new byte[NETWORK_SERVER_ID_LENGTH];

		//snprintf(salted_password, sizeof(salted_password), "%s", password);
		for (int i = 0; i < NETWORK_SERVER_ID_LENGTH; i++) {
			if(password.length() < NETWORK_SERVER_ID_LENGTH) {
				salted_password[i] = (byte) ' ';
			} else {
				salted_password[i] = (byte) password.charAt(i);
			}
		}
		/* Add the game seed and the server's ID as the salt. */
		for (int i = 0; i < NETWORK_SERVER_ID_LENGTH - 1; i++) {
			salted_password[i] ^= passwordServerId.charAt(i) ^ (passwordGameSeed >> (i % 32));
		}

		try {
			//Md5 checksum;
			MessageDigest checksum = MessageDigest.getInstance("MD5");
			/* Generate the MD5 hash */
			//checksum.Append(salted_password, sizeof(salted_password) - 1);
			checksum.update(salted_password, 0, NETWORK_SERVER_ID_LENGTH - 1);
			
			//checksum.Finish(digest);
			byte [] digest = checksum.digest();
			
			StringWriter hashed_password = new StringWriter(NETWORK_SERVER_ID_LENGTH);
			PrintWriter out = new PrintWriter(hashed_password);
			
			//for (int di = 0; di < 16; di++) sprintf(hashed_password + di * 2, "%02x", digest[di]);
			for (int di = 0; di < 16; di++) {
				out.printf("%02x", digest[di]);
			}
			
			//hashed_password[lengthof(hashed_password) - 1] = '\0';
			hashed_password.write(new char [] {'\0'}, hashed_password.toString().length() - 1, 1);
			
			return hashed_password.toString();
		} catch (NoSuchAlgorithmException e) {
			return "";
		}
	}
}
