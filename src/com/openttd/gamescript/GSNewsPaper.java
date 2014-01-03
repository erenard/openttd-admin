package com.openttd.gamescript;

import com.google.gson.JsonArray;
import com.google.gson.JsonObject;
import com.google.gson.JsonPrimitive;
import com.openttd.network.constant.GameScript;

/**
 * Newspaper game script.
 */
public class GSNewsPaper extends GSRequest<Boolean> {

	/**
	 * Enumeration for the news types that a script can create news for.
	 * see news_type.h and script_news.hpp
	 * @author erenard
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
		NT_ADVICE,          ///< Bits of news about vehicles of the companyF
		NT_NEW_VEHICLES,    ///< New vehicle has become available
		NT_ACCEPTANCE,      ///< A type of cargo is (no longer) accepted
		NT_SUBSIDIES,       ///< News about subsidies (announcements, expirations, acceptance)
		NT_GENERAL;         ///< General news (from towns)

		/* At the moment (1.3.0), no go only works with general, economy and subsidies.*/
		public boolean isValid() {
			return this == NT_GENERAL || this == NT_ECONOMY ||    this == NT_SUBSIDIES;
		};
	};

	private short companyId;
	private NewsType newsType;
	private String message;

	/**
	 * Send a newspaper to all companies.
	 * @param newsType	type of news
	 * @param message	content of the news
	 */
	public GSNewsPaper(NewsType newsType, String message) {
		this.companyId = -1;
		this.newsType = newsType;
		this.message = message;
	}

	/**
	 * Send a newspaper to a company.
	 * @param companyId	company id
	 * @param newsType	type of news
	 * @param message	content of the news
	 */
	public GSNewsPaper(short companyId, NewsType newsType, String message) {
		this.companyId = companyId;
		this.newsType = newsType;
		this.message = message;
	}

	@Override
	//TODO Improve this
	protected JsonObject toJson() {
		JsonArray arguments = new JsonArray();
		if(!newsType.isValid()) {
			newsType = NewsType.NT_GENERAL;
		}
		arguments.add(new JsonPrimitive(newsType.ordinal()));
		arguments.add(new JsonPrimitive(message));
		arguments.add(new JsonPrimitive(companyId));
		JsonObject json = new JsonObject();
		json.add(GameScript.ID, new JsonPrimitive(this.getId()));
		json.add(GameScript.CMD, new JsonPrimitive(GameScript.GSCommand.createNews.ordinal()));
		json.add(GameScript.ARGS, arguments);
		return json;
	}
	
	@Override
	protected Boolean fromJson(String json) {
		return Boolean.valueOf(json);
	}

}
