// TimerObjective is com.openttd.robot.rule.TimerObjective: the goal of the game
// is the best company performance when a fixed number of game years have passed,
// so this rule owns the clock: when the game ends, what the winner is told, and
// what is written to the leaderboard.
//
// Rules #1..#3 of the class comment are "!goal"/"!g", "!score"/"!cv"/"!cp" and
// the admin-only "$end". It listens for ChatEvent and DateEvent, and hands the
// players it found to the application's ExternalGameService.
//
// There is no Java test for this class, so timer_objective_test.go is new
// coverage written against this source as the specification.

package rule

import "sort"
import "strings"
import "com.openttd/admin"
import "com.openttd/admin/event"
import "com.openttd/admin/model"
import "com.openttd/gamescript"
import "com.openttd/robot"
import robotmodel "com.openttd/robot/model"
import "com.openttd/util"

var timerObjectiveLog = util.GetLogger("TimerObjective")

// ScoreBean is TimerObjective's private ScoreBean: one row of the leaderboard,
// a company, its owner and the two performance figures the score is made of.
type ScoreBean struct {
	score        int
	lastScore    int
	companyId    int
	companyName  string
	externalUser *robotmodel.ExternalUser
}

type TimerObjective struct {
	base          *AbstractRule
	externalUsers *ExternalUsers
	// externalGameService is the field initializer Java runs before the
	// constructor body, so the application has to inject the service first.
	externalGameService robot.ExternalGameService
	scenarioId          int64
	nbYear              int
	// gameEnded is gameEnded: the game is over, the weeks left are victory
	// weeks.
	gameEnded bool
	// endGameDate is the endGame field, not the endGame() method: the date the
	// game must be over by. Java keeps null until the first DateEvent adds nbYear
	// to it, or until an admin types "$end"; the port needs the extra flag because
	// util.Calendar is a value.
	endGameDate          util.Calendar
	hasEndGameDate       bool
	restartServerDate    util.Calendar
	hasRestartServerDate bool
	// finalScores is finalScores: the leaderboard as it was when the game ended.
	// Java's null stands for "never calculated", which the port keeps as nil.
	finalScores      []ScoreBean
	currentWeek      int
	currentTrimester int
}

// NewTimerObjective is TimerObjective(OpenttdAdmin openttdAdmin, ExternalUsers
// externalUsers, long scenarioId, int nbYear) - the signature
// strategyboard's OpenttdAdminHandler calls.
func NewTimerObjective(openttd *admin.OpenttdAdmin, externalUsers *ExternalUsers, scenarioId int64, nbYear int) *TimerObjective {
	var timer = &TimerObjective{
		externalUsers:       externalUsers,
		externalGameService: robot.GetInstance().GetExternalGameService(),
		scenarioId:          scenarioId,
		nbYear:              nbYear,
		gameEnded:           false,
		currentWeek:         0,
		currentTrimester:    0,
		finalScores:         nil}
	timer.base = NewAbstractRule(openttd, timer, []event.EventKind{event.EVENT_CHAT, event.EVENT_DATE})
	return timer
}

// OnDateEvent is onDateEvent(DateEvent): the clock of the game.
func (t *TimerObjective) OnDateEvent(dateEvent *event.DateEvent) {
	if timerObjectiveLog.IsDebugEnabled() {
		timerObjectiveLog.Debug("TimerObjective.onDateEvent " + dateEvent.ToString())
	}
	var now = dateEvent.GetOpenttd().GetDate()
	if !t.hasEndGameDate {
		// The first date seen decides the deadline: nbYear game years from now.
		t.endGameDate = now.Clone().AddYears(t.nbYear)
		t.hasEndGameDate = true
	}
	if t.isNewWeek(now) {
		var game = t.gameOf(dateEvent.GetOpenttd())
		if game == nil {
			return
		}
		if !t.gameEnded {
			// End game condition.
			if now.After(t.endGameDate) {
				t.finalScores = t.calculateScores(game)
				t.saveGame()
				t.restartServerDate = now.Clone().AddMonths(1)
				t.hasRestartServerDate = true
				t.gameEnded = true
			}
		}
		if t.gameEnded {
			if t.hasRestartServerDate && now.After(t.restartServerDate) {
				t.endGame()
			} else {
				t.broadcastVictory(game)
			}
		}
	}
	// Show game progression. Java tests gameEnded after the block above, so the
	// week the game ends in never gets a progress line.
	if !t.gameEnded && t.isNewTrimester(now) {
		var game = t.gameOf(dateEvent.GetOpenttd())
		if game != nil {
			t.showScore(-1, game)
		}
	}
}

// OnChatEvent is onChatEvent(ChatEvent).
func (t *TimerObjective) OnChatEvent(chatEvent *event.ChatEvent) {
	if timerObjectiveLog.IsDebugEnabled() {
		timerObjectiveLog.Debug("TimerObjective.onChatEvent " + chatEvent.ToString())
	}
	var clientId = chatEvent.GetClientId()
	var message = chatEvent.GetMessage()
	if message == "" {
		// Java's `if (message != null)`.
		return
	}
	var command = strings.ToLower(strings.TrimSpace(message))
	if command == "!goal" || command == "!g" {
		// Rule #1.
		t.showGoal(clientId)
	} else if command == "!score" || command == "!cv" || command == "!cp" {
		// Rule #2.
		var game = t.gameOf(chatEvent.GetOpenttd())
		if game != nil {
			t.showScore(clientId, game)
		}
	} else if command == "$end" {
		// Rule #3, admin only.
		var user = t.externalUsers.GetExternalUser(clientId)
		if user != nil && user.IsAdmin() {
			t.base.GetSend().ChatClient(int64(clientId), "Game ending now !")
			// Java assigns the game's own Calendar here, so the deadline is
			// now: the next DateEvent ends the game.
			t.endGameDate = chatEvent.GetOpenttd().GetDate()
			t.hasEndGameDate = true
		}
	}
}

// showGoal is showGoal(int clientId).
func (t *TimerObjective) showGoal(clientId int) {
	var send = t.base.GetSend()
	send.ChatClient(int64(clientId), "Goal ***")
	if t.hasEndGameDate {
		send.ChatClient(int64(clientId), "Reach the best company performance the "+
			util.FormatDateLongUK(t.endGameDate)+" to win the game.")
	} else {
		send.ChatClient(int64(clientId), "Reach the best company performance in "+
			util.FormatInt(int64(t.nbYear))+"years to win the game.")
	}
	send.ChatClient(int64(clientId), "Your score will be your company's performance.")
}

// showScore is showScore(long clientId, Game openttd). Java always addresses the
// client, even for the -1 it passes itself, so a progress report reaches nobody
// but the console log; the port keeps that.
func (t *TimerObjective) showScore(clientId int, game *model.Game) {
	var send = t.base.GetSend()
	send.ChatClient(int64(clientId), "Score ***")
	for _, bean := range t.calculateScores(game) {
		var ownerName = "Nobody"
		if bean.externalUser != nil {
			ownerName = bean.externalUser.GetName()
		}
		send.ChatClient(int64(clientId), util.FormatInt(int64(bean.score))+" ("+
			util.Signed(bean.score-bean.lastScore)+") - "+bean.companyName+" - "+ownerName+".")
	}
}

// broadcastVictory is broadcastVictory(Game openttd): the announcement, one
// leaderboard line per company, and the news paper for the winner.
func (t *TimerObjective) broadcastVictory(game *model.Game) {
	var send = t.base.GetSend()
	send.ChatBroadcast("Game Over ***")
	send.Rcon("setting min_active_clients 0")
	if t.finalScores != nil {
		var winner = true
		for _, bean := range t.finalScores {
			var line = ""
			if bean.externalUser != nil {
				if winner {
					line = bean.externalUser.GetName() + " wins the game and get " +
						util.FormatInt(int64(bean.score)) + " points on the leaderboard."
					var newsPaper = gamescript.NewGSNewsPaper(gamescript.NT_GENERAL, line)
					t.base.GetOpenttd().GetGSExecutor().Send(newsPaper)
					winner = false
				} else {
					line = bean.externalUser.GetName() + " get " +
						util.FormatInt(int64(bean.score)) + " points on the leaderboard."
				}
			} else {
				// DEVIATION FROM JAVA: getCompany(id).getName() throws when the
				// company is gone from the model; the port leaves the name empty.
				var companyName = ""
				var company = game.GetCompany(bean.companyId)
				if company != nil {
					companyName = company.GetName()
				}
				line = companyName + ": " + util.FormatInt(int64(bean.score)) +
					" points wasted, player not logged."
			}
			send.ChatBroadcast(line)
		}
	}
	if t.hasRestartServerDate {
		send.ChatBroadcast("Server will restart the " +
			util.FormatDateLongUK(t.restartServerDate) + " ***")
	} else {
		send.ChatBroadcast("Server will restart ***")
	}
}

// saveGame is saveGame(): the leaderboard rows the application keeps. The first
// player with a known user is the winner.
func (t *TimerObjective) saveGame() []robotmodel.GamePlayer {
	var gamePlayers = []robotmodel.GamePlayer{}
	var winnerSet = false
	for _, bean := range t.finalScores {
		var userText = "null"
		if bean.externalUser != nil {
			userText = bean.externalUser.ToString()
		}
		timerObjectiveLog.Info(util.FormatInt(int64(bean.companyId)) + "(" +
			util.FormatInt(int64(bean.score)) + "): " + userText)
		if bean.externalUser != nil {
			var player = robotmodel.NewGamePlayer()
			player.SetExternalUser(bean.externalUser)
			if !winnerSet {
				player.SetWinner(true)
				winnerSet = true
			}
			player.SetAccomplishment(float64(bean.score))
			player.SetDuration(1)
			player.SetCompanyId(bean.companyId)
			gamePlayers = append(gamePlayers, *player)
		}
	}
	if t.externalGameService == nil {
		// DEVIATION FROM JAVA: an unconfigured service throws NullPointerException
		// here; the port logs a warning and keeps the players for the caller.
		timerObjectiveLog.Warn("ExternalGameService not received yet, game not saved.")
		return gamePlayers
	}
	t.externalGameService.SaveGame(t.scenarioId, gamePlayers)
	return gamePlayers
}

// endGame is endGame(): telling the application the game is over.
func (t *TimerObjective) endGame() {
	if t.externalGameService == nil {
		timerObjectiveLog.Warn("ExternalGameService not received yet, game not ended.")
		return
	}
	t.externalGameService.EndGame(t.scenarioId)
}

// calculateScores is calculateScores(Game openttd): one bean per company, sorted
// by score and then reversed, so the best company comes first.
func (t *TimerObjective) calculateScores(game *model.Game) []ScoreBean {
	var beans = []ScoreBean{}
	for _, company := range game.GetCompanies() {
		if company != nil {
			var bean = ScoreBean{
				score:        company.GetLastPerformance(),
				lastScore:    company.GetPreviousPerformance(),
				companyId:    company.GetId(),
				companyName:  company.GetName(),
				externalUser: t.externalUsers.GetOwnerOf(company.GetId())}
			beans = append(beans, bean)
		}
	}
	return sortScoreBeans(beans)
}

// isNewTrimester is isNewMonth's cousin: MONTH / 3 + 4 * YEAR, the quarter of a
// game year, which is util.Calendar.Trimester().
func (t *TimerObjective) isNewTrimester(now util.Calendar) bool {
	var nowTrimester = now.Trimester()
	if nowTrimester > t.currentTrimester {
		t.currentTrimester = nowTrimester
		return true
	}
	return false
}

// isNewWeek is isNewWeek(Calendar now). Java reads Calendar.WEEK_OF_YEAR, which
// is locale sensitive; the port reads the ISO week of util.Calendar.WeekOfYear(),
// the deviation documented there.
func (t *TimerObjective) isNewWeek(now util.Calendar) bool {
	var nowWeek = now.WeekOfYear()
	if nowWeek != t.currentWeek {
		t.currentWeek = nowWeek
		return true
	}
	return false
}

// gameOf is dateEvent.getOpenttd() / chatEvent.getOpenttd(), down-cast from the
// event's Openttd seam to the admin/model/Game the rules read.
func (t *TimerObjective) gameOf(source event.Openttd) *model.Game {
	var game, ok = source.(*model.Game)
	if ok {
		return game
	}
	return nil
}

// compareScoreBeans is ScoreBean.compareTo(ScoreBean o): score first, then the
// previous score, and a company with a known owner last of the ties, so that
// reversing the list puts it first.
func compareScoreBeans(bean ScoreBean, other ScoreBean) int {
	if bean.score != other.score {
		return bean.score - other.score
	}
	if bean.lastScore != other.lastScore {
		return bean.lastScore - other.lastScore
	}
	if bean.externalUser != nil && other.externalUser == nil {
		return 1
	}
	if bean.externalUser == nil && other.externalUser != nil {
		return -1
	}
	return 0
}

// sortScoreBeans is Collections.sort(scoreBeans) followed by
// Collections.reverse(scoreBeans). Java's sort is stable, so the companies a game
// reports in the same order stay in that order among equal scores; the port uses
// SliceStable to keep that, then reverses the copy.
func sortScoreBeans(beans []ScoreBean) []ScoreBean {
	var sorted = []ScoreBean{}
	for _, bean := range beans {
		sorted = append(sorted, bean)
	}
	sort.SliceStable(sorted, func(i int, j int) bool {
		return compareScoreBeans(sorted[i], sorted[j]) < 0
	})
	var reversed = []ScoreBean{}
	for index := len(sorted) - 1; index >= 0; index -= 1 {
		reversed = append(reversed, sorted[index])
	}
	return reversed
}
