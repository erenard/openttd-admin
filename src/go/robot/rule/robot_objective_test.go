// Tests for the two objectives of the robot layer: TimerObjective and
// CompanyEconomyObjective.
//
// There is no Java test for either class - the Java tree has no tests at all for
// com.openttd.robot - so this file is new coverage, not a port. It was written
// against robot/rule/timer_objective.go and robot/rule/company_economy_objective.go
// as the specification, and pins the strings those rules send, the order the
// leaderboard is sorted in, and the moments the clock decides to act.
//
// The helpers it uses (NewAdminHarness, SendChat, SendDate, AwaitChat, ...) live
// in robot_rule_test.go, the other half of the same test package.

package rule_test

import "strings"
import "testing"
import "com.openttd"
import "com.openttd/network/constant"
import rule "com.openttd/robot/rule"
import "com.openttd/server"

// GameScriptJsons is the gamescript payloads the admin has sent, oldest first,
// which is how the winner's news paper can be seen.
func GameScriptJsons(harness *server.AdminHarness) []string {
	var jsons = []string{}
	for _, p := range harness.GetServer().Received() {
		if p.GetWireTypeId() == int(constant.ADMIN_PACKET_ADMIN_GAMESCRIPT) {
			jsons = append(jsons, p.GetPacket().ReadString())
		}
	}
	return jsons
}

// ---------------------------------------------------------------------------
// TimerObjective
// ---------------------------------------------------------------------------

// TestGoalStatesTheDeadline is Rule #1 of TimerObjective: "!goal" before the
// clock has started quotes the number of years, not a date.
func TestGoalStatesTheDeadline(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	rule.NewTimerObjective(harness.GetAdmin(), rule.NewExternalUsers(harness.GetAdmin()), 42, 2)
	SendChat(harness, "!goal")
	AwaitChat(harness, 3)
	var messages = ChatMessages(harness)
	if messages[0] != "Goal ***" {
		t.Errorf("goal header = %q", messages[0])
	}
	if messages[1] != "Reach the best company performance in 2years to win the game." {
		t.Errorf("goal deadline = %q", messages[1])
	}
	if messages[2] != "Your score will be your company's performance." {
		t.Errorf("goal score line = %q", messages[2])
	}
	if CountRcon(harness) != 0 {
		t.Errorf("goal sent rcon: %s", strings.Join(RconCommands(harness), " | "))
	}
}

// TestScoreListsTheCompanies is Rule #2 of TimerObjective: "!score" answers one
// line per company, best first, with the change since the previous week.
func TestScoreListsTheCompanies(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = NewExternalUsersFixture(harness, NewUser(11, "Dave", false))
	if fixture == nil {
		return
	}
	rule.NewTimerObjective(harness.GetAdmin(), fixture.users, 42, 2)
	SendClientInfo(harness, CLIENT_ID, "1.2.3.4.5.6.7.8", "Alice", COMPANY_ID)
	SendCompanyInfo(harness, COMPANY_ID, "First", true)
	SendCompanyInfo(harness, 8, "Second", true)
	SendCompanyEconomy(harness, COMPANY_ID, 500, 400)
	SendCompanyEconomy(harness, 8, 300, 300)
	fixture.users.SetOwnerOf(COMPANY_ID, NewUser(11, "Dave", false))
	var chat = CountChat(harness)
	SendChat(harness, "!score")
	AwaitChat(harness, chat+3)
	var messages = ChatMessages(harness)
	if messages[chat] != "Score ***" {
		t.Errorf("score header = %q", messages[chat])
	}
	if messages[chat+1] != "500 (+100) - First - Dave." {
		t.Errorf("score row = %q", messages[chat+1])
	}
	// A company nobody logged in for is still on the board, owned by Nobody.
	if messages[chat+2] != "300 (+0) - Second - Nobody." {
		t.Errorf("score row of an unlogged company = %q", messages[chat+2])
	}
}

// TestEndStopsTheGame is Rule #3 of TimerObjective plus the whole endgame: an
// admin types "$end", the next date closes the game, and the winner is told in
// chat, in the news paper, and in the leaderboard the application is given.
func TestEndStopsTheGame(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var service = InjectGameService(42)
	var fixture = NewExternalUsersFixture(harness, NewUser(12, "Dave", true))
	if fixture == nil {
		return
	}
	rule.NewTimerObjective(harness.GetAdmin(), fixture.users, 42, 2)
	SendClientInfo(harness, CLIENT_ID, "1.2.3.4.5.6.7.8", "Alice", COMPANY_ID)
	SendCompanyInfo(harness, COMPANY_ID, "First", true)
	SendCompanyEconomy(harness, COMPANY_ID, 500, 400)
	SendChat(harness, "!login 5d3f9a")
	AwaitChat(harness, 2)
	fixture.users.SetOwnerOf(COMPANY_ID, NewUser(12, "Dave", true))
	// The clock starts: the deadline is two years away, so the game is only
	// reported, not ended. 8 January 1951 is the fourth trimester of its game
	// year, and the rule's own trimester counter starts at 0, so the first
	// date packet always reports the progression.
	SendDate(harness, DAY_1951)
	AwaitChat(harness, 4)
	if !HasChat(harness, "Score ***") {
		t.Errorf("the first trimester reported nothing: %s",
			strings.Join(ChatMessages(harness), " | "))
	}
	// The admin ends the game.
	var chat = CountChat(harness)
	SendChat(harness, "$end")
	AwaitChat(harness, chat+1)
	if ChatMessages(harness)[chat] != "Game ending now !" {
		t.Errorf("end line = %q", ChatMessages(harness)[chat])
	}
	// The next week is past the deadline the "$end" command installed - the
	// date the game model carried when the admin typed it - so the game closes.
	var rcon = CountRcon(harness)
	SendDate(harness, DAY_1951+7)
	AwaitChat(harness, chat+4)
	var messages = ChatMessages(harness)
	if messages[chat+1] != "Game Over ***" {
		t.Errorf("game over line = %q", messages[chat+1])
	}
	if messages[chat+2] != "Dave wins the game and get 500 points on the leaderboard." {
		t.Errorf("winner line = %q", messages[chat+2])
	}
	if !HasRcon(harness, "setting min_active_clients 0") {
		t.Errorf("game over rcon = %s", strings.Join(RconCommands(harness)[rcon:], " | "))
	}
	if !strings.HasPrefix(messages[chat+3], "Server will restart the ") {
		t.Errorf("restart line = %q", messages[chat+3])
	}
	// The winner also gets the news paper.
	var jsons = GameScriptJsons(harness)
	if len(jsons) != 1 {
		t.Errorf("news papers sent: %d", len(jsons))
	} else if !strings.Contains(jsons[0], "Dave wins the game") {
		t.Errorf("news paper = %s", jsons[0])
	}
	// And the application gets the leaderboard: the logged player, the score as
	// the accomplishment, and the winner flag on the first row.
	if !service.saveGameCalled {
		t.Errorf("the game was never saved")
	}
	if len(service.saved) != 1 {
		t.Errorf("leaderboard rows: %d", len(service.saved))
	} else if service.saved[0].ToString() != "12:Dave accomplishment:500.0 duration:1.0 score:0 winner:true" {
		// Java leaves the score field of the GamePlayer alone here: the score
		// travels as the accomplishment, and the score stays the 0 it was born
		// with.
		t.Errorf("leaderboard row = %s", service.saved[0].ToString())
	}
	// A month after the date the game ended the server is told to end the game.
	// broadcastVictory sets the restart date to "one month from now", so a date
	// inside that month would still be a victory display rather than a restart.
	SendDate(harness, DAY_1951+60)
	openttd.Await(func() bool { return service.endGameCalled }, 10,
		"the application to be told the game ended")
}

// ---------------------------------------------------------------------------
// CompanyEconomyObjective
// ---------------------------------------------------------------------------

// TestGoalStatesTheObjective is Rule #1 of CompanyEconomyObjective: "!goal"
// quotes the figure a company has to reach. The objective is a performance, and
// objectiveValue is the double 1000 Java declares, so the line reads 1000.0.
func TestGoalStatesTheObjective(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	rule.NewCompanyEconomyObjective(harness.GetAdmin(), rule.NewExternalUsers(harness.GetAdmin()), 42)
	SendChat(harness, "!goal")
	AwaitChat(harness, 2)
	var messages = ChatMessages(harness)
	if messages[0] != "Goal ***" {
		t.Errorf("goal header = %q", messages[0])
	}
	if messages[1] != "The first company to reach a performance of 1000.0 win the game." {
		t.Errorf("goal objective = %q", messages[1])
	}
	if CountRcon(harness) != 0 {
		t.Errorf("goal sent rcon: %s", strings.Join(RconCommands(harness), " | "))
	}
}

// TestEconomyScoreListsTheCompanies is Rule #2 of CompanyEconomyObjective: "!score"
// answers one line per company, most advanced first, with the progress towards
// the objective as a percentage and the change since the start of the trimester.
func TestEconomyScoreListsTheCompanies(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = NewExternalUsersFixture(harness, NewUser(11, "Dave", false))
	if fixture == nil {
		return
	}
	rule.NewCompanyEconomyObjective(harness.GetAdmin(), fixture.users, 42)
	SendClientInfo(harness, CLIENT_ID, "1.2.3.4.5.6.7.8", "Alice", COMPANY_ID)
	SendCompanyInfo(harness, COMPANY_ID, "First", true)
	SendCompanyInfo(harness, 8, "Second", true)
	// Half way to the objective of 1000, and a fifth of the way: the progress
	// is the performance of the COMPANY_ECONOMY packet divided by it.
	SendCompanyEconomy(harness, COMPANY_ID, 500, 400)
	SendCompanyEconomy(harness, 8, 200, 0)
	SendChat(harness, "!login 5d3f9a")
	AwaitChat(harness, 2)
	var chat = CountChat(harness)
	SendChat(harness, "!score")
	AwaitChat(harness, chat+3)
	var messages = ChatMessages(harness)
	if messages[chat] != "Score ***" {
		t.Errorf("score header = %q", messages[chat])
	}
	if messages[chat+1] != "#1 - 500 - 50% (+50%) - First(Dave)." {
		t.Errorf("score row = %q", messages[chat+1])
	}
	// A company nobody logged in for is still on the board: the owner goes as an
	// empty pair of brackets, where Java would have thrown.
	if messages[chat+2] != "#2 - 200 - 20% (+20%) - Second()." {
		t.Errorf("score row of an unowned company = %q", messages[chat+2])
	}
}

// TestProgressReportedOncePerTrimester covers the score the rule broadcasts by
// itself: the progression of the game, once when the clock enters a new trimester,
// and the change of the next report measured against that snapshot.
func TestProgressReportedOncePerTrimester(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var fixture = NewExternalUsersFixture(harness, NewUser(11, "Dave", false))
	if fixture == nil {
		return
	}
	rule.NewCompanyEconomyObjective(harness.GetAdmin(), fixture.users, 42)
	SendCompanyInfo(harness, COMPANY_ID, "First", true)
	SendCompanyEconomy(harness, COMPANY_ID, 500, 0)
	// The first date packet is a new week and a new trimester, so the progression
	// goes out as a broadcast: the -1 the rule passes showScore addresses the
	// chat to everyone rather than to one client.
	SendDate(harness, DAY_1951)
	AwaitChat(harness, 2)
	if !HasChat(harness, "#1 - 500 - 50% (+50%) - First().") {
		t.Errorf("progression = %s", strings.Join(ChatMessages(harness), " | "))
		return
	}
	// A week later, still in the same trimester: nothing is reported.
	SendDate(harness, DAY_1951+7)
	Settle()
	if CountChat(harness) != 2 {
		t.Errorf("progression repeated: %s", strings.Join(ChatMessages(harness), " | "))
		return
	}
	// The next trimester reports again. The progress has not moved, so the change
	// is measured against the snapshot taken at the start of the trimester.
	SendDate(harness, DAY_APRIL)
	AwaitChat(harness, 4)
	if !HasChat(harness, "#1 - 500 - 50% (+0%) - First().") {
		t.Errorf("second progression = %s", strings.Join(ChatMessages(harness), " | "))
	}
}

// TestVictoryEndsTheGame covers the life cycle of a company that reaches the
// objective: the game is saved, the victory is broadcast for four weeks, and the
// application is then told to end the game.
func TestVictoryEndsTheGame(t *testing.T) {
	var harness = NewAdminHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var service = InjectGameService(42)
	var fixture = NewExternalUsersFixture(harness, NewUser(11, "Dave", false))
	if fixture == nil {
		return
	}
	rule.NewCompanyEconomyObjective(harness.GetAdmin(), fixture.users, 42)
	SendClientInfo(harness, CLIENT_ID, "1.2.3.4.5.6.7.8", "Alice", COMPANY_ID)
	SendCompanyInfo(harness, COMPANY_ID, "First", true)
	// A performance of 1000 is the objective itself, so the progress is 1.0.
	SendCompanyEconomy(harness, COMPANY_ID, 1000, 0)
	SendChat(harness, "!login 5d3f9a")
	AwaitChat(harness, 2)
	var chat = CountChat(harness)
	// 8 January 1951 is ISO week 1 of a year one longer than the inauguration
	// year of the company, so the player's duration is one year: the victory
	// is announced, and the first of the four victory weeks is displayed.
	SendDate(harness, DAY_1951)
	AwaitChat(harness, chat+4)
	var messages = ChatMessages(harness)
	if messages[chat] != "Game Over : First win !!!" {
		t.Errorf("victory line = %q", messages[chat])
	}
	if messages[chat+1] != "Score ***" {
		t.Errorf("victory score header = %q", messages[chat+1])
	}
	if messages[chat+2] != "#1 - 1000 - 100% (+100%) - First Dave score 100 points on the leaderboard." {
		t.Errorf("victory score row = %q", messages[chat+2])
	}
	if messages[chat+3] != "Server will restart." {
		t.Errorf("restart line = %q", messages[chat+3])
	}
	// The three remaining victory weeks repeat the announcement. The dates are
	// awaited one at a time: the game model holds a single date, so sending them
	// together would let the last one overwrite the clock before the rule sees
	// the earlier ones, and the rule would see one new week instead of four.
	SendDate(harness, DAY_1951+7)
	AwaitChat(harness, chat+8)
	SendDate(harness, DAY_1951+14)
	AwaitChat(harness, chat+12)
	SendDate(harness, DAY_1951+21)
	AwaitChat(harness, chat+16)
	var announced = 0
	for _, message := range ChatMessages(harness) {
		if message == "Game Over : First win !!!" {
			announced += 1
		}
	}
	if announced != 4 {
		t.Errorf("victory announced %d times", announced)
	}
	// The fifth week is out of display weeks, so the game is ended instead.
	SendDate(harness, DAY_1951+28)
	openttd.Await(func() bool { return service.endGameCalled }, 10,
		"the application to be told the game ended")
	if !service.saveGameCalled {
		t.Errorf("the game was never saved")
	}
	if len(service.saved) != 1 {
		t.Errorf("leaderboard rows: %d", len(service.saved))
	} else if service.saved[0].ToString() != "11:Dave accomplishment:1.0 duration:1.0 score:0 winner:true" {
		// The accomplishment and the duration are scaled to 0.0 ... 1.0, and the
		// score field of the GamePlayer stays the 0 it was born with: the points
		// travel in the map the leaderboard line above reads.
		t.Errorf("leaderboard row = %s", service.saved[0].ToString())
	}
}
