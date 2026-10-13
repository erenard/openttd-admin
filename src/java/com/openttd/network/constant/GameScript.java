package com.openttd.network.constant;

public interface GameScript {

	String CMD = "c";
	String ARGS = "a";
	String ID = "i";

	/**
	 * Command list, must be identical in AdminCmd/main.nut.
	 *
	 * @author erenard
	 */
	enum GSCommand {
		countIndustry,
		createNews,
		addGoal,
		removeGoal,
		removeAllGoal,
		setTownCargoGoal;
	};

}
