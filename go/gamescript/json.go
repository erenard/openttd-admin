// The JSON the gamescript speaks. Java uses gson: JsonObject keeps the order
// fields were added in, and the payload AdminCmd receives is that order. Go's
// encoding/json sorts map keys and turns every JSON number into a float64, so
// both directions are handled here: the requests are written by hand, and the
// replies are read by a scanner that keeps each value's original text, the way
// gson's JsonNumber does.
//
// AdminCmd's side is JSONDecode in src/squirrel/main.nut, which only needs "i",
// "c" and "a" to be present, so the field order is not load bearing on the wire;
// it is kept so the payload matches what the Java port put on the wire.
//
// The writers are exported because NetworkAdminSender builds the argument lists
// and the tests read the payloads back.

package gamescript

import "strconv"

// JsonField is one key and its already rendered value, in the order gson
// would have written them.
type JsonField struct {
	Key string
	Value string
}

// JsonObject is JsonObject.toString().
func JsonObject(fields []JsonField) string {
	var text = "{"
	for i := 0; i < len(fields); i++ {
		if i > 0 {
			text = text + ","
		}
		text = text + JsonString(fields[i].Key) + ":" + fields[i].Value
	}
	return text + "}"
}

// JsonList is JsonArray.toString(), given already rendered items.
func JsonList(items []string) string {
	var text = "["
	for i := 0; i < len(items); i++ {
		if i > 0 {
			text = text + ","
		}
		text = text + items[i]
	}
	return text + "]"
}

// JsonString is new JsonPrimitive(String). gson escapes the quote, the backslash
// and the control characters, and leaves every other byte alone, so UTF-8 in a
// news message or a goal text reaches the wire unchanged.
func JsonString(value string) string {
	var text = "\""
	for i := 0; i < len(value); i++ {
		var character = value[i]
		if character == 34 {
			text = text + "\\\""
		} else if character == 92 {
			text = text + "\\\\"
		} else if character == 8 {
			text = text + "\\b"
		} else if character == 9 {
			text = text + "\\t"
		} else if character == 10 {
			text = text + "\\n"
		} else if character == 12 {
			text = text + "\\f"
		} else if character == 13 {
			text = text + "\\r"
		} else if character < 32 {
			text = text + "\\u00" + hexPair(int(character))
		} else {
			text = text + value[i:i+1]
		}
	}
	return text + "\""
}

// JsonNumber is new JsonPrimitive(int).
func JsonNumber(value int) string {
	return strconv.FormatInt(int64(value), 10)
}

// JsonBool is new JsonPrimitive(boolean).
func JsonBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

// hexPair is the two hex digits gson writes for a control character.
func hexPair(value int) string {
	var digits = strconv.FormatInt(int64(value), 16)
	if len(digits) < 2 {
		return "0" + digits
	}
	return digits
}

// RawFields reads the top level fields of a JSON object and returns each value
// as the text the gamescript wrote: a quoted string unescaped, a number, a bool
// or null as its literal text. That is what gson does with a parsed number too,
// which is why AdminCmd's 50000000000 must not arrive as 5e10. Returns nil when
// the text is not a JSON object at all, which is how a reply gson could not map
// shows up.
func RawFields(text string) map[string]string {
	var n = len(text)
	var i = 0
	for (i < n && isSpace(text[i])) {
		i = i + 1
	}
	if i >= n || text[i] != 123 {
		return nil
	}
	var fields = map[string]string{}
	i = i + 1
	for (true) {
		for (i < n && isSpace(text[i])) {
			i = i + 1
		}
		if i >= n {
			return nil
		}
		if text[i] == 125 {
			break
		}
		if text[i] != 34 {
			return nil
		}
		var key, afterKey = scanString(text, i)
		if afterKey < 0 {
			return nil
		}
		i = afterKey
		for (i < n && isSpace(text[i])) {
			i = i + 1
		}
		if i >= n || text[i] != 58 {
			return nil
		}
		i = i + 1
		var value, afterValue = scanValue(text, i)
		if afterValue < 0 {
			return nil
		}
		i = afterValue
		fields[key] = value
		for (i < n && isSpace(text[i])) {
			i = i + 1
		}
		if i >= n {
			return nil
		}
		if text[i] == 44 {
			i = i + 1
		} else if text[i] == 125 {
			break
		} else {
			return nil
		}
	}
	return fields
}

// scanValue reads one JSON value and returns its text and the index just past
// it. Strings are unescaped; an object or an array is returned verbatim, since
// AdminCmd only ever answers with scalars and a nested value is not something
// the Java port could map into a String field either.
func scanValue(text string, index int) (string, int) {
	var n = len(text)
	var i = index
	for (i < n && isSpace(text[i])) {
		i = i + 1
	}
	if i >= n {
		return "", -1
	}
	if text[i] == 34 {
		return scanString(text, i)
	}
	if text[i] == 123 || text[i] == 91 {
		var start = i
		var depth = 0
		var inString = false
		for (i < n) {
			var character = text[i]
			if inString {
				if character == 92 {
					i = i + 2
					continue
				}
				if character == 34 {
					inString = false
				}
			} else if character == 34 {
				inString = true
			} else if character == 123 || character == 91 {
				depth = depth + 1
			} else if character == 125 || character == 93 {
				depth = depth - 1
				if depth == 0 {
					i = i + 1
					break
				}
			}
			i = i + 1
		}
		return text[start:i], i
	}
	var start = i
	for (i < n && text[i] != 44 && text[i] != 125 && text[i] != 93 && !isSpace(text[i])) {
		i = i + 1
	}
	return text[start:i], i
}

// scanString reads a quoted JSON string, unescaping it, and returns the index
// just past the closing quote, or -1 when the text is malformed. A \u escape is
// kept as written: the Java port only ever logged these values, and AdminCmd's
// own JSONDecode is the thing that has to agree with them.
func scanString(text string, index int) (string, int) {
	var n = len(text)
	var value = ""
	var i = index + 1
	for (i < n) {
		var character = text[i]
		if character == 34 {
			return value, i + 1
		}
		if character == 92 {
			i = i + 1
			if i >= n {
				return "", -1
			}
			var escape = text[i]
			if escape == 117 && i + 4 < n {
				value = value + "\\u" + text[i+1:i+5]
				i = i + 5
			} else if escape == 34 {
				value = value + "\""
				i = i + 1
			} else if escape == 92 {
				value = value + "\\"
				i = i + 1
			} else if escape == 47 {
				value = value + "/"
				i = i + 1
			} else if escape == 110 {
				value = value + "\n"
				i = i + 1
			} else if escape == 116 {
				value = value + "\t"
				i = i + 1
			} else if escape == 114 {
				value = value + "\r"
				i = i + 1
			} else if escape == 98 {
				value = value + "\b"
				i = i + 1
			} else if escape == 102 {
				value = value + "\f"
				i = i + 1
			} else if escape == 117 {
				value = value + "\\u"
				i = i + 1
			} else {
				value = value + text[i:i+1]
				i = i + 1
			}
		} else {
			value = value + text[i:i+1]
			i = i + 1
		}
	}
	return "", -1
}

// isSpace is the whitespace JSON allows between tokens.
func isSpace(character byte) bool {
	return character == 32 || character == 9 || character == 10 || character == 13
}
