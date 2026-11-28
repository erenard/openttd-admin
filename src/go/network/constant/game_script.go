// GameScript mirrors com.openttd.network.constant.GameScript: the three JSON
// field names AdminCmd's gamescript reads, and the command list.
//
// The command order must stay identical to AdminCmd/main.nut, which dispatches
// on the ordinal it finds in "c". Java scopes these names inside the interface
// (GameScript.ID); Go has no such scoping, so they carry a GS_ prefix.

package constant

const (
	GS_ID   = "i" // The request id, so the reply can be correlated.
	GS_CMD  = "c" // The command to run.
	GS_ARGS = "a" // The command's arguments.
)

// GSCommand: the countIndustry..setTownCargoGoal list; the wire value is the
// ordinal, which is also AdminCmd's dispatch index.
type GSCommand int

const (
	GS_COMMAND_COUNT_INDUSTRIE     GSCommand = 0 // How many industries the map has.
	GS_COMMAND_CREATE_NEWS         GSCommand = 1 // Put a message in the player's news ticker.
	GS_COMMAND_ADD_GOAL            GSCommand = 2 // Attach a goal to a company, town or tile.
	GS_COMMAND_REMOVE_GOAL         GSCommand = 3 // Take one goal away again.
	GS_COMMAND_REMOVE_ALL_GOAL     GSCommand = 4 // Clear the whole goal list.
	GS_COMMAND_SET_TOWN_CARGO_GOAL GSCommand = 5 // Set a town's cargo goal.
)

const GS_COMMAND_COUNT = 6

// GSCommandName is the enum's name, which is what the Java logs print. AdminCmd
// itself only ever sees the ordinal.
func GSCommandName(command GSCommand) string {
	if command < 0 || command >= GS_COMMAND_COUNT {
		return "GS_COMMAND_UNKNOWN"
	}
	return gsCommandNames[int(command)]
}

var gsCommandNames = []string{
	"countIndustry",
	"createNews",
	"addGoal",
	"removeGoal",
	"removeAllGoal",
	"setTownCargoGoal",
}

// GSCommandOf is GSCommand.valueOf(int).
func GSCommandOf(order int) GSCommand {
	if order < 0 || order >= GS_COMMAND_COUNT {
		return GS_COMMAND_UNKNOWN
	}
	return GSCommand(order)
}

const GS_COMMAND_UNKNOWN GSCommand = -1

// GSCommandValues is GSCommand.values(), in wire order.
func GSCommandValues() []GSCommand {
	var values []GSCommand
	for i := 0; i < GS_COMMAND_COUNT; i++ {
		values = append(values, GSCommand(i))
	}
	return values
}
