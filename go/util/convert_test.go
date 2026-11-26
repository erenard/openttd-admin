// Convert turns the uint32 day counter the admin network sends (ADMIN_PACKET
// SERVER_DATE) into a calendar date. Checked against 15.3's CalendarConvertDateToYMD,
// see UpstreamCalendar.
//
// See test/java/com/openttd/util/ConvertTest.java.

package util_test

import "strconv"
import "testing"
import "com.openttd/util"
import "com.openttd/server"

// TestYearMatchesUpstream: every game date must land in the year OpenTTD
// computes. Checked from 1900 to 2050 - OpenTTD's MAX_YEAR is 2050 - and across
// century years, which are the ones the leap rules disagree about.
func TestYearMatchesUpstream(t *testing.T) {
	for year := 1900; year <= 2050; year++ {
		// DateAtStartOfYear(year): 365 days per year plus one day for each
		// leap year before it, year 0 included.
		var startOfYear = year*365 + ((year-1)/4 - (year-1)/100 + (year-1)/400) + 1
		var yearLength = server.DaysInYear(year)
		for dayOfYear := 0; dayOfYear < yearLength; dayOfYear += 7 {
			var date = startOfYear + dayOfYear
			var calendar = util.DayToCalendar(int64(date))
			if calendar.Year != year {
				t.Errorf("date %s (year %d, day %d of %d days): expected year %d, got %d",
					strconv.Itoa(date), year, dayOfYear, yearLength, year, calendar.Year)
			}
		}
	}
}

// TestYearZero: OpenTTD counts year 0, java.util.Calendar's AD era does not,
// so dates in game year 0 map to java year 1. Only relevant for a map starting
// at year 0, so this is documented rather than fixed.
func TestYearZero(t *testing.T) {
	if server.YearOf(0) != 0 {
		t.Errorf("expected upstream year 0 for date 0, got %d", server.YearOf(0))
	}
	if util.DayToCalendar(0).Year != 1 {
		t.Errorf("expected java year 1 for game date 0, got %d", util.DayToCalendar(0).Year)
	}
}

// TestRoundTrip: objectives and robot rules compare calendars against game dates.
func TestRoundTrip(t *testing.T) {
	var dates = []int64{0, 1, 711751, 719527, 719528, int64(1950*365 + 1), 1000000}
	for i := 0; i < len(dates); i++ {
		if util.CalendarToDay(util.DayToCalendar(dates[i])) != dates[i] {
			t.Errorf("expected date %d to survive a round trip", dates[i])
		}
	}
}

// TestStartOfGameYear: January 1st of a game year, the date a new map reports
// in its WELCOME packet.
func TestStartOfGameYear(t *testing.T) {
	var year = 1950
	var startOfYear = year*365 + ((year-1)/4 - (year-1)/100 + (year-1)/400) + 1
	var calendar = util.DayToCalendar(int64(startOfYear))
	if calendar.Year != year {
		t.Errorf("expected year %d, got %d", year, calendar.Year)
	}
	if calendar.Month != util.JANUARY {
		t.Errorf("expected January, got month %d", calendar.Month)
	}
	if calendar.DayOfMonth != 1 {
		t.Errorf("expected day 1, got %d", calendar.DayOfMonth)
	}
	if !server.IsLeapYear(2000) {
		t.Errorf("expected 2000 to be a leap year")
	}
	if server.IsLeapYear(1900) {
		t.Errorf("expected 1900 not to be a leap year")
	}
}

// TestAddYearsAndMonths: the robot rules step the game clock forward to find a
// deadline - "nbYear game years from now", "one month after the game ended" -
// which is Java's Calendar.add(YEAR, n) / add(MONTH, n). Java keeps the day of
// the month and clamps it to the last day of the new month, so 31 January plus a
// month is 28 February, not 3 March, and 29 February plus a year is 28 February.
//
// There is no Java test for this: ConvertTest.java only checks the day to date
// conversion. The port rebuilds the millis from the civil date instead of using a
// timezone, and this is the test that caught the civil day count being mixed up
// with the game day counter.
func TestAddYearsAndMonths(t *testing.T) {
	// 712595 is 8 January 1951, 712960 is 8 January 1952, 713326 is 8 January
	// 1953: 1952 is the leap year in between, so the second year is 366 days.
	var january8 = util.DayToCalendar(712595)
	if util.CalendarToDay(january8.Clone().AddYears(1)) != 712960 {
		t.Errorf("8 January 1951 plus a year = %d",
			util.CalendarToDay(january8.Clone().AddYears(1)))
	}
	if util.CalendarToDay(january8.Clone().AddYears(2)) != 713326 {
		t.Errorf("8 January 1951 plus two years = %d",
			util.CalendarToDay(january8.Clone().AddYears(2)))
	}
	if util.CalendarToDay(january8.Clone().AddMonths(12)) != 712960 {
		t.Errorf("8 January 1951 plus twelve months = %d",
			util.CalendarToDay(january8.Clone().AddMonths(12)))
	}
	// A month, the way TimerObjective sets the restart date.
	if util.CalendarToDay(january8.Clone().AddMonths(1)) != 712626 {
		t.Errorf("8 January 1951 plus a month = %d",
			util.CalendarToDay(january8.Clone().AddMonths(1)))
	}
	// 712618 is 31 January 1951: February 1951 has 28 days, so the day clamps.
	var january31 = util.DayToCalendar(712618)
	if util.CalendarToDay(january31.Clone().AddMonths(1)) != 712646 {
		t.Errorf("31 January 1951 plus a month = %d",
			util.CalendarToDay(january31.Clone().AddMonths(1)))
	}
	// 713012 is 29 February 1952, and 1953 has no 29 February.
	var february29 = util.DayToCalendar(713012)
	if util.CalendarToDay(february29.Clone().AddYears(1)) != 713377 {
		t.Errorf("29 February 1952 plus a year = %d",
			util.CalendarToDay(february29.Clone().AddYears(1)))
	}
	// add() must not move the calendar it is called on: Java's clone() above
	// stands for the port's value copy, and the rule keeps the original date.
	if util.CalendarToDay(january8) != 712595 {
		t.Errorf("the calendar moved to %d", util.CalendarToDay(january8))
	}
}

// TestWeekTrimesterMonthOfYear: the three counters the objectives use to decide
// when to act. WeekOfYear is the ISO week, Monday based, which is a documented
// deviation from Java's locale sensitive getWEEK_OF_YEAR(); Trimester and Month
// OfYear are the MONTH / 3 + 4 * YEAR and MONTH + 12 * YEAR idioms the Java rules
// write out, counted from the civil year so game year 0 works.
func TestWeekTrimesterMonthOfYear(t *testing.T) {
	// 1 January 1951 is a Monday, so ISO week 1 of 1951 is 1..7 January and 8
	// January starts week 2. 23 April 1951 is the 113th day of the year, in week
	// 17.
	if util.DayToCalendar(712595).WeekOfYear() != 2 {
		t.Errorf("8 January 1951 week = %d", util.DayToCalendar(712595).WeekOfYear())
	}
	if util.DayToCalendar(712595+7).WeekOfYear() != 3 {
		t.Errorf("15 January 1951 week = %d", util.DayToCalendar(712595+7).WeekOfYear())
	}
	if util.DayToCalendar(712700).WeekOfYear() != 17 {
		t.Errorf("23 April 1951 week = %d", util.DayToCalendar(712700).WeekOfYear())
	}
	// The trimester and the month of the game both count from the start of the
	// game year, so a new one is a new quarter or a new month of that year.
	if util.DayToCalendar(712595).Trimester() != 7804 {
		t.Errorf("8 January 1951 trimester = %d", util.DayToCalendar(712595).Trimester())
	}
	if util.DayToCalendar(712700).Trimester() != 7805 {
		t.Errorf("23 April 1951 trimester = %d", util.DayToCalendar(712700).Trimester())
	}
	if util.DayToCalendar(712595).MonthOfYear() != 23412 {
		t.Errorf("8 January 1951 month of year = %d", util.DayToCalendar(712595).MonthOfYear())
	}
	if util.DayToCalendar(712700).MonthOfYear() != 23415 {
		t.Errorf("23 April 1951 month of year = %d", util.DayToCalendar(712700).MonthOfYear())
	}
	// Day 0 is 1 January of game year 0, the first month and the first trimester
	// of it, which is why a rule whose counter starts at 0 does not fire on the
	// very first date packet.
	if util.DayToCalendar(0).MonthOfYear() != 0 || util.DayToCalendar(0).Trimester() != 0 {
		t.Errorf("day 0 month %d trimester %d",
			util.DayToCalendar(0).MonthOfYear(), util.DayToCalendar(0).Trimester())
	}
	if util.DayToCalendar(31).MonthOfYear() != 1 || util.DayToCalendar(31).Trimester() != 0 {
		t.Errorf("day 31 month %d trimester %d",
			util.DayToCalendar(31).MonthOfYear(), util.DayToCalendar(31).Trimester())
	}
}
