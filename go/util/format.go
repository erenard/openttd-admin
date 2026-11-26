// Format is the port's stand-in for the java.text and java.util.Formatter calls
// the robot rules use to write their chat and leaderboard messages. Java's
// DecimalFormat, DateFormat and %255s style padding are all reproduced here so
// the strings the rules send are byte-identical to what the Java rules send.

package util

import "math"
import "strconv"
import "strings"

// UKMonthNames is DateFormat.getDateInstance(DateFormat.LONG, Locale.UK)'s month
// names: the British long form, "19 November 2026".
var UKMonthNames = []string{
	"January", "February", "March", "April", "May", "June", "July",
	"August", "September", "October", "November", "December"}

// FormatThousands is new DecimalFormat("#,##0").format(value): grouping every
// three digits, no decimals, a minus sign in front when the value is negative.
func FormatThousands(value int64) string {
	var negative = value < 0
	var remaining = value
	if negative {
		remaining = -value
	}
	var digits = ""
	for remaining >= 1000 {
		var group = int(remaining % 1000)
		remaining /= 1000
		// The leading group keeps no padding, the ones behind it are always
		// three digits wide.
		if group < 100 {
			var padded = FormatInt(int64(group))
			for len(padded) < 3 {
				padded = "0" + padded
			}
			digits = "," + padded + digits
		} else {
			digits = "," + FormatInt(int64(group)) + digits
		}
	}
	if remaining > 0 || digits == "" {
		digits = FormatInt(remaining) + digits
	}
	if negative {
		return "-" + digits
	}
	return digits
}

// FormatThousandsDouble is new DecimalFormat("#,##0").format(value) for a double
// , which is what the rules reach for when they format a company's value. Java
// rounds to the nearest integer with its default HALF_EVEN rounding mode before
// grouping; the port rounds half away from zero, which differs only for a value
// that is exactly x.5.
func FormatThousandsDouble(value float64) string {
	return FormatThousands(int64(ToInt(math.Round(value))))
}

// ToInt is Java's (int) narrowing conversion of a double: NaN becomes 0, positive
// infinity Integer.MAX_VALUE, negative infinity Integer.MIN_VALUE, and every
// other value is truncated towards zero. Java saturates instead of wrapping, and
// the rules depend on that when they divide a score by a duration that is zero.
func ToInt(value float64) int {
	if value != value {
		return 0
	}
	if value >= 2147483647.0 {
		return 2147483647
	}
	if value <= -2147483648.0 {
		return -2147483648
	}
	return int(value)
}

// FormatDateLongUK is DateFormat.getDateInstance(DateFormat.LONG, Locale.UK)
// .format(date): "d MMMM yyyy" with the day not zero padded.
func FormatDateLongUK(calendar Calendar) string {
	return FormatInt(int64(calendar.DayOfMonth)) + " " + UKMonthNames[calendar.Month] + " " + FormatInt(int64(calendar.Year))
}

// PadLeft is the %<width>s conversion: the value right aligned in a field of
// width characters, padded with spaces. A value already at least that wide is
// left alone, which is what Java's Formatter does.
func PadLeft(value string, width int) string {
	var padding = width - len(value)
	if padding <= 0 {
		return value
	}
	return Spaces(padding) + value
}

// PadRight is the %-<width>s conversion: the value left aligned in the field.
func PadRight(value string, width int) string {
	var padding = width - len(value)
	if padding <= 0 {
		return value
	}
	return value + Spaces(padding)
}

// Signed is the %+d conversion: an explicit sign in front of the number.
func Signed(value int) string {
	if value >= 0 {
		return "+" + FormatInt(int64(value))
	}
	return FormatInt(int64(value))
}

// FormatDouble is Double.toString: what StringBuilder.append(double) writes.
// Java always shows a decimal point ("0.0", "1.5"); the port keeps that for the
// values GamePlayer carries, which are small enough that Java's scientific
// notation never applies to them.
func FormatDouble(value float64) string {
	var text = strconv.FormatFloat(value, 'g', -1, 64)
	if !hasDot(text) && !hasExponent(text) {
		return text + ".0"
	}
	return text
}

func hasDot(value string) bool {
	return strings.Index(value, ".") != -1
}

func hasExponent(value string) bool {
	return strings.Index(value, "e") != -1 || strings.Index(value, "E") != -1
}

// FormatInt is Long.toString: the plain decimal form the formatters start from.
func FormatInt(value int64) string {
	return strconv.FormatInt(value, 10)
}

func Spaces(count int) string {
	var spaces = ""
	for count > 0 {
		spaces += " "
		count -= 1
	}
	return spaces
}
