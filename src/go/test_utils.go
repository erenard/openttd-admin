// TestUtils mirrors com.openttd.TestUtils: the polling helpers the tests use
// instead of sleeping a fixed amount of time. Java's await throws
// AssertionError; the Go port panics with the same message, which fails the
// test that is waiting.
package openttd

import "errors"
import "os"
import "strconv"
import "strings"
import "time"

// POLL_MILLIS is the interval Await checks the condition at.
const POLL_MILLIS = 50

// Wait is wait(s): a fixed sleep, used by the tests that need the game to
// advance rather than just to react.
func Wait(s int) {
	time.Sleep(time.Duration(s) * 1000 * time.Millisecond)
}

// Await is await(condition, timeoutSeconds, description): returns as soon as
// the condition holds, so a passing test costs what it takes, not the timeout.
func Await(condition func() bool, timeoutSeconds int, description string) {
	var deadline = time.Now().Add(time.Duration(timeoutSeconds) * 1000 * time.Millisecond)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(POLL_MILLIS * time.Millisecond)
	}
	panic(errors.New("Timed out after " + strconv.Itoa(timeoutSeconds) + "s waiting for " + description))
}

// AwaitSize is awaitSize(collection, size, timeoutSeconds, description). Go has
// no Collection to hold a reference to, so the caller passes a closure that
// reports the current size; the failure message reports it the way Java's does.
func AwaitSize(count func() int, size int, timeoutSeconds int, description string) {
	Await(func() bool { return count() >= size }, timeoutSeconds,
		description+" ("+strconv.Itoa(size)+" events, was "+strconv.Itoa(count())+")")
}

// Properties is the java.util.Properties readSecrets() hands back.
type Properties struct {
	values map[string]string
}

// GetProperty is getProperty(key): the empty string for a key the file does not
// have, which is what Java's null becomes in this port.
func (p *Properties) GetProperty(key string) string {
	var value, ok = p.values[key]
	if !ok {
		return ""
	}
	return value
}

// ReadSecrets is readSecrets(): the admin_password the integration ladder logs
// in with. Java hands ./openttd/secrets.cfg to Properties.load, which splits a
// line at its first '=', ':' or whitespace and ignores '#' and '!' comments; a
// line with no separator becomes a key with an empty value, which is where the
// '[secrets]' and '[network]' headers land. The port keeps those rules and skips
// the two Java features secrets.cfg never uses: a backslash line continuation
// and \uXXXX escapes in the value.
//
// Java reads the path relative to the repo root, because that is where Gradle
// runs. `go test` runs with the tested package's own directory as the working
// directory, so the candidates below walk up from src/go/admin and src/go/server
// as well as from the repo root itself.
func ReadSecrets() (*Properties, error) {
	var candidates []string = []string{
		"openttd/secrets.cfg",
		"../openttd/secrets.cfg",
		"../../openttd/secrets.cfg",
		"../../../openttd/secrets.cfg",
	}
	for _, path := range candidates {
		var text, err = os.ReadFile(path)
		if err != nil {
			continue
		}
		return parseProperties(string(text)), nil
	}
	return nil, errors.New("no openttd/secrets.cfg found in ./, ../ or ../../")
}

// parseProperties is Properties.load(): one line, one key and value.
func parseProperties(text string) *Properties {
	var properties = &Properties{values: map[string]string{}}
	for _, raw := range strings.Split(text, "\n") {
		var line = strings.ReplaceAll(raw, "\r", "")
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, "!") {
			continue
		}
		var key, value = splitProperty(line)
		properties.values[key] = value
	}
	return properties
}

// splitProperty is PropertyKeyReader.convertSingleLineFromString: the key ends
// at the first separator or whitespace, an optional '=' or ':' follows it, and
// the value is what follows that, both trimmed. The separator is consumed, not
// kept: secrets.cfg writes `admin_password = test-password`, and Java's reader
// skips the whitespace around the '=', so the value is the password itself.
// Keeping it would send `= test-password` as the admin password, which the
// server refuses with NETWORK_ERROR_WRONG_PASSWORD.
func splitProperty(line string) (string, string) {
	var index = 0
	for (index < len(line)) && (line[index] != byte('=') && line[index] != byte(':') &&
		line[index] != byte(' ') && line[index] != byte('\t')) {
		index++
	}
	if index == len(line) {
		return strings.TrimSpace(line), ""
	}
	var key = strings.TrimSpace(line[0:index])
	// Whitespace ends the key, then the separator, then the value's whitespace.
	for (index < len(line)) && (line[index] == byte(' ') || line[index] == byte('\t')) {
		index++
	}
	if index < len(line) && (line[index] == byte('=') || line[index] == byte(':')) {
		index++
	}
	for (index < len(line)) && (line[index] == byte(' ') || line[index] == byte('\t')) {
		index++
	}
	return key, strings.TrimSpace(line[index:len(line)])
}
