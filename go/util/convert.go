// Convert mirrors com.openttd.util.Convert: turning the uint32 day counter the
// admin network sends (ADMIN_PACKET_SERVER_DATE) into a calendar date.
//
// Java returns a java.util.Calendar; this port returns its own Calendar struct,
// because the fields the consumers read (YEAR, MONTH, DAY_OF_MONTH) are all this
// port needs, and computing them as civil days avoids the local-timezone and
// Julian-cutover behaviour java.util.Calendar.getInstance() brings in. The
// millis value is kept so CalendarToDay stays the exact inverse, as Java's is.
package util

// OPENTTD_TO_JAVA_OFFSET is Java's openttdToJavaOffset: the millis between the
// proleptic year 0 game date and the Java epoch.
const OPENTTD_TO_JAVA_OFFSET = -62167392000000

// OPENTTD_TO_JAVA_MULTIPLIER is one day in millis.
const OPENTTD_TO_JAVA_MULTIPLIER = 24 * 60 * 60 * 1000

// java.util.Calendar's month field: JANUARY is 0.
const (
	JANUARY = 0
	FEBRUARY = 1
	MARCH = 2
	APRIL = 3
	MAY = 4
	JUNE = 5
	JULY = 6
	AUGUST = 7
	SEPTEMBER = 8
	OCTOBER = 9
	NOVEMBER = 10
	DECEMBER = 11
)

type Calendar struct {
	// millis is getTimeInMillis(), the value CalendarToDay works from.
	millis int64
	// Year is get(YEAR): the era-based year, so the proleptic year 0 and below
	// read as 1 BC, 2 BC, ... That is why game year 0 reports java year 1.
	Year int
	// Month is get(MONTH), 0 based.
	Month int
	// DayOfMonth is get(DAY_OF_MONTH), 1 based.
	DayOfMonth int
}

// DayToCalendar is dayToCalendar: a Calendar reflecting the ingame day.
func DayToCalendar(day int64) Calendar {
	return newCalendar(OPENTTD_TO_JAVA_OFFSET + (day+2) * OPENTTD_TO_JAVA_MULTIPLIER)
}

// CalendarToDay is calendarToDay: the inverse, exact because it reads the millis
// back rather than the rounded date fields.
func CalendarToDay(calendar Calendar) int64 {
	var millis = calendar.millis - OPENTTD_TO_JAVA_OFFSET
	return millis / OPENTTD_TO_JAVA_MULTIPLIER - 2
}

// MillisOf is the accessor Java's Calendar.getTimeInMillis() gives.
func (c Calendar) GetMillis() int64 { return c.millis }

func newCalendar(millis int64) Calendar {
	var date = civilFromDays(floorDiv(millis, OPENTTD_TO_JAVA_MULTIPLIER))
	var eraYear = int(date.year)
	if eraYear <= 0 {
		// The AD era has no year 0: 0 CE is 1 BC, -1 CE is 2 BC.
		eraYear = 1 - int(date.year)
	}
	return Calendar{millis: millis, Year: eraYear, Month: int(date.month), DayOfMonth: int(date.day)}
}

type civilDate struct {
	year int64
	month int64
	day int64
}

// civilFromDays is the inverse of days_from_civil: days since 1970-01-01 to a
// proleptic Gregorian date, the calendar java.util.Calendar's Gregorian change
// uses from 1582 on, which is the range every game date can reach.
func civilFromDays(z int64) civilDate {
	// Days since 1970-01-01 -> days since 0000-03-01, where the 400 year cycle
	// of the proleptic Gregorian calendar starts (719468 days apart).
	z += 719468
	var era = floorDiv(z, 146097)
	var doe = z - era * 146097
	// doe counts from 0000-03-01 of the era, so a year's leap day belongs to the
	// year before it: era year y starts on dayOfEraStart(y).
	var low = int64(0)
	var high = int64(399)
	for low < high {
		var middle = (low + high + 1) / 2
		if dayOfEraStart(middle) <= doe {
			low = middle
		} else {
			high = middle - 1
		}
	}
	var yoe = low
	var doy = doe - dayOfEraStart(yoe)
	// doy is the day within a March-based year, which is where the leap day sits,
	// so the month formula needs no leap-year table.
	var mp = (5*doy + 2) / 153
	var day = doy - (153*mp + 2)/5 + 1
	var month = mp
	if mp < 10 {
		month += 3
	} else {
		month -= 9
	}
	var year = yoe + era * 400
	if month <= 2 {
		// January and February belong to the calendar year after the March-based
		// one the arithmetic counted.
		year += 1
	}
	return civilDate{year: year, month: month-1, day: day}
}

// dayOfEraStart is the day era year y begins on: 365 days a year plus the leap
// days of the years before it.
func dayOfEraStart(y int64) int64 {
	return 365*y + y/4 - y/100 + y/400
}

// FloorDiv is the division java.util.Calendar's millis arithmetic implies: Java
// truncates toward zero, a calendar day count rounds toward negative infinity.
func floorDiv(value int64, divisor int64) int64 {
	var quotient = value / divisor
	if value % divisor != 0 && (value < 0) != (divisor < 0) {
		quotient -= 1
	}
	return quotient
}
