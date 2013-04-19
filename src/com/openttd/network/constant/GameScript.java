package com.openttd.network.constant;

public class GameScript {
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

	/**
	 * Goal types that can be given to a goal.
	 * @author edlefou
	 * see script_goal.hpp
	 */
	public enum GoalType {
		/* Note: these values represent part of the in-game GoalType enum */
		GT_NONE,     ///< Destination is not linked.
		GT_TILE,     ///< Destination is a tile.
		GT_INDUSTRY, ///< Destination is an industry.
		GT_TOWN,     ///< Destination is a town.
		GT_COMPANY,  ///< Destination is a company.
	};

	/**
	 * Enumeration for the news types that a script can create news for.
	 * see script_news.hpp
	 * @author edlefou
	 */
	public enum NewsType {
		/* Arbitrary selection of NewsTypes which might make sense for scripts */
		NT_ACCIDENT,         ///< Category accidents.
		NT_COMPANY_INFO,     ///< Category company info.
		NT_ECONOMY,          ///< Category economy.
		NT_ADVICE,           ///< Category vehicle advice.
		NT_ACCEPTANCE,       ///< Category acceptance changes.
		NT_SUBSIDIES,        ///< Category subsidies.
		NT_GENERAL,          ///< Category general.
	};
}