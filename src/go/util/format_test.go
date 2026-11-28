// Format is the port's stand-in for the java.text.DecimalFormat, java.text.Date
// Format and java.util.Formatter calls the robot rules use to write their chat
// lines and leaderboard rows, so the strings they send have to be the ones the
// Java rules send.
//
// There is no Java test for this: test/java/com/openttd/util only holds
// ConvertTest.java, which covers the day to date conversion. These tests are new
// coverage written against this source as the specification, and they are what
// pins the exact strings the rule tests assert.

package util_test

import "strconv"
import "testing"
import "com.openttd/util"

// TestThousandsGroupsEveryThreeDigits is new DecimalFormat("#,##0").format(long)
// : grouping from the right, the leading group not padded, a minus sign kept,
// and zero written as "0" rather than the empty string DecimalFormat gives
// "#,##0" for in some locales.
func TestThousandsGroupsEveryThreeDigits(t *testing.T) {
	var values = []int64{int64(0), int64(7), int64(999), int64(1000),
		int64(1234567), int64(100000), int64(-1234), int64(-1000000)}
	var expected = []string{"0", "7", "999", "1,000", "1,234,567", "100,000",
		"-1,234", "-1,000,000"}
	for i := 0; i < len(values); i++ {
		var actual = util.FormatThousands(values[i])
		if actual != expected[i] {
			t.Errorf("FormatThousands(%s) = %s, expected %s",
				strconv.FormatInt(values[i], 10), actual, expected[i])
		}
	}
}

// TestThousandsOfADouble is the same pattern applied to the company value the
// economy objective reads from ADMIN_PACKET_SERVER_COMPANY_ECONOMY: Java rounds
// to an integer before grouping, so 999.4 is "999" and 999.5 is "1,000".
//
// DEVIATION FROM JAVA: Java's DecimalFormat rounds half to even, so 1000.5 is
// "1,000" there and "1,001" here; only values that are exactly x.5 differ.
func TestThousandsOfADouble(t *testing.T) {
	if util.FormatThousandsDouble(1000.0) != "1,000" {
		t.Errorf("1000.0 = %s", util.FormatThousandsDouble(1000.0))
	}
	if util.FormatThousandsDouble(999.4) != "999" {
		t.Errorf("999.4 = %s", util.FormatThousandsDouble(999.4))
	}
	if util.FormatThousandsDouble(999.5) != "1,000" {
		t.Errorf("999.5 = %s", util.FormatThousandsDouble(999.5))
	}
	if util.FormatThousandsDouble(0.0) != "0" {
		t.Errorf("0.0 = %s", util.FormatThousandsDouble(0.0))
	}
}

// TestToIntNarrowsTheWayJavaDoes is Java's (int) cast of a double, which the
// objectives reach for when they turn a percentage into a score: it truncates
// towards zero, and saturates instead of wrapping when the value does not fit
// an int - which is what happens when a company's duration is still zero.
func TestToIntNarrowsTheWayJavaDoes(t *testing.T) {
	if util.ToInt(2.9) != 2 || util.ToInt(-2.9) != -2 {
		t.Errorf("truncation: %d %d", util.ToInt(2.9), util.ToInt(-2.9))
	}
	// The dialect rejects a literal division by zero, so the three special
	// doubles are built from a variable.
	var zero = 0.0
	var nan = zero / zero
	var infinity = 1.0 / zero
	if util.ToInt(nan) != 0 {
		t.Errorf("NaN = %d", util.ToInt(nan))
	}
	if util.ToInt(infinity) != 2147483647 {
		t.Errorf("+infinity = %d", util.ToInt(infinity))
	}
	if util.ToInt(-infinity) != -2147483648 {
		t.Errorf("-infinity = %d", util.ToInt(-infinity))
	}
	if util.ToInt(1e12) != 2147483647 {
		t.Errorf("1e12 = %d", util.ToInt(1e12))
	}
}

// TestDateLongUK is DateFormat.getDateInstance(DateFormat.LONG, Locale.UK): the
// British long date, day not padded, full month name, and the game year rather
// than the calendar year, which is what the restart line of TimerObjective
// prints.
func TestDateLongUK(t *testing.T) {
	// 712700 is 23 April 1951, 712595 is 8 January 1951.
	if util.FormatDateLongUK(util.DayToCalendar(712700)) != "23 April 1951" {
		t.Errorf("712700 = %s", util.FormatDateLongUK(util.DayToCalendar(712700)))
	}
	if util.FormatDateLongUK(util.DayToCalendar(712595)) != "8 January 1951" {
		t.Errorf("712595 = %s", util.FormatDateLongUK(util.DayToCalendar(712595)))
	}
	// Day 0 is the first day of game year 0, and the month names are indexed
	// from it, so the very first date reads "1 January 1".
	if util.FormatDateLongUK(util.DayToCalendar(0)) != "1 January 1" {
		t.Errorf("day 0 = %s", util.FormatDateLongUK(util.DayToCalendar(0)))
	}
	if len(util.UKMonthNames) != 12 || util.UKMonthNames[10] != "November" {
		t.Errorf("month names %s", util.UKMonthNames[10])
	}
}

// TestPaddingIsTheFormatterField is the %255s and %-3s conversion of the
// Formatter calls in CompanyLifeCycle.showCompanies and the objectives: the
// value is padded to the width of the field, and a value that is already wider
// than the field is left alone rather than truncated.
func TestPaddingIsTheFormatterField(t *testing.T) {
	if util.PadLeft("6", 3) != "  6" {
		t.Errorf("PadLeft(6, 3) = %s", strconv.Quote(util.PadLeft("6", 3)))
	}
	if util.PadRight("First", 10) != "First     " {
		t.Errorf("PadRight(First, 10) = %s", strconv.Quote(util.PadRight("First", 10)))
	}
	if util.PadLeft("abcdefgh", 2) != "abcdefgh" {
		t.Errorf("a wide value must not be cut: %s", util.PadLeft("abcdefgh", 2))
	}
	if len(util.PadRight("First", 255)) != 255 {
		t.Errorf("the CEO column is %d wide", len(util.PadRight("First", 255)))
	}
	if util.Spaces(3) != "   " || util.Spaces(0) != "" {
		t.Errorf("Spaces: %s / %s", strconv.Quote(util.Spaces(3)), strconv.Quote(util.Spaces(0)))
	}
}

// TestSignedIsThePercentPlusConversion is the %+d the objectives use for the
// change of performance in a score row: a positive difference keeps an explicit
// "+", which is what makes "(+50%)" and "(-5%)" readable in the chat.
func TestSignedIsThePercentPlusConversion(t *testing.T) {
	if util.Signed(50) != "+50" {
		t.Errorf("Signed(50) = %s", util.Signed(50))
	}
	if util.Signed(-5) != "-5" {
		t.Errorf("Signed(-5) = %s", util.Signed(-5))
	}
	if util.Signed(0) != "+0" {
		t.Errorf("Signed(0) = %s", util.Signed(0))
	}
}

// TestDoubleKeepsTheDecimalPoint is StringBuilder.append(double), which is what
// GamePlayer.toString writes for the accomplishment and the duration of a
// leaderboard row: Java always shows a decimal point, so an integral value
// reads "1.0" and not "1".
//
// DEVIATION FROM JAVA: Java switches to "1.0E10" beyond 1e7, the port keeps
// writing plain digits; the values the rules carry are far below that.
func TestDoubleKeepsTheDecimalPoint(t *testing.T) {
	if util.FormatDouble(0.0) != "0.0" {
		t.Errorf("0.0 = %s", util.FormatDouble(0.0))
	}
	if util.FormatDouble(1.0) != "1.0" {
		t.Errorf("1.0 = %s", util.FormatDouble(1.0))
	}
	if util.FormatDouble(1.5) != "1.5" {
		t.Errorf("1.5 = %s", util.FormatDouble(1.5))
	}
	if util.FormatDouble(-2.0) != "-2.0" {
		t.Errorf("-2.0 = %s", util.FormatDouble(-2.0))
	}
	if util.FormatInt(42) != "42" {
		t.Errorf("FormatInt(42) = %s", util.FormatInt(42))
	}
}
