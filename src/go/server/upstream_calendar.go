// UpstreamCalendar mirrors com.openttd.server.UpstreamCalendar: the reference
// implementation of OpenTTD 15.3's calendar arithmetic, copied from
// `git show 15.3:src/timer/timer_game_common.cpp` (CalendarConvertDateToYMD) and
// timer_game_common.h (CalendarTime::DAYS_IN_YEAR / DAYS_IN_LEAP_YEAR).
//
// It pins Convert.dayToCalendar(): an admin gets game dates as a uint32 day
// counter, so a date that lands in the wrong year makes DateEvent consumers
// (strategyboard's objectives, openttd-robot's timers) act on the wrong day.
//
// OpenTTD counts year 0, so a date of `year * 365 + 1` is not that year: leap
// days have to be subtracted, which is exactly what YearOf does.
package server

// DAYS_IN_YEAR and DAYS_IN_LEAP_YEAR are CalendarTime's constants.
const DAYS_IN_YEAR = 365
const DAYS_IN_LEAP_YEAR = 366

// IsLeapYear is git show 15.3:src/timer/timer_game_common.h, CalendarTime::IsLeapYear.
func IsLeapYear(year int) bool {
	return (year%4 == 0 && (year%100 != 0 || year%400 == 0)) || (year%4000 == 0)
}

// DaysInYear is the length of one year, the value the C++ loop compares against.
func DaysInYear(year int) int {
	if IsLeapYear(year) {
		return DAYS_IN_LEAP_YEAR
	}
	return DAYS_IN_YEAR
}

// YearOf is the year part of CalendarConvertDateToYMD(date): large steps first
// (400 years hold 97 leap years), then centuries, then the 4 year cycles.
//
// date is the game date, days since year 0.
func YearOf(date int) int {
	var year = 400 * (date / (DAYS_IN_YEAR*400 + 97))
	var remaining = date % (DAYS_IN_YEAR*400 + 97)

	if remaining >= DAYS_IN_YEAR*100+25 {
		year += 100
		remaining -= DAYS_IN_YEAR*100 + 25
		year += 100 * (remaining / (DAYS_IN_YEAR*100 + 24))
		remaining = remaining % (DAYS_IN_YEAR*100 + 24)
	}

	if !IsLeapYear(year) && remaining >= DAYS_IN_YEAR*4 {
		year += 4
		remaining -= DAYS_IN_YEAR * 4
	}

	year += 4 * (remaining / (DAYS_IN_YEAR*4 + 1))
	remaining = remaining % (DAYS_IN_YEAR*4 + 1)

	for remaining >= DaysInYear(year) {
		remaining -= DaysInYear(year)
		year += 1
	}
	return year
}
