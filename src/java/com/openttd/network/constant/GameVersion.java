package com.openttd.network.constant;

public interface GameVersion {

    static final int MAJOR_VERSION = 12;
    static final int MINOR_VERSION = 2;
    /* See .version */
    static final String OPENTTD = MAJOR_VERSION + "." + MINOR_VERSION;
    /* See rev.cpp.in */
    static final long NEWGRF = (MAJOR_VERSION + 16) << 24 | MINOR_VERSION << 20 | 1 << 19 | 28004;
}
