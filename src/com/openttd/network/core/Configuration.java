package com.openttd.network.core;

public class Configuration {
	public String host = "localhost";
	public Integer adminPort = 3977;
	public Integer clientPort = 3979;
	public String password = "";
	public String name = "openttd";
	public String openttdVersion = "1.4.2";
	/* See rev.cpp.in and .ottdrev */
	private long revision = 26740;
	public long openttdNewgrfVersion = 1 << 28 | 4 << 24 | 2 << 20 | 1 << 19 | (revision & ((1 << 19) - 1));
}
