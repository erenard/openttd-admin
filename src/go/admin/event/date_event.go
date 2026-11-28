// DateEvent mirrors com.openttd.admin.event.DateEvent: the game's date changed.
// It carries no date of its own: its toString() asks the game model, which is
// the only thing the Openttd seam exists for. DateEventListener is com.openttd
// .admin.event.DateEventListener, kept in this file because it is one method.

package event

import "strconv"

// DateEventListener is DateEventListener.
type DateEventListener interface {
	OnDateEvent(event *DateEvent)
}

type DateEvent struct {
	openttd Openttd
}

// NewDateEvent is DateEvent(Game openttd).
func NewDateEvent(openttd Openttd) *DateEvent {
	return &DateEvent{openttd: openttd}
}

// GetOpenttd is getOpenttd, typed as the Openttd seam rather than Game.
func (e *DateEvent) GetOpenttd() Openttd { return e.openttd }

// ToString is toString: day-month-year of the model's date, the java Calendar
// fields in that order (Month is 0 based, as java.util.Calendar's is).
func (e *DateEvent) ToString() string {
	var date = e.openttd.GetDate()
	var sb = "DateEvent " + strconv.Itoa(date.DayOfMonth)
	sb = sb + "-" + strconv.Itoa(date.Month)
	sb = sb + "-" + strconv.Itoa(date.Year)
	return sb
}

// Kind is the DateEvent.class key this event is queued under.
func (e *DateEvent) Kind() EventKind { return EVENT_DATE }

// Notify is notify(DateEventListener).
func (e *DateEvent) Notify(listener EventListener) {
	var target, ok = listener.(DateEventListener)
	if ok {
		target.OnDateEvent(e)
		return
	}
	eventLog.Error(EventKindName(EVENT_DATE) + " cannot notify a listener that is not a DateEventListener: " + e.ToString())
}
