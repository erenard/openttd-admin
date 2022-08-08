package com.openttd.network.admin;

import static com.openttd.network.constant.NetworkType.NetworkErrorCode;

public class NetworkException extends Exception {

    private static final long serialVersionUID = 3163615568947842652L;

    private String message;

    public NetworkException(String message) {
        this.message = message;
    }

    public NetworkException(int networkErrorCodeId) {
        for (NetworkErrorCode code : NetworkErrorCode.values()) {
            if (code.ordinal() == networkErrorCodeId) {
                this.message = code.toString();
            }
        }
    }

    @Override
    public String getMessage() {
        return message;
    }
}
