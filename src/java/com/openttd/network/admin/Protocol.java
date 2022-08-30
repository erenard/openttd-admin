package com.openttd.network.admin;

import java.util.Formatter;
import java.util.Locale;
import java.util.Map;

import com.openttd.network.constant.TcpAdmin.AdminUpdateFrequency;
import com.openttd.network.constant.TcpAdmin.AdminUpdateType;
import java.util.EnumMap;

public class Protocol {

    private final short version;

    public Protocol(short version) {
        this.version = version;
    }

    public short getVersion() {
        return version;
    }

    private final Map<AdminUpdateType, Integer> protocols = new EnumMap<>(AdminUpdateType.class);

    public void putProtocol(int key, int value) {
        protocols.put(AdminUpdateType.valueOf(key), value);
    }

    public boolean hasProtocol(AdminUpdateType adminUpdateType, AdminUpdateFrequency adminUpdateFrequency) {
        Integer value = protocols.get(adminUpdateType);
        return value != null && (value & adminUpdateFrequency.mask) != 0;
    }

    @Override
    public String toString() {
        StringBuffer stringBuffer = new StringBuffer();
        stringBuffer.append("Protocol version '").append(version).append("' :\n");
        try ( Formatter formatter = new Formatter(stringBuffer, Locale.getDefault())) {
            for (AdminUpdateType adminUpdateType : AdminUpdateType.values()) {
                for (AdminUpdateFrequency adminUpdateFrequency : AdminUpdateFrequency.values()) {
                    if (hasProtocol(adminUpdateType, adminUpdateFrequency)) {
                        formatter.format("%-30s - %s\n", adminUpdateType, adminUpdateFrequency);
                    }
                }
            }
        }
        return stringBuffer.toString();
    }
}
