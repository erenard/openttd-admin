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
     *
     * @author erenard see script_goal.hpp
     */
    public enum GoalType {
        /* Note: these values represent part of the in-game GoalType enum */
        GT_NONE, ///< Destination is not linked.
        GT_TILE, ///< Destination is a tile.
        GT_INDUSTRY, ///< Destination is an industry.
        GT_TOWN, ///< Destination is a town.
        GT_COMPANY, ///< Destination is a company.
        GT_STORY_PAGE, ///< Destination is a story page.
    };

    /**
     * The effects a cargo can have on a town. see cargotype.h and
     * script_cargo.hpp
     */
    public enum TownEffect {
        TE_NONE, ///< Cargo has no effect.
        TE_PASSENGERS, ///< Cargo behaves passenger-like.
        TE_MAIL, ///< Cargo behaves mail-like.
        TE_GOODS, ///< Cargo behaves goods/candy-like.
        TE_WATER, ///< Cargo behaves water-like.
        TE_FOOD;            ///< Cargo behaves food/fizzy-drinks-like.
    };

}
