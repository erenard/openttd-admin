package com.openttd.server;

import java.util.EnumMap;
import java.util.Map;

import com.openttd.network.constant.TcpAdmin.AdminUpdateFrequency;
import com.openttd.network.constant.TcpAdmin.AdminUpdateType;

/**
 * Values copied from OpenTTD 15.3:
 *
 * - src/network/core/config.h: NETWORK_GAME_ADMIN_VERSION = 3
 * - src/network/network_admin.cpp: _admin_update_type_frequencies
 *
 * Keep in sync with the OpenTTD version in GameVersion; a version bump is one
 * commit per release (see AGENTS.md).
 */
public class UpstreamAdminProtocol {

    public static final short GAME_ADMIN_VERSION = 3;

    public static final Map<AdminUpdateType, Integer> FREQUENCIES = frequencies();

    private static Map<AdminUpdateType, Integer> frequencies() {
        Map<AdminUpdateType, Integer> frequencies = new EnumMap<>(AdminUpdateType.class);
        frequencies.put(AdminUpdateType.ADMIN_UPDATE_DATE,
                masks(AdminUpdateFrequency.ADMIN_FREQUENCY_POLL, AdminUpdateFrequency.ADMIN_FREQUENCY_DAILY,
                        AdminUpdateFrequency.ADMIN_FREQUENCY_WEEKLY, AdminUpdateFrequency.ADMIN_FREQUENCY_MONTHLY,
                        AdminUpdateFrequency.ADMIN_FREQUENCY_QUARTERLY, AdminUpdateFrequency.ADMIN_FREQUENCY_ANUALLY));
        frequencies.put(AdminUpdateType.ADMIN_UPDATE_CLIENT_INFO,
                masks(AdminUpdateFrequency.ADMIN_FREQUENCY_POLL, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC));
        frequencies.put(AdminUpdateType.ADMIN_UPDATE_COMPANY_INFO,
                masks(AdminUpdateFrequency.ADMIN_FREQUENCY_POLL, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC));
        frequencies.put(AdminUpdateType.ADMIN_UPDATE_COMPANY_ECONOMY,
                masks(AdminUpdateFrequency.ADMIN_FREQUENCY_POLL, AdminUpdateFrequency.ADMIN_FREQUENCY_WEEKLY,
                        AdminUpdateFrequency.ADMIN_FREQUENCY_MONTHLY, AdminUpdateFrequency.ADMIN_FREQUENCY_QUARTERLY,
                        AdminUpdateFrequency.ADMIN_FREQUENCY_ANUALLY));
        frequencies.put(AdminUpdateType.ADMIN_UPDATE_COMPANY_STATS,
                masks(AdminUpdateFrequency.ADMIN_FREQUENCY_POLL, AdminUpdateFrequency.ADMIN_FREQUENCY_WEEKLY,
                        AdminUpdateFrequency.ADMIN_FREQUENCY_MONTHLY, AdminUpdateFrequency.ADMIN_FREQUENCY_QUARTERLY,
                        AdminUpdateFrequency.ADMIN_FREQUENCY_ANUALLY));
        // No Poll bit for these four: polling them makes the server reply
        // ADMIN_PACKET_SERVER_ERROR (NETWORK_ERROR_ILLEGAL_PACKET).
        frequencies.put(AdminUpdateType.ADMIN_UPDATE_CHAT, masks(AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC));
        frequencies.put(AdminUpdateType.ADMIN_UPDATE_CONSOLE, masks(AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC));
        frequencies.put(AdminUpdateType.ADMIN_UPDATE_CMD_NAMES, masks(AdminUpdateFrequency.ADMIN_FREQUENCY_POLL));
        frequencies.put(AdminUpdateType.ADMIN_UPDATE_CMD_LOGGING, masks(AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC));
        frequencies.put(AdminUpdateType.ADMIN_UPDATE_GAMESCRIPT, masks(AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC));
        return frequencies;
    }

    public static int masks(AdminUpdateFrequency... frequencies) {
        int mask = 0;
        for (AdminUpdateFrequency frequency : frequencies) {
            mask |= frequency.mask;
        }
        return mask;
    }

    /**
     * The PROTOCOL packet as sent by ServerNetworkAdminSocketHandler::SendProtocol().
     *
     * @return packet writer
     */
    public static PacketWriter protocolPacket() {
        PacketWriter packet = new PacketWriter(com.openttd.network.constant.TcpAdmin.PacketServerType.ADMIN_PACKET_SERVER_PROTOCOL);
        packet.writeUint8(GAME_ADMIN_VERSION);
        for (AdminUpdateType type : AdminUpdateType.values()) {
            if (type == AdminUpdateType.ADMIN_UPDATE_END) {
                break;
            }
            packet.writeBool8(true);
            packet.writeUint16(type.ordinal());
            packet.writeUint16(FREQUENCIES.get(type));
        }
        packet.writeBool8(false);
        return packet;
    }

    /**
     * The WELCOME packet as sent by ServerNetworkAdminSocketHandler::SendWelcome().
     *
     * @return packet writer
     */
    public static PacketWriter welcomePacket() {
        PacketWriter packet = new PacketWriter(com.openttd.network.constant.TcpAdmin.PacketServerType.ADMIN_PACKET_SERVER_WELCOME);
        packet.writeString("Strategyboard test server");
        packet.writeString("15.3");
        packet.writeBool8(true);
        packet.writeString(""); // Used to be map-name
        packet.writeUint32(123456789L);
        packet.writeUint8(1); // Landscape: Temperate
        packet.writeUint32(1950 * 365 + 1); // 1950-01-01
        packet.writeUint16(64);
        packet.writeUint16(64);
        return packet;
    }
}
