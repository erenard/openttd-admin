package com.openttd.network.constant;

import java.util.Arrays;

/**
 * See network_type.h for updates
 */
public interface NetworkType {

    /**
     * Vehicletypes in the order they are send in info packets.
     */
    public enum NetworkVehicleType {
        NETWORK_VEH_TRAIN,
        NETWORK_VEH_LORRY,
        NETWORK_VEH_BUS,
        NETWORK_VEH_PLANE,
        NETWORK_VEH_SHIP,
        NETWORK_VEH_END;

        public static int length() {
            return Arrays
                    .asList(NetworkVehicleType.values())
                    .indexOf(NetworkVehicleType.NETWORK_VEH_END);
        }
    };

    public enum DestType {
        DESTTYPE_BROADCAST, // /< Send message/notice to all clients (All)
        DESTTYPE_TEAM, // /< Send message/notice to everyone playing the same company (Team)
        DESTTYPE_CLIENT, // /< Send message/notice to only a certain client (Private)
    };

    /**
     * Actions that can be used for NetworkTextMessage
     */
    public enum NetworkAction {
        NETWORK_ACTION_JOIN,
        NETWORK_ACTION_LEAVE,
        NETWORK_ACTION_SERVER_MESSAGE,
        NETWORK_ACTION_CHAT,
        NETWORK_ACTION_CHAT_COMPANY,
        NETWORK_ACTION_CHAT_CLIENT,
        NETWORK_ACTION_GIVE_MONEY,
        NETWORK_ACTION_NAME_CHANGE,
        NETWORK_ACTION_COMPANY_SPECTATOR,
        NETWORK_ACTION_COMPANY_JOIN,
        NETWORK_ACTION_COMPANY_NEW,
        NETWORK_ACTION_KICKED,
        NETWORK_ACTION_EXTERNAL_CHAT,
    };

    public enum NetworkErrorCode {
        NETWORK_ERROR_GENERAL, // Try to use this one like never

        /* Signals from clients */
        NETWORK_ERROR_DESYNC,
        NETWORK_ERROR_SAVEGAME_FAILED,
        NETWORK_ERROR_CONNECTION_LOST,
        NETWORK_ERROR_ILLEGAL_PACKET,
        NETWORK_ERROR_NEWGRF_MISMATCH,
        /* Signals from servers */
        NETWORK_ERROR_NOT_AUTHORIZED,
        NETWORK_ERROR_NOT_EXPECTED,
        NETWORK_ERROR_WRONG_REVISION,
        NETWORK_ERROR_NAME_IN_USE,
        NETWORK_ERROR_WRONG_PASSWORD,
        NETWORK_ERROR_COMPANY_MISMATCH, // Happens in CLIENT_COMMAND
        NETWORK_ERROR_KICKED,
        NETWORK_ERROR_CHEATER,
        NETWORK_ERROR_FULL,
        NETWORK_ERROR_TOO_MANY_COMMANDS,
        NETWORK_ERROR_TIMEOUT_PASSWORD,
        NETWORK_ERROR_TIMEOUT_COMPUTER,
        NETWORK_ERROR_TIMEOUT_MAP,
        NETWORK_ERROR_TIMEOUT_JOIN,
        NETWORK_ERROR_INVALID_CLIENT_NAME,
        NETWORK_ERROR_END;

        public static NetworkErrorCode valueOf(int order) {
            for (NetworkErrorCode value : values()) {
                if (value.ordinal() == order) {
                    return value;
                }
            }
            return null;
        }
    };

}
