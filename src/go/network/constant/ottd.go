// OTTD mirrors com.openttd.constant.OTTD: the two in-game enums the goal and
// town commands pass to the AdminCmd gamescript as plain ordinals.
//
// Java keeps this in its own package, com.openttd.constant. The Go port has no
// import aliasing, and NetworkAdminSender needs both these enums and the ones
// in com.openttd.network.constant, so they live in the same constant package.

package constant

// GoalType: what a goal's destination points at. See script_goal.hpp.
type GoalType int

const (
	GOAL_NONE       GoalType = 0 // Destination is not linked.
	GOAL_TILE       GoalType = 1 // Destination is a tile.
	GOAL_INDUSTRY   GoalType = 2 // Destination is an industry.
	GOAL_TOWN       GoalType = 3 // Destination is a town.
	GOAL_COMPANY    GoalType = 4 // Destination is a company.
	GOAL_STORY_PAGE GoalType = 5 // Destination is a story page.
)

const GOAL_COUNT = 6

// GoalTypeName is the enum's name, which is what the Java logs print. AdminCmd
// itself only ever sees the ordinal.
func GoalTypeName(goalType GoalType) string {
	if goalType < 0 || goalType >= GOAL_COUNT {
		return "GOAL_UNKNOWN"
	}
	return goalTypeNames[int(goalType)]
}

var goalTypeNames = []string{
	"GT_NONE",
	"GT_TILE",
	"GT_INDUSTRY",
	"GT_TOWN",
	"GT_COMPANY",
	"GT_STORY_PAGE",
}

// GoalTypeOf is GoalType.valueOf(int).
func GoalTypeOf(order int) GoalType {
	if order < 0 || order >= GOAL_COUNT {
		return GOAL_UNKNOWN
	}
	return GoalType(order)
}

const GOAL_UNKNOWN GoalType = -1

// GoalTypeValues is GoalType.values(), in wire order.
func GoalTypeValues() []GoalType {
	var values []GoalType
	for i := 0; i < GOAL_COUNT; i++ {
		values = append(values, GoalType(i))
	}
	return values
}

// TownEffect: how a cargo behaves for a town, used by setTownCargoGoal. See
// script_cargo's TownEffect in cargotype.h.
type TownEffect int

const (
	TOWN_EFFECT_NONE       TownEffect = 0 // Cargo has no effect.
	TOWN_EFFECT_PASSENGERS TownEffect = 1 // Cargo behaves passenger-like.
	TOWN_EFFECT_MAIL       TownEffect = 2 // Cargo behaves mail-like.
	TOWN_EFFECT_GOODS      TownEffect = 3 // Cargo behaves goods/candy-like.
	TOWN_EFFECT_WATER      TownEffect = 4 // Cargo behaves water-like.
	TOWN_EFFECT_FOOD       TownEffect = 5 // Cargo behaves food/fizzy-drinks-like.
)

const TOWN_EFFECT_COUNT = 6

// TownEffectName is the enum's name, for the logs.
func TownEffectName(effect TownEffect) string {
	if effect < 0 || effect >= TOWN_EFFECT_COUNT {
		return "TOWN_EFFECT_UNKNOWN"
	}
	return townEffectNames[int(effect)]
}

var townEffectNames = []string{
	"TE_NONE",
	"TE_PASSENGERS",
	"TE_MAIL",
	"TE_GOODS",
	"TE_WATER",
	"TE_FOOD",
}

// TownEffectOf is TownEffect.valueOf(int).
func TownEffectOf(order int) TownEffect {
	if order < 0 || order >= TOWN_EFFECT_COUNT {
		return TOWN_EFFECT_UNKNOWN
	}
	return TownEffect(order)
}

const TOWN_EFFECT_UNKNOWN TownEffect = -1
