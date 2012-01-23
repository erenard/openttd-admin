package com.openttd.admin.model;

import java.util.Calendar;

public class Company {
	private int id;
	private String name;
	private boolean usePassword;
	private long value;
	private int performance;
	private int previousPerformance;
	private Calendar inauguration;

	public int getId() {
		return id;
	}

	public void setId(int id) {
		this.id = id;
	}

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

	public long getValue() {
		return value;
	}

	public void setValue(long value) {
		this.value = value;
	}

	public int getPerformance() {
		return performance;
	}

	public void setPerformance(int performance) {
		this.performance = performance;
	}

	public int getPreviousPerformance() {
		return previousPerformance;
	}

	public void setPreviousPerformance(int previousPerformance) {
		this.previousPerformance = previousPerformance;
	}

	public Calendar getInauguration() {
		return inauguration;
	}

	public void setInauguration(Calendar inauguration) {
		this.inauguration = inauguration;
	}
}
