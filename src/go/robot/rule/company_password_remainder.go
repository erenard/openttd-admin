// CompanyPasswordRemainder is com.openttd.robot.rule.CompanyPasswordRemainder:
// the monthly nag that reminds a company without a password to set one.
//
// It listens for DateEvent only, and warns once per month, which is why it
// caches the month it last spoke in. Java computes that key as
// `MONTH + 12 * YEAR`, which is exactly util.Calendar.MonthOfYear().
//
// There is no Java test for this class, so company_password_remainder_test.go is
// new coverage written against this source as the specification.

package rule

import "com.openttd/admin"
import "com.openttd/admin/event"
import "com.openttd/admin/model"
import "com.openttd/util"

var companyPasswordRemainderLog = util.GetLogger("CompanyPasswordRemainder")

type CompanyPasswordRemainder struct {
	base *AbstractRule
	// currentMonth is currentMonth: Java leaves it at 0, which is January of
	// java year 0, so the first DateEvent of a game always warns.
	currentMonth int
}

// NewCompanyPasswordRemainder is CompanyPasswordRemainder(OpenttdAdmin
// openttdAdmin).
func NewCompanyPasswordRemainder(openttd *admin.OpenttdAdmin) *CompanyPasswordRemainder {
	var remainder = &CompanyPasswordRemainder{currentMonth: 0}
	remainder.base = NewAbstractRule(openttd, remainder, []event.EventKind{event.EVENT_DATE})
	return remainder
}

// OnDateEvent is onDateEvent(DateEvent).
func (r *CompanyPasswordRemainder) OnDateEvent(dateEvent *event.DateEvent) {
	if companyPasswordRemainderLog.IsDebugEnabled() {
		companyPasswordRemainderLog.Debug("CompanyPasswordRemainder.onDateEvent " + dateEvent.ToString())
	}
	var game = r.gameOf(dateEvent.GetOpenttd())
	if game == nil {
		return
	}
	var now = game.GetDate()
	if r.isNewMonth(now) {
		for _, company := range game.GetCompanies() {
			if company != nil && !company.IsUsePassword() {
				r.warningNoPassword(company.GetId())
			}
		}
	}
}

// isNewMonth is isNewMonth(Calendar now): true when the month key moved forward.
func (r *CompanyPasswordRemainder) isNewMonth(now util.Calendar) bool {
	var nowMonth = now.MonthOfYear()
	if nowMonth > r.currentMonth {
		r.currentMonth = nowMonth
		return true
	}
	return false
}

// warningNoPassword is warningNoPassword(int companyId).
func (r *CompanyPasswordRemainder) warningNoPassword(companyId int) {
	r.base.GetSend().ChatCompany(companyId, "Warning: your company has NO PASSWORD.")
}

// gameOf is dateEvent.getOpenttd(), down-cast from the event's Openttd seam to
// the admin/model/Game the rules read.
func (r *CompanyPasswordRemainder) gameOf(source event.Openttd) *model.Game {
	var game, ok = source.(*model.Game)
	if ok {
		return game
	}
	return nil
}
