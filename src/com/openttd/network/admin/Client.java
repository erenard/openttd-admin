package com.openttd.network.admin;

public class Client implements Cloneable {

	private final long id;

	public Client(long id) {
		this.id = id;
	}

	private String ip;
	private String name;
	private short language; // Always 0, deprecated
	private long joinDate;
	private short companyId;

	@Override
	public String toString() {
		return id + ":" + name + ", " + ip + ", " + companyId;
	}

	@Override
	protected Object clone() throws CloneNotSupportedException {
		Client clone = new Client(id);
		clone.setCompanyId(companyId);
		clone.setIp(ip);
		clone.setJoinDate(joinDate);
		clone.setLanguage(language);
		clone.setName(name);
		return clone;
	}

	public String getIp() {
		return ip;
	}

	public void setIp(String ip) {
		this.ip = ip;
	}

	public String getName() {
		return name;
	}

	public void setName(String name) {
		this.name = name;
	}

    @Deprecated
	public short getLanguage() {
		return language;
	}

    @Deprecated
	public void setLanguage(short language) {
		this.language = language;
	}

	public long getJoinDate() {
		return joinDate;
	}

	public void setJoinDate(long joinDate) {
		this.joinDate = joinDate;
	}

	public short getCompanyId() {
		return companyId;
	}

	public void setCompanyId(short companyId) {
		this.companyId = companyId;
	}

	public long getId() {
		return id;
	}
}
