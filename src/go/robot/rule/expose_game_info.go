// ExposeGameInfo is com.openttd.robot.rule.ExposeGameInfo: the rule that keeps
// the embedding application's copy of the game state fresh, once a week, so the
// strategyboard web pages have something to show between the game's own updates.
//
// It listens for DateEvent only. Unlike TimerObjective and
// CompanyEconomyObjective it reads the service from the singleton at call time
// instead of capturing it in a field, which is what Java does here, so a service
// injected after the rule was built is still used.
//
// There is no Java test for this class, so expose_game_info_test.go is new
// coverage written against this source as the specification.

package rule

import "com.openttd/admin"
import "com.openttd/admin/event"
import "com.openttd/admin/model"
import "com.openttd/robot"
import "com.openttd/util"

var exposeGameInfoLog = util.GetLogger("ExposeGameInfo")

type ExposeGameInfo struct {
	base *AbstractRule
	// scenarioId is scenarioId: the leaderboard row this game belongs to.
	scenarioId int64
	// currentWeek is currentWeek: the week it last exposed the game in. Java
	// starts it at -1, so the first DateEvent always exposes.
	currentWeek int
}

// NewExposeGameInfo is ExposeGameInfo(OpenttdAdmin openttdAdmin, long scenarioId).
func NewExposeGameInfo(openttd *admin.OpenttdAdmin, scenarioId int64) *ExposeGameInfo {
	var expose = &ExposeGameInfo{scenarioId: scenarioId, currentWeek: -1}
	expose.base = NewAbstractRule(openttd, expose, []event.EventKind{event.EVENT_DATE})
	return expose
}

// OnDateEvent is onDateEvent(DateEvent).
func (e *ExposeGameInfo) OnDateEvent(dateEvent *event.DateEvent) {
	if exposeGameInfoLog.IsDebugEnabled() {
		exposeGameInfoLog.Debug("ExposeGameInfo.onDateEvent " + dateEvent.ToString())
	}
	var game = e.gameOf(dateEvent.GetOpenttd())
	if game == nil {
		return
	}
	var now = game.GetDate()
	if e.isNewWeek(now) {
		var service = robot.GetInstance().GetExternalGameService()
		// DEVIATION FROM JAVA: an unconfigured service throws NullPointerException
		// there; the port logs a warning and does nothing.
		if service == nil {
			exposeGameInfoLog.Warn("ExternalGameService not received yet, game info not exposed.")
			return
		}
		service.ExposeGame(e.scenarioId, game.GetGameInfo())
	}
}

// isNewWeek is isNewWeek(Calendar now): true when the week number changed. Java
// reads Calendar.WEEK_OF_YEAR, which is locale sensitive; the port reads the
// ISO week of util.Calendar.WeekOfYear(), the deviation documented there.
func (e *ExposeGameInfo) isNewWeek(now util.Calendar) bool {
	var nowWeek = now.WeekOfYear()
	if nowWeek != e.currentWeek {
		e.currentWeek = nowWeek
		return true
	}
	return false
}

// gameOf is dateEvent.getOpenttd(), down-cast from the event's Openttd seam to
// the admin/model/Game the rules read.
func (e *ExposeGameInfo) gameOf(source event.Openttd) *model.Game {
	var game, ok = source.(*model.Game)
	if ok {
		return game
	}
	return nil
}
