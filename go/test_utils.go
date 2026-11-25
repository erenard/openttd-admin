// TestUtils mirrors com.openttd.TestUtils: the polling helpers the tests use
// instead of sleeping a fixed amount of time. Java's await throws
// AssertionError; the Go port panics with the same message, which fails the
// test that is waiting.
package openttd

import "errors"
import "strconv"
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
	panic(errors.New("Timed out after "+strconv.Itoa(timeoutSeconds)+"s waiting for "+description))
}

// AwaitSize is awaitSize(collection, size, timeoutSeconds, description). Go has
// no Collection to hold a reference to, so the caller passes a closure that
// reports the current size; the failure message reports it the way Java's does.
func AwaitSize(count func() int, size int, timeoutSeconds int, description string) {
	Await(func() bool { return count() >= size }, timeoutSeconds,
		description+" ("+strconv.Itoa(size)+" events, was "+strconv.Itoa(count())+")")
}
