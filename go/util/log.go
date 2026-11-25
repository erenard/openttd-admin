// Log is the port's stand-in for slf4j's Logger. The Java code logs through
// LoggerFactory.getLogger(Class) and the embedding apps configure the backend
// in log4j.xml; every call site passes an already concatenated message, so the
// shim only needs levels and a name prefix. Output goes to stderr, like log4j's
// default in the demos.

package util

import "fmt"
import "os"

type Level int

const (
	TRACE Level = 0
	DEBUG Level = 1
	INFO Level = 2
	WARN Level = 3
	ERROR Level = 4
	OFF Level = 5
)

// threshold is the most verbose level that gets printed. The Java apps set the
// equivalent in log4j.xml; INFO matches what the demos and tests show.
var threshold = INFO

func SetLevel(level Level) {
	threshold = level
}

func GetThreshold() Level {
	return threshold
}

type Logger struct {
	name string
}

func GetLogger(name string) *Logger {
	return &Logger{name: name}
}

func (l *Logger) IsTraceEnabled() bool {
	return threshold <= TRACE
}

func (l *Logger) Trace(message string) {
	if threshold <= TRACE {
		l.write("TRACE", message)
	}
}

func (l *Logger) Debug(message string) {
	if threshold <= DEBUG {
		l.write("DEBUG", message)
	}
}

func (l *Logger) Info(message string) {
	if threshold <= INFO {
		l.write("INFO", message)
	}
}

func (l *Logger) Warn(message string) {
	if threshold <= WARN {
		l.write("WARN", message)
	}
}

func (l *Logger) Error(message string) {
	if threshold <= ERROR {
		l.write("ERROR", message)
	}
}

func (l *Logger) write(level string, message string) {
	fmt.Fprintln(os.Stderr, level+" "+l.name+": "+message)
}
