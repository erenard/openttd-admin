package com.openttd.network.udp;

import java.io.IOException;
import java.util.HashMap;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.openttd.network.constant.PacketUdpType;
import com.openttd.network.core.Packet;
import com.openttd.network.core.Socket;

public class GameInfoClient {

	private static final Logger log = LoggerFactory.getLogger(GameInfoClient.class);

	private Socket socket;

	public GameInfoClient(String host, Integer port) throws IOException {
		this.socket = Socket.newUdpSocket(host, port);
	}

	GameInfo gameInfo = new GameInfo();

	// network/network_server.cpp -> NetworkSocketHandler::SendCompanyInformation
	synchronized void readCompany(Packet packet) {
		short number = packet.readUint8();
		CompanyInfo company = gameInfo.getCompany(number);
		if (company == null) company = new CompanyInfo(number);
		company.setName(packet.readString());
		company.setInauguratedYear(packet.readUint32());
		company.setValue(packet.readInt64());
		company.setMoney(packet.readInt64());
		company.setIncome(packet.readInt64());
		company.setPerformance(packet.readUint16());
		company.setUsePassword(packet.readBool8());
		int[] vehicules = new int[5];
		vehicules[0] = packet.readUint16();
		vehicules[1] = packet.readUint16();
		vehicules[2] = packet.readUint16();
		vehicules[3] = packet.readUint16();
		vehicules[4] = packet.readUint16();
		company.setVehicules(vehicules);
		int[] stations = new int[5];
		stations[0] = packet.readUint16();
		stations[1] = packet.readUint16();
		stations[2] = packet.readUint16();
		stations[3] = packet.readUint16();
		stations[4] = packet.readUint16();
		company.setStations(stations);
		gameInfo.putCompany(number, company);
		// Do not remove
		packet.readString();
	}

	// network/core/udp.cpp -> NetworkUDPSocketHandler::Send_NetworkGameInfo
	void readGameInfo(Packet packet) {
		int version = packet.readUint8();
		gameInfo.setVersion(version);
		if (version >= 4) {
			int grfCount = packet.readUint8();
			gameInfo.setGrfCount(grfCount);
			gameInfo.setGrfs(new HashMap<byte[], byte[]>());
			for (int i = 0; i < grfCount; i++) {
				byte[] grfid = packet.readBytes(4);
				byte[] md5sum = packet.readBytes(16);
				gameInfo.getGrfs().put(grfid, md5sum);
			}
		}
		if (version >= 3) {
			gameInfo.setGameDate(packet.readUint32());
			gameInfo.setGameStartDate(packet.readUint32());
		}
		if (version >= 2) {
			gameInfo.setCompaniesMax(packet.readUint8());
			gameInfo.setCompaniesOn(packet.readUint8());
			gameInfo.setSpectatorsMax(packet.readUint8());
		}
		if (version >= 1) {
			gameInfo.setServerName(packet.readString());
			gameInfo.setServerRevision(packet.readString());
			gameInfo.setServerLanguage(packet.readUint8());
			gameInfo.setUsePassword(packet.readBool8());
			gameInfo.setClientsMax(packet.readUint8());
			gameInfo.setClientsOn(packet.readUint8());
			gameInfo.setSpectatorsOn(packet.readUint8());
			if (version < 3) {
				gameInfo.setGameDate(packet.readUint16());
				gameInfo.setGameStartDate(packet.readUint16());
			}
			gameInfo.setMapName(packet.readString());
			gameInfo.setMapWidth(packet.readUint16());
			gameInfo.setMapHeight(packet.readUint16());
			gameInfo.setMapSet(packet.readUint8());
			gameInfo.setDedicated(packet.readBool8());
		}
	}

	private class Updater extends Thread {
		@Override
		public void run() {
			try {
				socket.sendEmptyPacket(PacketUdpType.PACKET_UDP_CLIENT_FIND_SERVER);
				receive: while (true) {
					List<Packet> packets = socket.receive();
					for (Packet packet : packets) {
						PacketUdpType packetUdpType = PacketUdpType.valueOf(packet.readUint8());
						log.debug(packetUdpType.toString());
						switch (packetUdpType) {
						case PACKET_UDP_SERVER_RESPONSE: {
							readGameInfo(packet);
							socket.sendEmptyPacket(PacketUdpType.PACKET_UDP_CLIENT_DETAIL_INFO);
							break;
						}
						case PACKET_UDP_SERVER_DETAIL_INFO: {
							int version = packet.readUint8();
							gameInfo.setVersion(version);
							short companyCount = packet.readUint8();
							for (short i = 0; i < companyCount; i++) {
								readCompany(packet);
							}
							break receive;
						}
						}
					}
				}
			} catch (IOException e) {
				log.error(e.getMessage(), e);
			}
		}
	};

	private Updater updater = null;

	public GameInfo getGameInfo() {
		gameInfo = new GameInfo();
		if (updater == null || !updater.isAlive()) {
			updater = new Updater();
			updater.start();
		}
		try {
			updater.join();
		} catch (InterruptedException e) {
			e.printStackTrace();
		}
		return gameInfo;
	}

	/**
	 * main for test purpose
	 * 
	 * @throws IOException
	 */
	public static void main(String[] args) throws IOException {
		log.info("Testing UDP GameInfoClient");
		log.info("Connection...");
		GameInfoClient gameInfoClient = new GameInfoClient("localhost", 3979);
		log.info("Connected !");
		log.info("Requesting...");
		GameInfo gameInfo = gameInfoClient.getGameInfo();
		log.info("Results :");
		log.info(gameInfo.toString());
	}
}
