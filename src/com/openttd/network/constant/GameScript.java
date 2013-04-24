package com.openttd.network.constant;

public class GameScript {
	public static final String CMD = "c";
	public static final String ARGS = "a";
	/**
	 * Command list, must be identical in AdminCmd/main.nut.
	 * @author edlefou
	 */
	public enum GSCommand {
		countIndustry,
		createNews,
		addGoal,
		removeGoal,
		removeAllGoal;
	};

}