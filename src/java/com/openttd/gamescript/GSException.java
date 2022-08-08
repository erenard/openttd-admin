package com.openttd.gamescript;

public class GSException extends RuntimeException {

    private static final long serialVersionUID = 5075041520008428366L;

    GSException(String exception) {
        super(exception);
    }
}
