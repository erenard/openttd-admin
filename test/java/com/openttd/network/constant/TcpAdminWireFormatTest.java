package com.openttd.network.constant;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertNull;
import static org.junit.jupiter.api.Assertions.assertSame;

import java.util.ArrayList;
import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import com.openttd.network.constant.NetworkType.DestType;
import com.openttd.network.constant.NetworkType.NetworkAction;
import com.openttd.network.constant.NetworkType.NetworkErrorCode;
import com.openttd.network.constant.NetworkType.NetworkVehicleType;
import com.openttd.network.constant.TcpAdmin.AdminCompanyRemoveReason;
import com.openttd.network.constant.TcpAdmin.AdminUpdateFrequency;
import com.openttd.network.constant.TcpAdmin.AdminUpdateType;
import com.openttd.network.constant.TcpAdmin.PacketAdminType;
import com.openttd.network.constant.TcpAdmin.PacketServerType;

/**
 * Wire format parity with the OpenTTD binary in use. The expected ids below are
 * copied from the 15.3 tag, not from the working checkout (which is 16.0-dev):
 *
 * - git show 15.3:src/network/core/tcp_admin.h (AdminPacketAdminType,
 *   AdminPacketServerType, AdminUpdateType, AdminUpdateFrequency,
 *   AdminCompanyRemoveReason)
 * - git show 15.3:src/network/network_type.h (NetworkVehicleType, DestType,
 *   NetworkAction, NetworkErrorCode)
 * - git show 15.3:src/network/core/config.h (NETWORK_GAME_ADMIN_VERSION = 3)
 *
 * Java enums derive their wire value from ordinal(): admin packets use the
 * ordinal, server packets the ordinal plus 100. Upstream appends new enum items
 * at the end and keeps retired ids in place (ADMIN_PACKET_SERVER_CMD_LOGGING_OLD
 * is the id ADMIN_PACKET_SERVER_CMD_LOGGING used in admin protocol version 1),
 * so inserting or reordering a constant silently changes the protocol.
 */
public class TcpAdminWireFormatTest {

    private static final int SERVER_PACKET_START = 100;

    @Test
    @DisplayName("ADMIN_PACKET_ADMIN_* ids match tcp_admin.h")
    public void testAdminCommandIds() {
        assertWireIds(PacketAdminType.class, PacketAdminType.values(), new int[]{
                0, // ADMIN_PACKET_ADMIN_JOIN
                1, // ADMIN_PACKET_ADMIN_QUIT
                2, // ADMIN_PACKET_ADMIN_UPDATE_FREQUENCY
                3, // ADMIN_PACKET_ADMIN_POLL
                4, // ADMIN_PACKET_ADMIN_CHAT
                5, // ADMIN_PACKET_ADMIN_RCON
                6, // ADMIN_PACKET_ADMIN_GAMESCRIPT
                7, // ADMIN_PACKET_ADMIN_PING
                8, // ADMIN_PACKET_ADMIN_EXTERNAL_CHAT
                9, // ADMIN_PACKET_ADMIN_JOIN_SECURE
                10, // ADMIN_PACKET_ADMIN_AUTH_RESPONSE
        });
        assertEquals(11, PacketAdminType.values().length);
    }

    @Test
    @DisplayName("ADMIN_PACKET_SERVER_* ids match tcp_admin.h, CMD_LOGGING is 127 not 123")
    public void testServerPacketIds() {
        assertWireIds(PacketServerType.class, serverConstants(), new int[]{
                100, // ADMIN_PACKET_SERVER_FULL
                101, // ADMIN_PACKET_SERVER_BANNED
                102, // ADMIN_PACKET_SERVER_ERROR
                103, // ADMIN_PACKET_SERVER_PROTOCOL
                104, // ADMIN_PACKET_SERVER_WELCOME
                105, // ADMIN_PACKET_SERVER_NEWGAME
                106, // ADMIN_PACKET_SERVER_SHUTDOWN
                107, // ADMIN_PACKET_SERVER_DATE
                108, // ADMIN_PACKET_SERVER_CLIENT_JOIN
                109, // ADMIN_PACKET_SERVER_CLIENT_INFO
                110, // ADMIN_PACKET_SERVER_CLIENT_UPDATE
                111, // ADMIN_PACKET_SERVER_CLIENT_QUIT
                112, // ADMIN_PACKET_SERVER_CLIENT_ERROR
                113, // ADMIN_PACKET_SERVER_COMPANY_NEW
                114, // ADMIN_PACKET_SERVER_COMPANY_INFO
                115, // ADMIN_PACKET_SERVER_COMPANY_UPDATE
                116, // ADMIN_PACKET_SERVER_COMPANY_REMOVE
                117, // ADMIN_PACKET_SERVER_COMPANY_ECONOMY
                118, // ADMIN_PACKET_SERVER_COMPANY_STATS
                119, // ADMIN_PACKET_SERVER_CHAT
                120, // ADMIN_PACKET_SERVER_RCON
                121, // ADMIN_PACKET_SERVER_CONSOLE
                122, // ADMIN_PACKET_SERVER_CMD_NAMES
                123, // ADMIN_PACKET_SERVER_CMD_LOGGING_OLD
                124, // ADMIN_PACKET_SERVER_GAMESCRIPT
                125, // ADMIN_PACKET_SERVER_RCON_END
                126, // ADMIN_PACKET_SERVER_PONG
                127, // ADMIN_PACKET_SERVER_CMD_LOGGING
                128, // ADMIN_PACKET_SERVER_AUTH_REQUEST
                129, // ADMIN_PACKET_SERVER_ENABLE_ENCRYPTION
        });
    }

    @Test
    @DisplayName("every server packet id decodes back to its constant")
    public void testServerPacketRoundTrip() {
        for (PacketServerType type : serverConstants()) {
            assertSame(type, PacketServerType.valueOf(type.ordinal() + SERVER_PACKET_START), type.toString());
        }
        // The sentinel is only a fallback, it must not claim a real packet id.
        assertSame(PacketServerType.INVALID_ADMIN_PACKET, PacketServerType.valueOf(99));
        assertSame(PacketServerType.INVALID_ADMIN_PACKET, PacketServerType.valueOf(130));
        assertSame(PacketServerType.INVALID_ADMIN_PACKET, PacketServerType.valueOf(0));
    }

    @Test
    @DisplayName("admin command ids decode back, unknown ids are null")
    public void testAdminCommandRoundTrip() {
        for (PacketAdminType type : PacketAdminType.values()) {
            assertSame(type, PacketAdminType.valueOf(type.ordinal()), type.toString());
        }
        assertNull(PacketAdminType.valueOf(11));
        assertNull(PacketAdminType.valueOf(100));
    }

    @Test
    @DisplayName("ADMIN_UPDATE_* ids match tcp_admin.h and the sentinel is last")
    public void testUpdateTypes() {
        assertWireIds(AdminUpdateType.class, AdminUpdateType.values(), new int[]{
                0, // ADMIN_UPDATE_DATE
                1, // ADMIN_UPDATE_CLIENT_INFO
                2, // ADMIN_UPDATE_COMPANY_INFO
                3, // ADMIN_UPDATE_COMPANY_ECONOMY
                4, // ADMIN_UPDATE_COMPANY_STATS
                5, // ADMIN_UPDATE_CHAT
                6, // ADMIN_UPDATE_CONSOLE
                7, // ADMIN_UPDATE_CMD_NAMES
                8, // ADMIN_UPDATE_CMD_LOGGING
                9, // ADMIN_UPDATE_GAMESCRIPT
                10, // ADMIN_UPDATE_END
        });
        // The server iterates 0..ADMIN_UPDATE_END, an unknown id maps to the sentinel.
        assertSame(AdminUpdateType.ADMIN_UPDATE_END, AdminUpdateType.valueOf(11));
        assertSame(AdminUpdateType.ADMIN_UPDATE_END, AdminUpdateType.valueOf(-1));
    }

    @Test
    @DisplayName("ADMIN_FREQUENCY_* masks match tcp_admin.h's EnumBitSet")
    public void testUpdateFrequencyMasks() {
        assertEquals(0x01, AdminUpdateFrequency.ADMIN_FREQUENCY_POLL.mask);
        assertEquals(0x02, AdminUpdateFrequency.ADMIN_FREQUENCY_DAILY.mask);
        assertEquals(0x04, AdminUpdateFrequency.ADMIN_FREQUENCY_WEEKLY.mask);
        assertEquals(0x08, AdminUpdateFrequency.ADMIN_FREQUENCY_MONTHLY.mask);
        assertEquals(0x10, AdminUpdateFrequency.ADMIN_FREQUENCY_QUARTERLY.mask);
        assertEquals(0x20, AdminUpdateFrequency.ADMIN_FREQUENCY_ANUALLY.mask);
        assertEquals(0x40, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC.mask);
        assertEquals(7, AdminUpdateFrequency.values().length);
    }

    @Test
    @DisplayName("ADMIN_CRR_* ids match tcp_admin.h")
    public void testCompanyRemoveReasons() {
        assertWireIds(AdminCompanyRemoveReason.class, AdminCompanyRemoveReason.values(), new int[]{0, 1, 2, 3});
        assertSame(AdminCompanyRemoveReason.ADMIN_CRR_MANUAL, AdminCompanyRemoveReason.valueOf(0));
        assertSame(AdminCompanyRemoveReason.ADMIN_CRR_AUTOCLEAN, AdminCompanyRemoveReason.valueOf(1));
        assertSame(AdminCompanyRemoveReason.ADMIN_CRR_BANKRUPT, AdminCompanyRemoveReason.valueOf(2));
        assertSame(AdminCompanyRemoveReason.ADMIN_CRR_END, AdminCompanyRemoveReason.valueOf(4));
    }

    @Test
    @DisplayName("NETWORK_ERROR_* ids match network_type.h, including the auth codes")
    public void testNetworkErrorCodes() {
        assertWireIds(NetworkErrorCode.class, NetworkErrorCode.values(), new int[]{
                0, // NETWORK_ERROR_GENERAL
                1, // NETWORK_ERROR_DESYNC
                2, // NETWORK_ERROR_SAVEGAME_FAILED
                3, // NETWORK_ERROR_CONNECTION_LOST
                4, // NETWORK_ERROR_ILLEGAL_PACKET
                5, // NETWORK_ERROR_NEWGRF_MISMATCH
                6, // NETWORK_ERROR_NOT_AUTHORIZED
                7, // NETWORK_ERROR_NOT_EXPECTED
                8, // NETWORK_ERROR_WRONG_REVISION
                9, // NETWORK_ERROR_NAME_IN_USE
                10, // NETWORK_ERROR_WRONG_PASSWORD
                11, // NETWORK_ERROR_COMPANY_MISMATCH
                12, // NETWORK_ERROR_KICKED
                13, // NETWORK_ERROR_CHEATER
                14, // NETWORK_ERROR_FULL
                15, // NETWORK_ERROR_TOO_MANY_COMMANDS
                16, // NETWORK_ERROR_TIMEOUT_PASSWORD
                17, // NETWORK_ERROR_TIMEOUT_COMPUTER
                18, // NETWORK_ERROR_TIMEOUT_MAP
                19, // NETWORK_ERROR_TIMEOUT_JOIN
                20, // NETWORK_ERROR_INVALID_CLIENT_NAME
                21, // NETWORK_ERROR_NOT_ON_ALLOW_LIST
                22, // NETWORK_ERROR_NO_AUTHENTICATION_METHOD_AVAILABLE
                23, // NETWORK_ERROR_END
        });
        assertSame(NetworkErrorCode.NETWORK_ERROR_WRONG_PASSWORD, NetworkErrorCode.valueOf(10));
        assertSame(NetworkErrorCode.NETWORK_ERROR_NOT_ON_ALLOW_LIST, NetworkErrorCode.valueOf(21));
        assertSame(NetworkErrorCode.NETWORK_ERROR_NO_AUTHENTICATION_METHOD_AVAILABLE, NetworkErrorCode.valueOf(22));
        assertNull(NetworkErrorCode.valueOf(24));
    }

    @Test
    @DisplayName("chat enums match network_type.h (part of the admin network API)")
    public void testChatEnums() {
        assertWireIds(DestType.class, DestType.values(), new int[]{0, 1, 2});
        assertWireIds(NetworkAction.class, NetworkAction.values(), new int[]{
                0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12
        });
        assertEquals(NetworkAction.NETWORK_ACTION_SERVER_MESSAGE.ordinal(), 2);
        assertEquals(NetworkAction.NETWORK_ACTION_CHAT_CLIENT.ordinal(), 5);
        assertEquals(NetworkAction.NETWORK_ACTION_EXTERNAL_CHAT.ordinal(), 12);
        assertEquals(DestType.DESTTYPE_BROADCAST.ordinal(), 0);
        assertEquals(DestType.DESTTYPE_TEAM.ordinal(), 1);
        assertEquals(DestType.DESTTYPE_CLIENT.ordinal(), 2);
    }

    /**
     * COMPANY_STATS sends five vehicle counts then five station counts, in
     * NetworkVehicleType order (network_type.h: 0..NETWORK_VEH_END).
     */
    @Test
    public void testVehicleTypeLength() {
        assertEquals(5, NetworkVehicleType.length());
        assertEquals(6, NetworkVehicleType.values().length);
        assertEquals(NetworkVehicleType.NETWORK_VEH_TRAIN.ordinal(), 0);
        assertEquals(NetworkVehicleType.NETWORK_VEH_LORRY.ordinal(), 1);
        assertEquals(NetworkVehicleType.NETWORK_VEH_BUS.ordinal(), 2);
        assertEquals(NetworkVehicleType.NETWORK_VEH_PLANE.ordinal(), 3);
        assertEquals(NetworkVehicleType.NETWORK_VEH_SHIP.ordinal(), 4);
        assertEquals(NetworkVehicleType.NETWORK_VEH_END.ordinal(), 5);
    }

    @Test
    public void testSubProtocol() {
        for (PacketAdminType type : PacketAdminType.values()) {
            assertEquals(PacketType.SubProtocol.TCP_ADMIN, type.getType());
        }
        for (PacketServerType type : serverConstants()) {
            assertEquals(PacketType.SubProtocol.TCP_ADMIN, type.getType());
        }
    }

    /** Everything except the INVALID_ADMIN_PACKET sentinel. */
    private static PacketServerType[] serverConstants() {
        List<PacketServerType> constants = new ArrayList<>();
        for (PacketServerType type : PacketServerType.values()) {
            if (type != PacketServerType.INVALID_ADMIN_PACKET) {
                constants.add(type);
            }
        }
        return constants.toArray(new PacketServerType[0]);
    }

    private static <E extends Enum<E>> void assertWireIds(Class<E> enumClass, E[] constants, int[] expectedIds) {
        assertEquals(expectedIds.length, constants.length, enumClass.getSimpleName() + " has a different number of constants than 15.3");
        for (int i = 0; i < expectedIds.length; i++) {
            int actualId = enumClass == PacketServerType.class
                    ? constants[i].ordinal() + SERVER_PACKET_START
                    : constants[i].ordinal();
            assertEquals(expectedIds[i], actualId, enumClass.getSimpleName() + "." + constants[i] + " wire id");
        }
    }
}
