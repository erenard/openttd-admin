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
		var startOfYear = year * 365 + ((year - 1) / 4 - (year - 1) / 100 + (year - 1) / 400) + 1
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
	var dates = []int64{0, 1, 711751, 719527, 719528, int64(1950 * 365 + 1), 1000000}
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
	var startOfYear = year * 365 + ((year - 1) / 4 - (year - 1) / 100 + (year - 1) / 400) + 1
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
