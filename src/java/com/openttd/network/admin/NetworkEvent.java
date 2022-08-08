package com.openttd.network.admin;

import com.openttd.network.constant.TcpAdmin.PacketServerType;

public class NetworkEvent {

    private final PacketServerType packetServerType;
    private long clientId;
    private short companyId;
    private short extraId;
    private char color;
    private String origin;
    private String message;
    private long date;

    public NetworkEvent(PacketServerType packetServerType) {
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

    public long getClientId() {
        return clientId;
    }

    public void setClientId(long clientId) {
        this.clientId = clientId;
    }

    public short getCompanyId() {
        return companyId;
    }

    public short getExtraId() {
        return extraId;
    }

    public void setExtraId(short extraId) {
        this.extraId = extraId;
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
}
