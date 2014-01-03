package com.openttd.network.constant;

public interface GameScript {
	public static final String CMD = "c";
	public static final String ARGS = "a";
	public static final String ID = "i";
	/**
	 * Command list, must be identical in AdminCmd/main.nut.
	 * @author erenard
	 */
	public enum GSCommand {
		countIndustry,
		createNews,
		addGoal,
		removeGoal,
		removeAllGoal,
		setTownCargoGoal;
	};

}