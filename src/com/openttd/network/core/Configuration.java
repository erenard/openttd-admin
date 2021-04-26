package com.openttd.network.core;

public class Configuration {
	public String host = "localhost";
	public Integer adminPort = 3977;
	public Integer clientPort = 3979;
	public String password = "";
	public String name = "openttd";
	/* See .version */
	private final int version[] = {1, 11, 1};
	public final String openttdVersion = version[0] + "." + version[1] + "." + version[2];
	private final long revision = 28004;
	public final long openttdNewgrfVersion = version[0] << 28 | version[1] << 24 | 0 << 20 | 1 << 19 | revision;
}
