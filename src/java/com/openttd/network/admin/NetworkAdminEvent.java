package com.openttd.network.admin;

import com.openttd.network.constant.NetworkType.NetworkErrorCode;
import com.openttd.network.constant.TcpAdmin.AdminCompanyRemoveReason;
import com.openttd.network.constant.TcpAdmin.PacketServerType;

public class NetworkAdminEvent {

    private final PacketServerType packetServerType;
    private long clientId;
    private long pingId;
    private short companyId;
    private char color;
    private String origin;
    private String message;
    private long date;
    private AdminCompanyRemoveReason companyRemoveReason;
    private NetworkErrorCode clientError;

    public NetworkAdminEvent(PacketServerType packetServerType) {
        this.packetServerType = packetServerType;
    }

    void setClientEvent(PacketServerType packetServerType, long clientId) {
        this.clientId = clientId;
    }

    void setCompanyEvent(PacketServerType packetServerType, short companyId) {
        this.companyId = companyId;
    }

    void setRConEvent(PacketServerType packetServerType, char color, String message) {
        this.color = color;
        this.message = message;
    }

    void setRConEndEvent(PacketServerType packetType, String rcon) {
        this.message = rcon;
    }

    void setDateEvent(PacketServerType packetServerType, long date) {
        this.date = date;
    }

    void setChatEvent(PacketServerType packetServerType, long clientId, String message) {
        this.clientId = clientId;
        this.message = message;
    }

    void setGameScriptEvent(PacketServerType packetServerType, String json) {
        this.message = json;
    }

    void setConsoleEvent(String origin, String message) {
        this.origin = origin;
        this.message = message;
    }

    void setPongEvent(long pingId) {
        this.pingId = pingId;
    }

    void setCompanyRemoveEvent(PacketServerType packetType, short companyId, short reason) {
        this.companyId = companyId;
        this.companyRemoveReason = AdminCompanyRemoveReason.valueOf(reason);
    }

    void setClientErrorEvent(PacketServerType packetType, long clientId, short error) {
        this.clientId = clientId;
        this.clientError = NetworkErrorCode.valueOf(error);
    }

    public long getClientId() {
        return clientId;
    }

    public short getCompanyId() {
        return companyId;
    }

    public long getPingId() {
        return pingId;
    }

    public String getMessage() {
        return message;
    }

    public long getDate() {
        return date;
    }

    public char getColor() {
        return color;
    }

    public String getOrigin() {
        return origin;
    }

    public PacketServerType getPacketServerType() {
        return packetServerType;
    }

    public AdminCompanyRemoveReason getCompanyRemoveReason() {
        return companyRemoveReason;
    }

    public NetworkErrorCode getClientError() {
        return clientError;
    }

    @Override
    public String toString() {
        StringBuilder sb = new StringBuilder(packetServerType.toString());
        sb.append(" clientId:").append(clientId)
                .append(" pingId:").append(pingId)
                .append(" companyId:").append(companyId)
                .append(" color:").append(color)
                .append(" origin:").append(origin)
                .append(" message:").append(message)
                .append(" date:").append(date)
                .append(" companyRemoveReason:").append(companyRemoveReason)
                .append(" clientError:").append(clientError);
        return sb.toString();
    }
}
