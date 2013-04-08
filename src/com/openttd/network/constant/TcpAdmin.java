package com.openttd.network.constant;

/**
 * See tcp_admin.h for updates
 */
public class TcpAdmin {
	public enum PacketAdminType implements PacketType {
		ADMIN_PACKET_ADMIN_JOIN,             ///< The admin announces and authenticates itself to the server.
		ADMIN_PACKET_ADMIN_QUIT,             ///< The admin tells the server that it is quitting.
		ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY, ///< The admin tells the server the update frequency of a particular piece of information.
		ADMIN_PACKET_ADMIN_POLL,             ///< The admin explicitly polls for a piece of information.
		ADMIN_PACKET_ADMIN_CHAT,             ///< The admin sends a chat message to be distributed.
		ADMIN_PACKET_ADMIN_RCON,             ///< The admin sends a remote console command.
		ADMIN_PACKET_ADMIN_GAMESCRIPT;       ///< The admin sends a JSON string for the GameScript.

		public static PacketAdminType valueOf(int order) {
			for (PacketAdminType value : values()) {
				if (value.ordinal() == order) {
					return value;
				}
			}
			return null;
		}

		@Override
		public int getType() {
			return TCP;
		}
	}

	public enum PacketServerType implements PacketType {
		ADMIN_PACKET_SERVER_FULL,            ///< The server tells the admin it cannot accept the admin.
		ADMIN_PACKET_SERVER_BANNED,          ///< The server tells the admin it is banned.
		ADMIN_PACKET_SERVER_ERROR,           ///< The server tells the admin an error has occurred.
		ADMIN_PACKET_SERVER_PROTOCOL,        ///< The server tells the admin its protocol version.
		ADMIN_PACKET_SERVER_WELCOME,         ///< The server welcomes the admin to a game.
		ADMIN_PACKET_SERVER_NEWGAME,         ///< The server tells the admin its going to start a new game.
		ADMIN_PACKET_SERVER_SHUTDOWN,        ///< The server tells the admin its shutting down.

		ADMIN_PACKET_SERVER_DATE,            ///< The server tells the admin what the current game date is.
		ADMIN_PACKET_SERVER_CLIENT_JOIN,     ///< The server tells the admin that a client has joined.
		ADMIN_PACKET_SERVER_CLIENT_INFO,     ///< The server gives the admin information about a client.
		ADMIN_PACKET_SERVER_CLIENT_UPDATE,   ///< The server gives the admin an information update on a client.
		ADMIN_PACKET_SERVER_CLIENT_QUIT,     ///< The server tells the admin that a client quit.
		ADMIN_PACKET_SERVER_CLIENT_ERROR,    ///< The server tells the admin that a client caused an error.
		ADMIN_PACKET_SERVER_COMPANY_NEW,     ///< The server tells the admin that a new company has started.
		ADMIN_PACKET_SERVER_COMPANY_INFO,    ///< The server gives the admin information about a company.
		ADMIN_PACKET_SERVER_COMPANY_UPDATE,  ///< The server gives the admin an information update on a company.
		ADMIN_PACKET_SERVER_COMPANY_REMOVE,  ///< The server tells the admin that a company was removed.
		ADMIN_PACKET_SERVER_COMPANY_ECONOMY, ///< The server gives the admin some economy related company information.
		ADMIN_PACKET_SERVER_COMPANY_STATS,   ///< The server gives the admin some statistics about a company.
		ADMIN_PACKET_SERVER_CHAT,            ///< The server received a chat message and relays it.
		ADMIN_PACKET_SERVER_RCON,            ///< The server's reply to a remove console command.
		ADMIN_PACKET_SERVER_CONSOLE,         ///< The server gives the admin the data that got printed to its console.
		ADMIN_PACKET_SERVER_CMD_NAMES,       ///< The server sends out the names of the DoCommands to the admins.
		ADMIN_PACKET_SERVER_CMD_LOGGING,     ///< The server gives the admin copies of incoming command packets.
		ADMIN_PACKET_SERVER_GAMESCRIPT,      ///< The server gives the admin information from the GameScript in JSON.

		INVALID_ADMIN_PACKET;

		private final static int PacketServerTypeStartValue = 100;

		public static PacketServerType valueOf(int order) {
			for (PacketServerType value : values()) {
				if (value.ordinal() + PacketServerTypeStartValue == order) {
					return value;
				}
			}
			return INVALID_ADMIN_PACKET;
		}

		@Override
		public int getType() {
			return TCP;
		}
	}

	/** Update types an admin can register a frequency for */
	public enum AdminUpdateType {
		ADMIN_UPDATE_DATE,            ///< Updates about the date of the game.
		ADMIN_UPDATE_CLIENT_INFO,     ///< Updates about the information of clients.
		ADMIN_UPDATE_COMPANY_INFO,    ///< Updates about the generic information of companies.
		ADMIN_UPDATE_COMPANY_ECONOMY, ///< Updates about the economy of companies.
		ADMIN_UPDATE_COMPANY_STATS,   ///< Updates about the statistics of companies.
		ADMIN_UPDATE_CHAT,            ///< The admin would like to have chat messages.
		ADMIN_UPDATE_CONSOLE,         ///< The admin would like to have console messages.
		ADMIN_UPDATE_CMD_NAMES,       ///< The admin would like a list of all DoCommand names.
		ADMIN_UPDATE_CMD_LOGGING,     ///< The admin would like to have DoCommand information.
		ADMIN_UPDATE_GAMESCRIPT,      ///< The admin would like to have gamescript messages.
		ADMIN_UPDATE_END;             ///< Must ALWAYS be on the end of this list!! (period)
		public static AdminUpdateType valueOf(int order) {
			for (AdminUpdateType value : values()) {
				if (value.ordinal() == order) {
					return value;
				}
			}
			return ADMIN_UPDATE_END;
		}
	};

	/** Update frequencies an admin can register. */
	public enum AdminUpdateFrequency {
		ADMIN_FREQUENCY_POLL(0x01),		 ///< The admin can poll this.
		ADMIN_FREQUENCY_DAILY(0x02),	 ///< The admin gets information about this on a daily basis.
		ADMIN_FREQUENCY_WEEKLY(0x04),	 ///< The admin gets information about this on a weekly basis.
		ADMIN_FREQUENCY_MONTHLY(0x08),	 ///< The admin gets information about this on a monthly basis.
		ADMIN_FREQUENCY_QUARTERLY(0x10), ///< The admin gets information about this on a quarterly basis.
		ADMIN_FREQUENCY_ANUALLY(0x20),	 ///< The admin gets information about this on a yearly basis.
		ADMIN_FREQUENCY_AUTOMATIC(0x40); ///< The admin gets information about this when it changes.
		public int mask;

		private AdminUpdateFrequency(int s) {
			this.mask = s;
		}
	};

	/** Reasons for removing a company - communicated to admins. */
	public enum AdminCompanyRemoveReason {
		ADMIN_CRR_MANUAL,	 ///< The company is manually removed.
		ADMIN_CRR_AUTOCLEAN, ///< The company is removed due to autoclean.
		ADMIN_CRR_BANKRUPT,	 ///< The company went belly-up.
		ADMIN_CRR_END		 ///< Sentinel for end.
	};

}
