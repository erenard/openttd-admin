// CompanyEconomyObjective is com.openttd.robot.rule.CompanyEconomyObjective: the
// goal of the game is one company reaching a fixed economic figure, so this rule
// tracks how far every company has got, and ends the game as soon as one of them
// is there.
//
// Rules #1 and #2 of the class comment are "!goal"/"!g" and "!score"/"!cv"/"!cp";
// the third thing it answers to is the admin's "$score", which dumps the
// leaderboard it would send the application. It listens for CompanyEvent, DateEvent
// and ChatEvent.
//
// There is no Java test for this class, so company_economy_objective_test.go is
// new coverage written against this source as the specification.

package rule

import "sort"
import "strings"
import "com.openttd/admin"
import "com.openttd/admin/event"
import "com.openttd/admin/model"
import "com.openttd/robot"
import robotmodel "com.openttd/robot/model"
import "com.openttd/util"

var companyEconomyObjectiveLog = util.GetLogger("CompanyEconomyObjective")

// ObjectiveType is CompanyEconomyObjective's private enum ObjectiveType { VALUE,
// PERFORMANCE }. Java never prints it and never parses it from text, so the port
// gives it no name table.
type ObjectiveType int

const (
	OBJECTIVE_VALUE       ObjectiveType = 0 // A company wins by reaching a value.
	OBJECTIVE_PERFORMANCE ObjectiveType = 1 // A company wins by reaching a performance.
)

// ProgressEntry is the Map.Entry<Short, Double> the rule iterates over: a company
// and how far it has got towards the objective.
type ProgressEntry struct {
	companyId int
	progress  float64
}

type CompanyEconomyObjective struct {
	base          *AbstractRule
	externalUsers *ExternalUsers
	// externalGameService is the field initializer Java runs before the
	// constructor body, so the application has to inject the service first.
	externalGameService robot.ExternalGameService
	// objectiveType is the final field Java initialises to PERFORMANCE; the
	// VALUE branch below is dead code in Java, and the port keeps it so the two
	// objectives stay comparable.
	objectiveType      ObjectiveType
	objectiveValue     float64
	objectiveReached   bool
	currentWeek        int
	currentTrimester   int
	winnerName         string
	victoryDisplayWeek int
	scenarioId         int64
	// scoreByCompanyId, progressByCompanyId and progressHistoryByCompanyId are
	// the three maps of the class: the leaderboard points, the current progress
	// of every company, and the progress it had at the start of the trimester.
	scoreByCompanyId           map[int]int
	progressByCompanyId        map[int]float64
	progressHistoryByCompanyId map[int]float64
}

// NewCompanyEconomyObjective is CompanyEconomyObjective(OpenttdAdmin openttdAdmin,
// ExternalUsers externalUsers, long scenarioId).
func NewCompanyEconomyObjective(openttd *admin.OpenttdAdmin, externalUsers *ExternalUsers, scenarioId int64) *CompanyEconomyObjective {
	var economy = &CompanyEconomyObjective{
		externalUsers:              externalUsers,
		externalGameService:        robot.GetInstance().GetExternalGameService(),
		objectiveType:              OBJECTIVE_PERFORMANCE,
		objectiveValue:             1000,
		objectiveReached:           false,
		currentWeek:                0,
		currentTrimester:           0,
		scenarioId:                 scenarioId,
		scoreByCompanyId:           map[int]int{},
		progressByCompanyId:        map[int]float64{},
		progressHistoryByCompanyId: map[int]float64{}}
	economy.base = NewAbstractRule(openttd, economy,
		[]event.EventKind{event.EVENT_COMPANY, event.EVENT_DATE, event.EVENT_CHAT})
	return economy
}

// OnDateEvent is onDateEvent(DateEvent): test for victory, then show the
// progression.
func (e *CompanyEconomyObjective) OnDateEvent(dateEvent *event.DateEvent) {
	if companyEconomyObjectiveLog.IsDebugEnabled() {
		companyEconomyObjectiveLog.Debug("CompanyEconomyObjective.onDateEvent " + dateEvent.ToString())
	}
	var now = dateEvent.GetOpenttd().GetDate()
	if e.isNewWeek(now) {
		var game = e.gameOf(dateEvent.GetOpenttd())
		if game == nil {
			return
		}
		if !e.objectiveReached {
			// Test victory. Java does not break out of this loop, so a second
			// company that also reached the objective overwrites the winner and
			// saves the game again; the port keeps that.
			for _, entry := range e.getSortedProgressionByCompanyId() {
				if entry.progress >= 1 {
					var company = game.GetCompany(entry.companyId)
					if company != nil {
						// DEVIATION FROM JAVA: getCompany(id) is dereferenced
						// without a check there; the port keeps the previous
						// winner when the company is gone from the model.
						e.winnerName = company.GetName()
					}
					e.victoryDisplayWeek = 4
					e.saveGame(game)
					e.objectiveReached = true
				}
			}
		}
		if e.objectiveReached {
			if e.victoryDisplayWeek > 0 {
				// Display game over and victory.
				e.broadcastVictory(game)
				e.victoryDisplayWeek -= 1
			} else {
				e.endGame()
			}
		}
	}
	// Show game progression, and start the next trimester's history again.
	if !e.objectiveReached && e.isNewTrimester(now) {
		var game = e.gameOf(dateEvent.GetOpenttd())
		if game != nil {
			e.showScore(-1, game)
		}
		e.progressHistoryByCompanyId = map[int]float64{}
		for companyId, progress := range e.progressByCompanyId {
			e.progressHistoryByCompanyId[companyId] = progress
		}
	}
}

// OnCompanyEvent is onCompanyEvent(CompanyEvent): keep the progress of a company
// up to date, and forget a company that was deleted.
func (e *CompanyEconomyObjective) OnCompanyEvent(companyEvent *event.CompanyEvent) {
	if companyEconomyObjectiveLog.IsDebugEnabled() {
		companyEconomyObjectiveLog.Debug("CompanyEconomyObjective.onCompanyEvent " + companyEvent.ToString())
	}
	if e.objectiveReached {
		return
	}
	var companyId = companyEvent.GetCompanyId()
	var action = companyEvent.GetAction()
	if action == event.ACTION_UPDATE {
		// Update scores.
		var game = e.gameOf(companyEvent.GetOpenttd())
		if game == nil {
			return
		}
		var company = game.GetCompany(companyId)
		if company == nil {
			// DEVIATION FROM JAVA: the company is dereferenced without a check
			// there, so an update that arrives before the company does throws.
			return
		}
		var progression = 0.0
		if e.objectiveType == OBJECTIVE_PERFORMANCE {
			progression = float64(company.GetLastPerformance()) / e.objectiveValue
		} else {
			progression = float64(company.GetLastValue()) / e.objectiveValue
		}
		e.progressByCompanyId[companyId] = progression
	} else if action == event.ACTION_DELETE {
		// Java's DELETE case has no break, but its default case is empty, so
		// nothing else happens.
		delete(e.scoreByCompanyId, companyId)
		delete(e.progressByCompanyId, companyId)
		delete(e.progressHistoryByCompanyId, companyId)
	}
}

// OnChatEvent is onChatEvent(ChatEvent).
func (e *CompanyEconomyObjective) OnChatEvent(chatEvent *event.ChatEvent) {
	if companyEconomyObjectiveLog.IsDebugEnabled() {
		companyEconomyObjectiveLog.Debug("CompanyEconomyObjective.onChatEvent " + chatEvent.ToString())
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
		e.showGoal(clientId)
	} else if command == "!score" || command == "!cv" || command == "!cp" {
		// Rule #2.
		var game = e.gameOf(chatEvent.GetOpenttd())
		if game != nil {
			e.showScore(clientId, game)
		}
	} else if command == "$score" {
		// The admin's leaderboard dump: the GamePlayer rows, as the application
		// would receive them.
		var game = e.gameOf(chatEvent.GetOpenttd())
		if game == nil {
			return
		}
		var send = e.base.GetSend()
		for _, player := range e.calculateScores(game) {
			send.ChatClient(int64(clientId), player.ToString())
		}
	}
}

// showGoal is showGoal(int clientId).
func (e *CompanyEconomyObjective) showGoal(clientId int) {
	var send = e.base.GetSend()
	send.ChatClient(int64(clientId), "Goal ***")
	if e.objectiveType == OBJECTIVE_PERFORMANCE {
		send.ChatClient(int64(clientId), "The first company to reach a performance of "+
			util.FormatDouble(e.objectiveValue)+" win the game.")
	} else {
		send.ChatClient(int64(clientId), "The first company to reach a value of "+
			util.FormatDouble(e.objectiveValue*2)+" EUR win the game.")
	}
}

// showScore is showScore(long clientId, Game openttd): one line per company,
// addressed to the client, or broadcast when clientId is the -1 the rule passes
// itself.
func (e *CompanyEconomyObjective) showScore(clientId int, game *model.Game) {
	var send = e.base.GetSend()
	send.ChatClient(int64(clientId), "Score ***")
	var idx = 0
	for _, entry := range e.getSortedProgressionByCompanyId() {
		idx += 1
		var companyId = entry.companyId
		var company = game.GetCompany(companyId)
		if company == nil {
			// DEVIATION FROM JAVA: the company is dereferenced without a check
			// there; the port skips a company that is no longer in the model.
			continue
		}
		var line = "#" + util.FormatInt(int64(idx))
		if e.objectiveType == OBJECTIVE_PERFORMANCE {
			line += " - " + util.FormatInt(int64(company.GetLastPerformance()))
		} else {
			line += " - " + util.FormatThousandsDouble(float64(company.GetLastValue())*2) + " EUR"
		}
		var progressPercent = util.ToInt(entry.progress * 100)
		var lastProgress, found = e.progressHistoryByCompanyId[companyId]
		var lastProgressPercent = 0
		if found {
			lastProgressPercent = util.ToInt(lastProgress * 100)
		}
		line += " - " + util.FormatInt(int64(progressPercent)) + "%"
		line += " (" + util.Signed(progressPercent-lastProgressPercent) + "%)"
		line += " - " + company.GetName()
		var owner = e.externalUsers.GetOwnerOf(companyId)
		if e.objectiveReached {
			var points, hasPoints = e.scoreByCompanyId[companyId]
			var pointsText = "0"
			if hasPoints {
				pointsText = util.FormatInt(int64(points))
			}
			// DEVIATION FROM JAVA: getOwnerOf(...).getName() throws for a company
			// nobody owns; the port leaves the name empty.
			var ownerName = ""
			if owner != nil {
				ownerName = owner.GetName()
			}
			line += " " + ownerName + " score " + pointsText + " points on the leaderboard"
		} else {
			var ownerName = ""
			if owner != nil {
				ownerName = owner.GetName()
			}
			line += "(" + ownerName + ")"
		}
		line += "."
		if clientId < 0 {
			send.ChatBroadcast(line)
		} else {
			send.ChatClient(int64(clientId), line)
		}
	}
}

// broadcastVictory is broadcastVictory(Game openttd).
func (e *CompanyEconomyObjective) broadcastVictory(game *model.Game) {
	var send = e.base.GetSend()
	send.ChatBroadcast("Game Over : " + e.winnerName + " win !!!")
	e.showScore(-1, game)
	send.ChatBroadcast("Server will restart.")
}

// saveGame is saveGame(Game openttd).
func (e *CompanyEconomyObjective) saveGame(game *model.Game) {
	if e.externalGameService == nil {
		// DEVIATION FROM JAVA: an unconfigured service throws NullPointerException
		// here; the port logs a warning instead.
		companyEconomyObjectiveLog.Warn("ExternalGameService not received yet, game not saved.")
		return
	}
	e.externalGameService.SaveGame(e.scenarioId, e.calculateScores(game))
}

// endGame is endGame().
func (e *CompanyEconomyObjective) endGame() {
	if e.externalGameService == nil {
		companyEconomyObjectiveLog.Warn("ExternalGameService not received yet, game not ended.")
		return
	}
	e.externalGameService.EndGame(e.scenarioId)
}

// calculateScores is calculateScores(Game openttd): one GamePlayer per client that
// is logged in and owns a company, scaled so the best accomplishment is 1.0 and
// the longest duration is 1.0, and the points that scaling produces.
func (e *CompanyEconomyObjective) calculateScores(game *model.Game) []robotmodel.GamePlayer {
	var openttdDate = game.GetDate()
	var gamePlayers = []robotmodel.GamePlayer{}
	var maxDuration = int64(0)
	var maxAccomplishment = 0.0
	for _, client := range game.GetClients() {
		var externalUser = e.externalUsers.GetExternalUser(client.GetId())
		var company = game.GetCompany(client.GetCompany())
		if externalUser != nil && company != nil {
			var player = robotmodel.NewGamePlayer()
			player.SetExternalUser(externalUser)
			player.SetCompanyId(company.GetId())
			var inaugurationYear = company.GetInauguratedYear()
			var duration = int64(openttdDate.Year) - inaugurationYear
			player.SetDuration(float64(duration))
			if duration > maxDuration {
				maxDuration = duration
			}
			var accomplishment = 0.0
			if e.objectiveType == OBJECTIVE_PERFORMANCE {
				accomplishment = float64(company.GetLastPerformance())
			} else {
				accomplishment = float64(company.GetLastValue())
			}
			player.SetAccomplishment(accomplishment)
			if accomplishment > maxAccomplishment {
				maxAccomplishment = accomplishment
			}
			gamePlayers = append(gamePlayers, *player)
		}
	}
	for index, player := range gamePlayers {
		// Duration scaling to 0.0 ... 1.0
		var duration = 0.0
		if maxDuration != 0 {
			duration = player.GetDuration() / float64(maxDuration)
		}
		player.SetDuration(duration)
		// Accomplishment scaling to 0.0 ... 1.0
		var accomplishment = 0.0
		if maxAccomplishment != 0 {
			accomplishment = player.GetAccomplishment() / maxAccomplishment
		}
		player.SetAccomplishment(accomplishment)
		if accomplishment == 1.0 {
			player.SetWinner(true)
		}
		// Java's (int) cast saturates, so a player with no duration gets
		// Integer.MAX_VALUE points rather than a wrapped around number.
		e.scoreByCompanyId[player.GetCompanyId()] = util.ToInt(100 * accomplishment / duration)
		gamePlayers[index] = player
	}
	return gamePlayers
}

// isNewTrimester is isNewTrimester(Calendar now): MONTH / 3 + 4 * YEAR, which is
// util.Calendar.Trimester().
func (e *CompanyEconomyObjective) isNewTrimester(now util.Calendar) bool {
	var nowTrimester = now.Trimester()
	if nowTrimester > e.currentTrimester {
		e.currentTrimester = nowTrimester
		return true
	}
	return false
}

// isNewWeek is isNewWeek(Calendar now). Java reads Calendar.WEEK_OF_YEAR, which
// is locale sensitive; the port reads the ISO week of util.Calendar.WeekOfYear(),
// the deviation documented there.
func (e *CompanyEconomyObjective) isNewWeek(now util.Calendar) bool {
	var nowWeek = now.WeekOfYear()
	if nowWeek != e.currentWeek {
		e.currentWeek = nowWeek
		return true
	}
	return false
}

// getSortedProgressionByCompanyId is Collections.sort(entries, by value) followed
// by Collections.reverse: the companies, most progress first.
//
// DEVIATION FROM JAVA: the entries come from a HashMap, whose iteration order Java
// leaves unspecified, so equal progress came out in an arbitrary order. The port
// breaks ties by company id, which makes the leaderboard stable.
func (e *CompanyEconomyObjective) getSortedProgressionByCompanyId() []ProgressEntry {
	var entries = []ProgressEntry{}
	for companyId, progress := range e.progressByCompanyId {
		entries = append(entries, ProgressEntry{companyId: companyId, progress: progress})
	}
	sort.SliceStable(entries, func(i int, j int) bool {
		if entries[i].progress != entries[j].progress {
			return entries[i].progress < entries[j].progress
		}
		return entries[i].companyId < entries[j].companyId
	})
	var sorted = []ProgressEntry{}
	for index := len(entries) - 1; index >= 0; index -= 1 {
		sorted = append(sorted, entries[index])
	}
	return sorted
}

// gameOf is companyEvent.getOpenttd() / chatEvent.getOpenttd(), down-cast from
// the event's Openttd seam to the admin/model/Game the rules read.
func (e *CompanyEconomyObjective) gameOf(source event.Openttd) *model.Game {
	var game, ok = source.(*model.Game)
	if ok {
		return game
	}
	return nil
}
