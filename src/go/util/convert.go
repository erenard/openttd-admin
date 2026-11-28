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

// java.util.Calendar's DAY_OF_WEEK constants: SUNDAY is 1, so the demos'
// `date.get(Calendar.DAY_OF_WEEK) == Calendar.MONDAY` reads MONDAY as 2.
const (
	SUNDAY    = 1
	MONDAY    = 2
	TUESDAY   = 3
	WEDNESDAY = 4
	THURSDAY  = 5
	FRIDAY    = 6
	SATURDAY  = 7
)

// java.util.Calendar's month field: JANUARY is 0.
const (
	JANUARY   = 0
	FEBRUARY  = 1
	MARCH     = 2
	APRIL     = 3
	MAY       = 4
	JUNE      = 5
	JULY      = 6
	AUGUST    = 7
	SEPTEMBER = 8
	OCTOBER   = 9
	NOVEMBER  = 10
	DECEMBER  = 11
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
	// civilYear is the proleptic year of the same date, kept so the week and
	// month arithmetic below can work on a year that has a year 0.
	civilYear int
}

// DayToCalendar is dayToCalendar: a Calendar reflecting the ingame day.
func DayToCalendar(day int64) Calendar {
	return newCalendar(OPENTTD_TO_JAVA_OFFSET + (day+2)*OPENTTD_TO_JAVA_MULTIPLIER)
}

// CalendarToDay is calendarToDay: the inverse, exact because it reads the millis
// back rather than the rounded date fields.
func CalendarToDay(calendar Calendar) int64 {
	var millis = calendar.millis - OPENTTD_TO_JAVA_OFFSET
	return millis/OPENTTD_TO_JAVA_MULTIPLIER - 2
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
	return Calendar{
		millis: millis, Year: eraYear, Month: int(date.month), DayOfMonth: int(date.day),
		civilYear: int(date.year)}
}

// Clone is Java's (Calendar) date.clone(): an independent copy, so a rule can
// remember "the date I last did this" without it moving with the game clock.
func (c Calendar) Clone() Calendar {
	return c
}

// AddYears is date.add(Calendar.YEAR, nbYear): the game clock the rules step
// forward to find "the same day next year".
func (c Calendar) AddYears(nbYear int) Calendar {
	return shift(c, nbYear, 0)
}

// AddMonths is date.add(Calendar.MONTH, 1).
func (c Calendar) AddMonths(nbMonth int) Calendar {
	return shift(c, 0, nbMonth)
}

// After is date.after(other): strictly later, comparing the millis.
func (c Calendar) After(other Calendar) bool {
	return c.millis > other.millis
}

// Trimester is the Java idiom MONTH / 3 + 4 * YEAR the objectives use to ask
// whether a third of the game year has passed.
func (c Calendar) Trimester() int {
	return c.Month/3 + 4*c.civilYear
}

// MonthOfYear is the Java idiom MONTH + 12 * YEAR behind the monthly warnings.
func (c Calendar) MonthOfYear() int {
	return c.Month + 12*c.civilYear
}

// WeekOfYear is the ISO week, Monday based, the week 1 of a year being the week
// that holds 4 January. Java's getWEEK_OF_YEAR() is locale sensitive (its first
// day of the week is Sunday in the default locale), so the two agree on which
// dates share a week and differ only in the number printed for the week that
// straddles New Year; the rules only compare it for equality.
func (c Calendar) WeekOfYear() int {
	var ordinal = daysFromCivil(c.civilYear, c.Month+1, c.DayOfMonth)
	var monday = ordinal - mod(ordinal+3, 7)
	var year = c.civilYear
	var week = int((monday-weekOneMonday(year))/7 + 1)
	if week <= 0 {
		// Early January belongs to the last week of the year before.
		week = int((monday-weekOneMonday(year-1))/7 + 1)
	} else if week >= 53 {
		// Late December belongs to week 1 of the year after when its week holds
		// that year's 4 January.
		var next = weekOneMonday(year + 1)
		if monday >= next {
			week = int((monday-next)/7 + 1)
		}
	}
	return week
}

// DayOfWeek is get(DAY_OF_WEEK): 1 for a Sunday, 2 for a Monday, ... Java reads
// the field off a Calendar in the default locale; the game date has no timezone,
// so the port computes it from the civil date instead, which agrees because the
// Java epoch day is a Thursday in both.
func (c Calendar) DayOfWeek() int {
	var ordinal = daysFromCivil(c.civilYear, c.Month+1, c.DayOfMonth)
	return int(mod(ordinal+4, 7)) + 1
}

// shift is Calendar.add(YEAR, n) and Calendar.add(MONTH, n) for the fields the
// rules use: the day of month is clamped the way java.util.Calendar does when
// the target month is shorter (31 March plus one month is 30 April).
func shift(calendar Calendar, nbYear int, nbMonth int) Calendar {
	var year = calendar.civilYear + nbYear
	var month = calendar.Month + nbMonth
	year += month / 12
	month = int(mod(int64(month), 12))
	var day = calendar.DayOfMonth
	var last = daysInMonth(year, month)
	if day > last {
		day = last
	}
	// daysFromCivil counts civil days since 1970, and DayToCalendar's offset
	// already puts the game epoch at civil day -719528, so the millis of a civil
	// date are simply its civil day count times one day.
	return newCalendar(daysFromCivil(year, month+1, day) * OPENTTD_TO_JAVA_MULTIPLIER)
}

func daysInMonth(year int, month int) int {
	var days = []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	var length = days[month]
	if month == 1 && mod(int64(year), 4) == 0 && (mod(int64(year), 100) != 0 || mod(int64(year), 400) == 0) {
		length = 29
	}
	return length
}

// daysFromCivil is civilFromDays inverted: a proleptic Gregorian date to days
// since 1970-01-01.
func daysFromCivil(year int, month int, day int) int64 {
	var y = int64(year)
	var m = int64(month)
	var d = int64(day)
	if m <= 2 {
		y -= 1
	}
	var era = floorDiv(y, 400)
	var yoe = y - era*400
	// March based month: January and February count from the year before, so
	// they map to mp 9 and 10 (m + 9), and March..December to m - 3.
	var mp = m - 3
	if m <= 2 {
		mp = m + 9
	}
	var doy = (153*mp+2)/5 + d - 1
	var doe = yoe*365 + yoe/4 - yoe/100 + doy
	return era*146097 + doe - 719468
}

// weekOneMonday is the Monday of the week holding 4 January, which is where
// ISO week 1 of a year starts.
func weekOneMonday(year int) int64 {
	var jan4 = daysFromCivil(year, 1, 4)
	return jan4 - mod(jan4+3, 7)
}

// Mod is floorMod: a remainder that is never negative.
func mod(value int64, divisor int64) int64 {
	var remainder = value % divisor
	if remainder < 0 {
		remainder += divisor
	}
	return remainder
}

type civilDate struct {
	year  int64
	month int64
	day   int64
}

// civilFromDays is the inverse of days_from_civil: days since 1970-01-01 to a
// proleptic Gregorian date, the calendar java.util.Calendar's Gregorian change
// uses from 1582 on, which is the range every game date can reach.
func civilFromDays(z int64) civilDate {
	// Days since 1970-01-01 -> days since 0000-03-01, where the 400 year cycle
	// of the proleptic Gregorian calendar starts (719468 days apart).
	z += 719468
	var era = floorDiv(z, 146097)
	var doe = z - era*146097
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
	var day = doy - (153*mp+2)/5 + 1
	var month = mp
	if mp < 10 {
		month += 3
	} else {
		month -= 9
	}
	var year = yoe + era*400
	if month <= 2 {
		// January and February belong to the calendar year after the March-based
		// one the arithmetic counted.
		year += 1
	}
	return civilDate{year: year, month: month - 1, day: day}
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
	if value%divisor != 0 && (value < 0) != (divisor < 0) {
		quotient -= 1
	}
	return quotient
}
