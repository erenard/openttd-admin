// parse holds the port's stand-in for Java's Integer.parseInt and
// Short.parseShort. Those throw NumberFormatException, which the rules catch;
// the port returns whether the text parsed, so the rules can send their usage
// line instead. The range checks match the Java methods: a value outside the
// target type's range is reported as unparsable, not silently truncated.

package rule

import "strings"

// MIN_SHORT and MAX_SHORT are Short.MIN_VALUE and Short.MAX_VALUE.
const MIN_SHORT = -32768
const MAX_SHORT = 32767

// MIN_INT and MAX_INT are Integer.MIN_VALUE and Integer.MAX_VALUE.
const MIN_INT = -2147483648
const MAX_INT = 2147483647

// ParseShort is Short.parseShort(text): see ParseNumber for what it accepts.
func ParseShort(text string) (int, bool) {
	var value, ok = ParseNumber(text)
	if !ok {
		return 0, false
	}
	if value < MIN_SHORT || value > MAX_SHORT {
		return 0, false
	}
	return int(value), true
}

// ParseInt is Integer.parseInt(text): see ParseNumber for what it accepts.
func ParseInt(text string) (int, bool) {
	var value, ok = ParseNumber(text)
	if !ok {
		return 0, false
	}
	if value < MIN_INT || value > MAX_INT {
		return 0, false
	}
	return int(value), true
}

// ParseNumber is the shared body: an optional sign and at least one digit, and
// nothing else. Java's parse methods reject the same shapes, plus a leading
// plus sign, which they accept; the port accepts it too.
func ParseNumber(text string) (int64, bool) {
	var digits = text
	var sign = int64(1)
	if strings.HasPrefix(digits, "-") {
		sign = -1
		digits = digits[1:]
	} else if strings.HasPrefix(digits, "+") {
		digits = digits[1:]
	}
	if digits == "" {
		return 0, false
	}
	var value = int64(0)
	for _, digit := range digits {
		if digit < '0' || digit > '9' {
			return 0, false
		}
		value = value*10 + int64(digit-'0')
		if value > int64(MAX_INT)+1 {
			// No Java integer type can hold this, so both parse methods would
			// have thrown; the extra one is Integer.MIN_VALUE's magnitude.
			return 0, false
		}
	}
	return sign * value, true
}
