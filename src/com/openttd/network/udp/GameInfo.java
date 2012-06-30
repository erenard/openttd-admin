package com.openttd.network.udp;

import java.io.PrintWriter;
import java.io.StringWriter;
import java.text.SimpleDateFormat;
import java.util.Date;
import java.util.HashMap;
import java.util.Map;
import java.util.Map.Entry;

import com.openttd.util.Convert;

public class GameInfo {
	private int version;
	private String serverName;
	private String serverRevision;
	private boolean dedicated;
	private String mapName;
	private short mapSet;
	private long gameStartDate;
	private int mapWidth;
	private int mapHeight;
	private int grfCount;
	private Map<byte[], byte[]> grfs;
	private long gameDate;
	private short companiesMax;
	private short spectatorsMax;
	private short companiesOn;
	private short serverLanguage;
	private boolean usePassword;
	private short clientsMax;
	private short clientsOn;
	private short spectatorsOn;

	public int getVersion() {
		return version;
	}

	public void setVersion(int version) {
		this.version = version;
	}

	public int getGrfCount() {
		return grfCount;
	}

	public void setGrfCount(int grfCount) {
		this.grfCount = grfCount;
	}

	public Map<byte[], byte[]> getGrfs() {
		return grfs;
	}

	public void setGrfs(Map<byte[], byte[]> grfs) {
		this.grfs = grfs;
	}

	public short getCompaniesMax() {
		return companiesMax;
	}

	public void setCompaniesMax(short companiesMax) {
		this.companiesMax = companiesMax;
	}

	public short getSpectatorsMax() {
		return spectatorsMax;
	}

	public void setSpectatorsMax(short spectatorsMax) {
		this.spectatorsMax = spectatorsMax;
	}

	public short getCompaniesOn() {
		return companiesOn;
	}

	public void setCompaniesOn(short companiesOn) {
		this.companiesOn = companiesOn;
	}

	public String getServerName() {
		return serverName;
	}

	public void setServerName(String serverName) {
		this.serverName = serverName;
	}

	public String getServerRevision() {
		return serverRevision;
	}

	public void setServerRevision(String serverRevision) {
		this.serverRevision = serverRevision;
	}

	public short getServerLanguage() {
		return serverLanguage;
	}

	public void setServerLanguage(short serverLanguage) {
		this.serverLanguage = serverLanguage;
	}

	public boolean isUsePassword() {
		return usePassword;
	}

	public void setUsePassword(boolean usePassword) {
		this.usePassword = usePassword;
	}

	public short getClientsMax() {
		return clientsMax;
	}

	public void setClientsMax(short clientsMax) {
		this.clientsMax = clientsMax;
	}

	public short getClientsOn() {
		return clientsOn;
	}

	public void setClientsOn(short clientsOn) {
		this.clientsOn = clientsOn;
	}

	public short getSpectatorsOn() {
		return spectatorsOn;
	}

	public void setSpectatorsOn(short spectatorsOn) {
		this.spectatorsOn = spectatorsOn;
	}

	public String getMapName() {
		return mapName;
	}

	public void setMapName(String mapName) {
		this.mapName = mapName;
	}

	public int getMapWidth() {
		return mapWidth;
	}

	public void setMapWidth(int mapWidth) {
		this.mapWidth = mapWidth;
	}

	public int getMapHeight() {
		return mapHeight;
	}

	public void setMapHeight(int mapHeight) {
		this.mapHeight = mapHeight;
	}

	public short getMapSet() {
		return mapSet;
	}

	public void setMapSet(short mapSet) {
		this.mapSet = mapSet;
	}

	public boolean isDedicated() {
		return dedicated;
	}

	public void setDedicated(boolean dedicated) {
		this.dedicated = dedicated;
	}

	public long getGameStartDate() {
		return gameStartDate;
	}

	public void setGameStartDate(long gameStartDate) {
		this.gameStartDate = gameStartDate;
	}

	public long getGameDate() {
		return gameDate;
	}

	public void setGameDate(long gameDate) {
		this.gameDate = gameDate;
	}

	@Override
	public String toString() {
		StringWriter writer = new StringWriter();
		PrintWriter pw = new PrintWriter(writer);
		pw.println("GameInfo v" + version + ".");
		pw.println("Server:" + serverName + ", revision: " + serverRevision + ", language: " + serverLanguage + ", dedicated: " + dedicated
				+ ", passworded: " + usePassword + ".");
		pw.println("Map: " + mapName + ", Set: " + mapSet + ", " + mapWidth + "x" + mapHeight + ", " + grfCount + "grf(s).");
		Date gameDate = Convert.dayToCalendar(this.gameDate).getTime();
		Date gameStartDate = Convert.dayToCalendar(this.gameStartDate).getTime();
		SimpleDateFormat sdf = new SimpleDateFormat("dd/MM/yyyy");
		pw.println("Date: " + sdf.format(gameDate) + ", started: " + sdf.format(gameStartDate) + ".");
		for (Entry<Integer, CompanyInfo> entry : companyByNumber.entrySet()) {
			pw.println(entry.getValue());
		}
		pw.flush();
		return writer.getBuffer().toString();
	}

	Map<Integer, CompanyInfo> companyByNumber = new HashMap<Integer, CompanyInfo>();

	public void putCompany(int number, CompanyInfo company) {
		companyByNumber.put(number, company);
	}

	public CompanyInfo getCompany(int number) {
		return companyByNumber.get(number);
	}
}
