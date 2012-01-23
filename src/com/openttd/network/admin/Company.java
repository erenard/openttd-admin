package com.openttd.network.admin;

import java.math.BigInteger;

public class Company implements Cloneable {

	private final short id;

	public Company(short id) {
		this.id = id;
	}

	// info + update
	private String name;
	private String president;
	private short color;
	private boolean usePassword;
	private long inauguratedYear;
	private boolean ai;
	// update
	private short bankruptcy;
	private short[] shareOwners = new short[4];
	// economy
	private long money;
	private BigInteger loan;
	private long income;

	private BigInteger lastValue;
	private BigInteger previousValue;

	private int lastPerformance;
	private int previousPerformance;

	private int deliveredCargo;
	private int lastDeliveredCargo;
	private int previousDeliveredCargo;

	// stats
	private char[] vehicules = new char[5];
	private char[] stations = new char[5];

	public char[] getVehicules() {
		return vehicules;
	}

	public void setVehicules(char[] vehicules) {
		this.vehicules = vehicules;
	}

	public char[] getStations() {
		return stations;
	}

	public void setStations(char[] stations) {
		this.stations = stations;
	}

	public short getId() {
		return id;
	}

	public String getName() {
		return name;
	}

	public void setName(String name) {
		this.name = name;
	}

	public String getPresident() {
		return president;
	}

	public void setPresident(String president) {
		this.president = president;
	}

	public short getColor() {
		return color;
	}

	public void setColor(short color) {
		this.color = color;
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

	public short getBankruptcy() {
		return bankruptcy;
	}

	public void setBankruptcy(short bankruptcy) {
		this.bankruptcy = bankruptcy;
	}

	public short[] getShareOwners() {
		return shareOwners;
	}

	public void setShareOwners(short[] shareOwners) {
		this.shareOwners = shareOwners;
	}

	public long getMoney() {
		return money;
	}

	public void setMoney(long money) {
		this.money = money;
	}

	public BigInteger getLoan() {
		return loan;
	}

	public void setLoan(BigInteger loan) {
		this.loan = loan;
	}

	public long getIncome() {
		return income;
	}

	public void setIncome(long income) {
		this.income = income;
	}

	public BigInteger getLastValue() {
		return lastValue;
	}

	public void setLastValue(BigInteger lastValue) {
		this.lastValue = lastValue;
	}

	public BigInteger getPreviousValue() {
		return previousValue;
	}

	public void setPreviousValue(BigInteger previousValue) {
		this.previousValue = previousValue;
	}

	public int getLastPerformance() {
		return lastPerformance;
	}

	public void setLastPerformance(int lastPerformance) {
		this.lastPerformance = lastPerformance;
	}

	public int getPreviousPerformance() {
		return previousPerformance;
	}

	public void setPreviousPerformance(int previousPerformance) {
		this.previousPerformance = previousPerformance;
	}

	public int getDeliveredCargo() {
		return deliveredCargo;
	}

	public void setDeliveredCargo(int deliveredCargo) {
		this.deliveredCargo = deliveredCargo;
	}

	public int getLastDeliveredCargo() {
		return lastDeliveredCargo;
	}

	public void setLastDeliveredCargo(int lastDeliveredCargo) {
		this.lastDeliveredCargo = lastDeliveredCargo;
	}

	public int getPreviousDeliveredCargo() {
		return previousDeliveredCargo;
	}

	public void setPreviousDeliveredCargo(int previousDeliveredCargo) {
		this.previousDeliveredCargo = previousDeliveredCargo;
	}

	@Override
	protected Object clone() throws CloneNotSupportedException {
		Company clone = new Company(id);
		clone.setAi(ai);
		clone.setBankruptcy(bankruptcy);
		clone.setColor(color);
		clone.setDeliveredCargo(deliveredCargo);
		clone.setInauguratedYear(inauguratedYear);
		clone.setIncome(income);
		clone.setLastDeliveredCargo(lastDeliveredCargo);
		clone.setLastPerformance(lastPerformance);
		if (lastValue != null) clone.setLastValue(new BigInteger(lastValue.toString()));
		if (loan != null) clone.setLoan(new BigInteger(loan.toString()));
		clone.setMoney(money);
		clone.setName(name);
		clone.setPresident(president);
		clone.setPreviousDeliveredCargo(previousDeliveredCargo);
		clone.setPreviousPerformance(previousPerformance);
		if (previousValue != null) clone.setPreviousValue(new BigInteger(previousValue.toString()));
		if (shareOwners != null) clone.setShareOwners(shareOwners.clone());
		if (stations != null) clone.setStations(stations.clone());
		clone.setUsePassword(usePassword);
		if (vehicules != null) clone.setVehicules(vehicules.clone());
		return clone;
	}

	@Override
	public String toString() {
		return id + ":" + name + ", " + inauguratedYear + ", password:" + usePassword + "\t" + money + "\t" + income + "\t" + lastValue
				+ "\t" + lastPerformance;
	}

}
