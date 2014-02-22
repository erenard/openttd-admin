package com.openttd.network.admin;

import java.io.IOException;
import java.math.BigInteger;
import java.net.ConnectException;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.openttd.admin.model.Game;
import com.openttd.network.constant.TcpAdmin.PacketServerType;
import com.openttd.network.core.Configuration;
import com.openttd.network.core.Packet;
import com.openttd.network.core.Socket;

/**
 * Network listening thread, own a second speaking thread.
 */
public class NetworkAdmin extends Thread {

	private static final Logger log = LoggerFactory.getLogger(NetworkAdmin.class);

	private final Configuration configuration;

	public NetworkAdmin(Configuration configuration) {
		this.configuration = configuration;
	}

	// Listen and speak thread token
	private boolean running;
	// Network game model
	private NetworkModel networkModel;
	// Network speak thread
	private NetworkAdminSender send;
	// Client game model
	private Game game;

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

	@Override
	public void run() {
		log.debug("started");
		this.running = true;
		// Auto-Reconnection
		while (this.running) {
			Socket socket = null;
			try {
				log.debug("connecting...");
				socket = Socket.newTcpSocket(configuration.host, configuration.adminPort);
				send = new NetworkAdminSender(socket);
				send.startup();
				networkModel = new NetworkModel();
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

	/**
	 * Extract infos from the server's packets,
	 * Update the networkModel with these infos,
	 * Update the server's protocols,
	 * Dispatch a game related event.
	 *
	 * See network_admin.cpp for updates, search "Send::"
	 * @param packet
	 * @throws NetworkException
	 */
	@SuppressWarnings("unused")
	private void handlePacket(Packet packet) throws NetworkException {
		int packetTypeId = packet.readUint8();
		PacketServerType packetType = PacketServerType.valueOf(packetTypeId);
		log.debug("Rcv " + packetType);
		NetworkAdminEvent event = new NetworkAdminEvent(packetType);
		switch (packetType) {
		/* ********************
		 * Clients management *
		 * ********************
		 */
		/**
		 * Notification of a new client:
		 * uint32 ID of the new client.
		 */
		case ADMIN_PACKET_SERVER_CLIENT_JOIN: {
			// This packet arrives after CLIENT_INFO
			long clientId = packet.readUint32();
			event.setClientEvent(packetType, clientId);
			break;
		}
		/**
		 * Client information of a specific client:
		 * uint32 ID of the client.
		 * string Network address of the client.
		 * string Name of the client.
		 * uint8 Language of the client.
		 * uint32 Date the client joined the game.
		 * uint8 ID of the company the client is playing as (255 for spectators).
		 */
		case ADMIN_PACKET_SERVER_CLIENT_INFO: {
			long clientId = packet.readUint32();
			event.setClientEvent(packetType, clientId);
			Client client = networkModel.retreiveClient(clientId);
			client.setIp(packet.readString());
			client.setName(packet.readString());
			client.setLanguage(packet.readUint8());
			client.setJoinDate(packet.readUint32());
			client.setCompanyId(packet.readUint8());
			break;
		}
		/**
		 * Client update details on a specific client (e.g. after rename or move):
		 * uint32 ID of the client.
		 * string Name of the client.
		 * uint8 ID of the company the client is playing as (255 for spectators).
		 */
		case ADMIN_PACKET_SERVER_CLIENT_UPDATE: {
			long clientId = packet.readUint32();
			event.setClientEvent(packetType, clientId);
			Client client = networkModel.retreiveClient(clientId);
			client.setName(packet.readString());
			client.setCompanyId(packet.readUint8());
			break;
		}
		/**
		 * Notification about a client error (and thus the clients disconnection).
		 * uint32 ID of the client that made the error.
		 * uint8 Error the client made (see NetworkErrorCode).
		 */
		case ADMIN_PACKET_SERVER_CLIENT_ERROR: {
			long clientId = packet.readUint32();
			short error = packet.readUint8();
			event.setClientEvent(packetType, clientId);

			networkModel.deleteClient(clientId);
			break;
		}
		/**
		 * Notification about a client leaving the game.
		 * uint32 ID of the client that just left.
		 */
		case ADMIN_PACKET_SERVER_CLIENT_QUIT: {
			long clientId = packet.readUint32();
			event.setClientEvent(packetType, clientId);
			networkModel.deleteClient(clientId);
			break;
		}
		/* **********************
		 * Companies management *
		 * **********************
		 */
		/**
		 * Notification of a new company:
		 * uint8 ID of the new company.
		 */
		case ADMIN_PACKET_SERVER_COMPANY_NEW: {
			short companyId = packet.readUint8();
			event.setCompanyEvent(packetType, companyId);
			// This packet arrives after COMPANY_INFO
			break;
		}
		/**
		 * Company information on a specific company:
		 * uint8 ID of the company.
		 * string Name of the company.
		 * string Name of the companies manager.
		 * uint8 Main company colour.
		 * bool Company is password protected.
		 * uint32 Year the company was inaugurated.
		 * bool Company is an AI.
		 */
		case ADMIN_PACKET_SERVER_COMPANY_INFO: {
			short companyId = packet.readUint8();
			event.setCompanyEvent(packetType, companyId);
			Company company = networkModel.retreiveCompany(companyId);
			company.setName(packet.readString());
			company.setPresident(packet.readString());
			company.setColor(packet.readUint8());
			company.setUsePassword(packet.readBool8());
			company.setInauguratedYear(packet.readUint32());
			company.setAi(packet.readBool8());
			company.setBankruptcy(packet.readUint8());
			short[] shareOwners = new short[4];
			for (short i = 0; i < 4; i++) {
				shareOwners[i] = packet.readUint8();
			}
			company.setShareOwners(shareOwners);
			break;
		}
		/**
		 * Company information of a specific company:
		 * uint8 ID of the company.
		 * string Name of the company.
		 * string Name of the companies manager.
		 * uint8 Main company colour.
		 * bool Company is password protected.
		 * uint8 Quarters of bankruptcy.
		 * uint8 Owner of share 1.
		 * uint8 Owner of share 2.
		 * uint8 Owner of share 3.
		 * uint8 Owner of share 4.
		 */
		case ADMIN_PACKET_SERVER_COMPANY_UPDATE: {
			short companyId = packet.readUint8();
			event.setCompanyEvent(packetType, companyId);
			Company company = networkModel.retreiveCompany(companyId);
			company.setName(packet.readString());
			company.setPresident(packet.readString());
			company.setColor(packet.readUint8());
			company.setUsePassword(packet.readBool8());
			company.setBankruptcy(packet.readUint8());
			short[] shareOwners = new short[4];
			for (short i = 0; i < 4; i++) {
				shareOwners[i] = packet.readUint8();
			}
			company.setShareOwners(shareOwners);
			break;
		}
		/**
		 * Economy update of a specific company:
		 * uint8 ID of the company.
		 * int64 Money.
		 * uint64 Loan.
		 * int64 Income.
		 * uint16 Delivered cargo.
		 * uint64 Company value (last quarter).
		 * uint16 Performance (last quarter).
		 * uint16 Delivered cargo (last quarter).
		 * uint64 Company value (previous quarter).
		 * uint16 Performance (previous quarter).
		 * uint16 Delivered cargo (previous quarter).
		 */
		case ADMIN_PACKET_SERVER_COMPANY_ECONOMY: {
			short companyId = packet.readUint8();
			event.setCompanyEvent(packetType, companyId);
			Company company = networkModel.retreiveCompany(companyId);
			company.setMoney(packet.readInt64());
			company.setLoan(packet.readUint64());
			company.setIncome(packet.readInt64());
			company.setDeliveredCargo(packet.readUint16());

			company.setLastValue(packet.readUint64());
			company.setLastPerformance(packet.readUint16());
			company.setLastDeliveredCargo(packet.readUint16());

			company.setPreviousValue(packet.readUint64());
			company.setPreviousPerformance(packet.readUint16());
			company.setPreviousDeliveredCargo(packet.readUint16());
			break;
		}
		/**
		 * Company statistics on stations and vehicles:
		 * uint8 ID of the company.
		 * uint16 Number of trains.
		 * uint16 Number of lorries.
		 * uint16 Number of busses.
		 * uint16 Number of planes.
		 * uint16 Number of ships.
		 * uint16 Number of train stations.
		 * uint16 Number of lorry stations.
		 * uint16 Number of bus stops.
		 * uint16 Number of airports and heliports.
		 * uint16 Number of harbours.
		 */
		case ADMIN_PACKET_SERVER_COMPANY_STATS: {
			short companyId = packet.readUint8();
			event.setCompanyEvent(packetType, companyId);
			Company company = networkModel.retreiveCompany(companyId);
			char[] vehicules = new char[5];
			for (int i = 0; i < 5; i++) {
				vehicules[i] = packet.readUint16();
			}
			company.setVehicules(vehicules);
			char[] stations = new char[5];
			for (int i = 0; i < 5; i++) {
				stations[i] = packet.readUint16();
			}
			company.setStations(stations);
			break;
		}
		/**
		 * Notification about a removed company (e.g. due to banrkuptcy).
		 * uint8 ID of the company.
		 * uint8 Reason for being removed (see #AdminCompanyRemoveReason).
		 */
		case ADMIN_PACKET_SERVER_COMPANY_REMOVE: {
			short companyId = packet.readUint8();
			short reason = packet.readUint8();
			event.setCompanyEvent(packetType, companyId);
			networkModel.deleteCompany(companyId);
			break;
		}
		/**
		 * Send incoming command packets to the admin network.
		 * This is for logging purposes only.
		 * NOTICE: Data provided with this packet is not stable and will not be
		 * treated as such. Do not rely on IDs or names to be constant
		 * across different versions / revisions of OpenTTD.
		 * Data provided in this packet is for logging purposes only.
		 * uint32 ID of the client sending the command.
		 * uint8 ID of the company (0..MAX_COMPANIES-1).
		 * uint16 ID of the command.
		 * uint32 P1 (variable data passed to the command).
		 * uint32 P2 (variable data passed to the command).
		 * uint32 Tile where this is taking place.
		 * string Text passed to the command.
		 * uint32 Frame of execution.
		 */
		case ADMIN_PACKET_SERVER_CMD_LOGGING: {
			long clientId = packet.readUint32();
			int companyId = packet.readUint8();
			int commandId = packet.readUint16();
			long p1 = packet.readUint32();
			long p2 = packet.readUint32();
			long tile = packet.readUint32();
			String text = packet.readString();
			long frame = packet.readUint32();
			break;
		}
		/**
		 * Send DoCommand names to the bot upon request only.
		 * Multiple of these packets can follow each other in order to provide
		 * all known DoCommand names.
		 * NOTICE: Data provided with this packet is not stable and will not be
		 * treated as such. Do not rely on IDs or names to be constant
		 * across different versions / revisions of OpenTTD.
		 * Data provided in this packet is for logging purposes only.
		 * These three fields are repeated until the packet is full:
		 * bool Data to follow.
		 * uint16 ID of the DoCommand.
		 * string Name of DoCommand.
		 */
		case ADMIN_PACKET_SERVER_CMD_NAMES: {
			while (packet.readBool8()) {
				int cmdId = packet.readUint16();
				String cmdName = packet.readString();
			}
			break;
		}
		/* ****************
		 * Others packets *
		 * ****************
		 */
		/**
		 * Send what would be printed on the server's console also into the admin network.
		 * string The origin of the text, e.g. "console" for console, or "net" for network related (debug) messages.
		 * string Text as found on the console of the server.
		 */
		case ADMIN_PACKET_SERVER_CONSOLE: {
			String origin = packet.readString();
			String message = packet.readString();
			event.setConsoleEvent(origin, message);
			break;
		}
		/**
		 * Send chat from the game into the admin network:
		 * uint8 Action such as NETWORK_ACTION_CHAT_CLIENT (see #NetworkAction).
		 * uint8 Destination type such as DESTTYPE_BROADCAST (see #DestType).
		 * uint32 ID of the client who sent this message.
		 * string Message.
		 * uint64 Money (only when it is a 'give money' action).
		 */
		case ADMIN_PACKET_SERVER_CHAT: {
			short actionId = packet.readUint8();
			short destinationType = packet.readUint8();
			long clientId = packet.readUint32();
			String message = packet.readString();
			BigInteger data = packet.readUint64();
			event.setChatEvent(packetType, clientId, message);
			break;
		}
		/**
		 * Send the current date of the game:
		 * uint32 Current game date.
		 */
		case ADMIN_PACKET_SERVER_DATE: {
			GameInfo gameInfo = networkModel.getGameInfo();
			long date = packet.readUint32();
			gameInfo.setCurrentDate(date);
			event.setDateEvent(packetType, date);
			break;
		}
		/**
		 * An error was caused by this admin connection (connection gets closed).
		 * uint8 NetworkErrorCode the error caused.
		 */
		case ADMIN_PACKET_SERVER_ERROR: {
			throw new NetworkException(packet.readUint8());
		}
		/**
		 * Notification about a newgame.
		 */
		case ADMIN_PACKET_SERVER_NEWGAME: {
			break;
		}
		/**
		 * Inform a just joined admin about the protocol specifics:
		 * uint8 Protocol version.
		 * bool Further protocol data follows (repeats through all update packet types).
		 * uint16 Update packet type.
		 * uint16 Frequencies allowed for this update packet (bitwise).
		 */
		case ADMIN_PACKET_SERVER_PROTOCOL: {
			Protocol protocol = new Protocol(packet.readUint8());
			while (packet.readBool8()) {
				protocol.putProtocol(packet.readUint16(), packet.readUint16());
			}
			send.setProtocol(protocol);
			log.debug(protocol.toString());
			break;
		}
		/**
		 * Result of an rcon command:
		 * uint16 Colour as it would be used on the server or a client.
		 * string Output of the executed command.
		 */
		case ADMIN_PACKET_SERVER_RCON: {
			char color = packet.readUint16();
			String message = packet.readString();
			event.setRConEvent(packetType, color, message);
			break;
		}
		/**
		 * Send a JSON string to the current active GameScript.
		 * json  JSON string for the GameScript.
		 */
		case ADMIN_PACKET_SERVER_GAMESCRIPT: {
			String json = packet.readString();
			event.setGameScriptEvent(packetType, json);
			break;
		}
		/**
		 * Notify the admin connection that the rcon command has finished.
		 * string The command as requested by the admin connection.
		 */
		case ADMIN_PACKET_SERVER_RCON_END: {
			String rcon = packet.readString();
			event.setRConEndEvent(packetType, rcon);
			break;
		}
		/**
		 * The source IP address is banned (connection gets closed).
		 */
		case ADMIN_PACKET_SERVER_BANNED:
			/**
			 * The server is full (connection gets closed).
			 */
		case ADMIN_PACKET_SERVER_FULL:
			/**
			 * Notification about the server shutting down.
			 */
		case ADMIN_PACKET_SERVER_SHUTDOWN: {
			throw new NetworkException(packetType.toString());
		}
		/**
		 * Welcome a connected admin to the game:
		 * string Name of the Server (e.g. as advertised to master server).
		 * string OpenTTD version string.
		 * bool Server is dedicated.
		 * string Name of the Map.
		 * uint32 Random seed of the Map.
		 * uint8 Landscape of the Map.
		 * uint32 Start date of the Map.
		 * uint16 Map width.
		 * uint16 Map height.
		 */
		case ADMIN_PACKET_SERVER_WELCOME: {
			GameInfo gameInfo = new GameInfo();

			gameInfo.setServerName(packet.readString());
			gameInfo.setServerRevision(packet.readString());
			gameInfo.setServerDedicated(packet.readBool8());

			gameInfo.setMapName(packet.readString());
			gameInfo.setMapSeed(packet.readUint32());
			gameInfo.setMapSet(packet.readUint8());

			gameInfo.setStartDate(packet.readUint32());
			gameInfo.setMapWidth(packet.readUint16());
			gameInfo.setMapHeight(packet.readUint16());

			networkModel.setGameInfo(gameInfo);
			break;
		}
		/**
		 * Send a ping-reply (pong) to the admin that sent us the ping packet.
		 * uint32  Integer identifier - should be the same as read from the admins ping packet.
		 */
		case ADMIN_PACKET_SERVER_PONG: {
			long id = packet.readUint32();
			event.setPongEvent(id);
			break;
		}
		case INVALID_ADMIN_PACKET:
			throw new NetworkException(packetType.toString());
		}
		if (game != null) {
			game.updateAdmin(event, networkModel.copy(), send);
		}
	}

	public NetworkAdminSender getSend() {
		return send;
	}

	public void setGame(Game game) {
		this.game = game;
	}
}
