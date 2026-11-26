// CompanyEvent mirrors com.openttd.admin.event.CompanyEvent: a company started,
// changed, or was removed. CompanyEventListener is com.openttd.admin.event
// .CompanyEventListener, kept in this file because it is one method.

package event

import "strconv"

// CompanyEventListener is CompanyEventListener.
type CompanyEventListener interface {
	OnCompanyEvent(event *CompanyEvent)
}

type CompanyEvent struct {
	openttd   Openttd
	companyId int
	action    Action
}

// NewCompanyEvent is CompanyEvent(Game openttd, short companyId, Action action):
// Java's short becomes an int, as everywhere else in the port.
func NewCompanyEvent(openttd Openttd, companyId int, action Action) *CompanyEvent {
	return &CompanyEvent{openttd: openttd, companyId: companyId, action: action}
}

// GetOpenttd is getOpenttd, typed as the Openttd seam rather than Game.
func (e *CompanyEvent) GetOpenttd() Openttd { return e.openttd }

// GetCompanyId is getCompanyId.
func (e *CompanyEvent) GetCompanyId() int { return e.companyId }

// GetAction is getAction.
func (e *CompanyEvent) GetAction() Action { return e.action }

// ToString is toString: the action's name, then the company id.
func (e *CompanyEvent) ToString() string {
	return ActionName(e.action) + ":" + strconv.Itoa(e.companyId)
}

// Kind is the CompanyEvent.class key this event is queued under.
func (e *CompanyEvent) Kind() EventKind { return EVENT_COMPANY }

// Notify is notify(CompanyEventListener).
func (e *CompanyEvent) Notify(listener EventListener) {
	var target, ok = listener.(CompanyEventListener)
	if ok {
		target.OnCompanyEvent(e)
		return
	}
	eventLog.Error(EventKindName(EVENT_COMPANY) + " cannot notify a listener that is not a CompanyEventListener: " + e.ToString())
}
