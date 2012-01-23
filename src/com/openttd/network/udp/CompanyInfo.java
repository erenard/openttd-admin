package com.openttd.network.udp;

public class CompanyInfo {

	private final short id;

	public CompanyInfo(short id) {
		this.id = id;
	}

	private String name;
	private long inauguratedYear;
	private long value;
	private long money;
	private long income;
	private int performance;
	private boolean usePassword;
	private int[] vehicules = new int[5];
	private int[] stations = new int[5];
	private boolean ai;

	public String getName() {
		return name;
	}

	public void setName(String name) {
		this.name = name;
	}

	public boolean isUsePassword() {
		return usePassword;
	}

	public void setUsePassword(boolean usePassword) {
		this.usePassword = usePassword;
	}

	public long getInauguratedYear() {
		return inauguratedYear;
	}

	public void setInauguratedYear(long inauguratedYear) {
		this.inauguratedYear = inauguratedYear;
	}

	public boolean isAi() {
		return ai;
	}

	public void setAi(boolean ai) {
		this.ai = ai;
	}

	public long getValue() {
		return value;
	}

	public void setValue(long value) {
		this.value = value;
	}

	public long getMoney() {
		return money;
	}

	public void setMoney(long money) {
		this.money = money;
	}

	public long getIncome() {
		return income;
	}

	public void setIncome(long income) {
		this.income = income;
	}

	public int getPerformance() {
		return performance;
	}

	public void setPerformance(int performance) {
		this.performance = performance;
	}

	public int[] getVehicules() {
		return vehicules;
	}

	public void setVehicules(int[] vehicules) {
		this.vehicules = vehicules;
	}

	public int[] getStations() {
		return stations;
	}

	public void setStations(int[] stations) {
		this.stations = stations;
	}

	public short getId() {
		return id;
	}

	@Override
	public String toString() {
		return id + ":" + name + ", " + inauguratedYear + ", password:" + usePassword + "\t" + money + "\t" + income + "\t" + value + "\t"
				+ performance;
	}

}
