/*
 * To change this template, choose Tools | Templates
 * and open the template in the editor.
 */
package com.openttd.constant;

/**
 *
 * @author lubuntu
 */
public class OTTD {
	
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
	 * see news_type.h and script_news.hpp
	 * @author edlefou
	 */
	public enum NewsType {
		/* Arbitrary selection of NewsTypes which might make sense for scripts */
		NT_ARRIVAL_COMPANY, ///< First vehicle arrived for company
		NT_ARRIVAL_OTHER,   ///< First vehicle arrived for competitor
		NT_ACCIDENT,        ///< An accident or disaster has occurred
		NT_COMPANY_INFO,    ///< Company info (new companies, bankruptcy messages)
		NT_INDUSTRY_OPEN,   ///< Opening of industries
		NT_INDUSTRY_CLOSE,  ///< Closing of industries
		NT_ECONOMY,         ///< Economic changes (recession, industry up/dowm)
		NT_INDUSTRY_COMPANY,///< Production changes of industry serviced by local company
		NT_INDUSTRY_OTHER,  ///< Production changes of industry serviced by competitor(s)
		NT_INDUSTRY_NOBODY, ///< Other industry production changes
		NT_ADVICE,          ///< Bits of news about vehicles of the company
		NT_NEW_VEHICLES,    ///< New vehicle has become available
		NT_ACCEPTANCE,      ///< A type of cargo is (no longer) accepted
		NT_SUBSIDIES,       ///< News about subsidies (announcements, expirations, acceptance)
		NT_GENERAL;         ///< General news (from towns)

		/* At the moment (1.3.0), no go only works with general, economy and subsidies.*/
		public boolean isValid() {
			return this == NT_GENERAL || this == NT_ECONOMY || this == NT_SUBSIDIES;
		};
	};
}
