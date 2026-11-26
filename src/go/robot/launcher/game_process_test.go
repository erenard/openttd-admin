// Tests for the launcher of the robot layer: GameProcessOptions, the command
// line and working directory, and GameProcess, the process it starts, the log
// lines it reads from it and the way it shuts it down.
//
// There is no Java test for this package - test/java/com/openttd has no launcher
// directory, and the Java class only runs against a real dedicated server in
// the integration suite - so this file is new coverage written against the Go
// source as the specification. The "server" is a shell script that prints the
// three kinds of line a 15.3 dedicated server prints and reads its console,
// which is enough to pin the plumbing: the working directory, the merged output
// stream, the level each line is logged at, the quit command, and the kill that
// the port added because this toolchain has no SIGTERM.

package launcher_test

import "os"
import "strings"
import "testing"
import "com.openttd"
import "com.openttd/util"
import "com.openttd/robot/launcher"

// SCRIPT is the fake dedicated server: the three output lines a real one writes
// while starting up - a debug line, an error line and a plain one - and a loop
// that reads the console and leaves when it is told to quit. The case matches
// *quit* because GameProcess writes "quit\r\n" and the shell keeps the CR.
var SCRIPT = "#!/bin/sh\n" +
	"echo \"dbg: network.engine.tcp_ip4v.address = 127.0.0.1\"\n" +
	"echo \"ERROR: packet not understood\"\n" +
	"echo \"Starting dedicated server version 15.3\"\n" +
	"while read line; do\n" +
	"  echo \"dbg: console: $line\"\n" +
	"  case \"$line\" in\n" +
	"    *quit*) exit 0;;\n" +
	"  esac\n" +
	"done\n"

// SCRIPT_THAT_IGNORES_QUIT is the server destroy() has to kill: it reads its
// console forever and never leaves by itself.
var SCRIPT_THAT_IGNORES_QUIT = "#!/bin/sh\n" +
	"echo \"Starting dedicated server version 15.3\"\n" +
	"while read line; do\n" +
	"  echo \"dbg: console: $line\"\n" +
	"done\n"

// SCRIPT_WRITES_A_MARKER is the server used to prove the working directory: it
// ignores the switches GameProcessOptions adds, writes a file with a relative
// name, and leaves. A dedicated server does the same thing with its scenario.
var SCRIPT_WRITES_A_MARKER = "#!/bin/sh\n" +
	"touch marker.dat\n" +
	"exit 0\n"

// NewTempDir is the directory the process is started in, which is where a
// dedicated server would find its scenario and write its savegames.
func NewTempDir() string {
	var dir, mkdirError = os.MkdirTemp("", "gameprocess")
	if mkdirError != nil {
		panic("cannot make a temporary directory: " + mkdirError.Error())
	}
	return dir
}

// NewServerScript writes one of the two scripts and makes it executable, so
// os.StartProcess can run it as the command line's argv[0].
func NewServerScript(dir string, name string, content string) string {
	var path = dir + "/" + name
	var writeError = os.WriteFile(path, []byte(content), 0o755)
	if writeError != nil {
		panic("cannot write " + path + ": " + writeError.Error())
	}
	return path
}

// LogCapture is the log4j sink of the Java apps: a file the log shim writes to
// while the test reads it back. The previous output and threshold are kept so
// the test can put them back.
type LogCapture struct {
	path              string
	file              *os.File
	previous          *os.File
	previousThreshold util.Level
}

func NewLogCapture(dir string) *LogCapture {
	var capture = &LogCapture{
		path:              dir + "/gameprocess.log",
		file:              nil,
		previous:          util.GetOutput(),
		previousThreshold: util.GetThreshold()}
	var file, openError = os.OpenFile(capture.path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o644)
	if openError != nil {
		panic("cannot open the log: " + openError.Error())
	}
	capture.file = file
	util.SetOutput(file)
	return capture
}

// Close puts the log shim back the way it was.
func (c *LogCapture) Close() {
	c.file.Close()
	util.SetOutput(c.previous)
	util.SetLevel(c.previousThreshold)
}

// Text is everything the process has logged so far.
func (c *LogCapture) Text() string {
	var bytes, readError = os.ReadFile(c.path)
	if readError != nil {
		return ""
	}
	return string(bytes)
}

// HasLine is "did the process log this line", the assertion every test below
// makes about the output of the server.
func (c *LogCapture) HasLine(line string) bool {
	return strings.Contains(c.Text(), line)
}

// NewOptions is the command line of the launcher: the script as the executable,
// and the two switches GameProcessOptions always adds.
func NewOptions(dir string, script string) *launcher.GameProcessOptions {
	var options = launcher.NewGameProcessOptions(dir, script)
	return options
}

// TestOptionsBuildTheCommandLine is GameProcessOptions: the launcher always adds
// -D, which daemonizes the server, and -d 1, which is the debug level that makes
// it print the "dbg: " lines GameProcess routes to the debug log, and addOption
// appends the scenario, the password and the console port the caller passes.
func TestOptionsBuildTheCommandLine(t *testing.T) {
	var options = launcher.NewGameProcessOptions("/opt/openttd", "otddemo")
	var commandLine = options.GetOptions()
	if len(commandLine) != 3 {
		t.Errorf("command line has %d entries: %s", len(commandLine),
			strings.Join(commandLine, " "))
	}
	if commandLine[0] != "otddemo" || commandLine[1] != "-D" || commandLine[2] != "-d 1" {
		t.Errorf("command line is %s", strings.Join(commandLine, " "))
	}
	options.AddOption("-c scenario.dat")
	options.AddOption("-n")
	if len(options.GetOptions()) != 5 {
		t.Errorf("addOption did not append: %s", strings.Join(options.GetOptions(), " "))
	}
	if options.GetOpenttdPath() != "/opt/openttd" {
		t.Errorf("path = %s", options.GetOpenttdPath())
	}
}

// TestCreateRunsInTheOptionsDirectory covers the directory(openttdPath) of the
// ProcessBuilder: the server must start in the directory it was given, which is
// where it reads its scenario and writes its savegames. /bin/touch with a
// relative file name is the smallest way to see that.
func TestCreateRunsInTheOptionsDirectory(t *testing.T) {
	var dir = NewTempDir()
	var script = NewServerScript(dir, "server.sh", SCRIPT_WRITES_A_MARKER)
	var options = launcher.NewGameProcessOptions(dir, script)
	var process = launcher.NewGameProcess(options)
	var started, startError = process.Create()
	if startError != nil {
		t.Errorf("create failed: %s", startError.Error())
	}
	if started == nil {
		t.Errorf("no process was started")
	}
	openttd.Await(func() bool { return process.IsExited() }, 10, "the process to exit")
	if process.GetExitValue() != 0 {
		t.Errorf("the server left with %d", process.GetExitValue())
	}
	var info, statError = os.Stat(dir + "/marker.dat")
	if statError != nil {
		t.Errorf("the process did not work in %s: %s", dir, statError.Error())
	} else if info.Size() != 0 {
		t.Errorf("marker.dat is %d bytes", info.Size())
	}
	if !process.IsRunning() {
		t.Errorf("create() must leave the process running")
	}
	process.Destroy()
}

// TestCreateReportsAProcessThatCannotStart is the IOException of
// ProcessBuilder.start(): a command line that cannot be executed has to reach
// the caller as an error, and the process field stays null so destroy() has
// nothing to quit or kill.
func TestCreateReportsAProcessThatCannotStart(t *testing.T) {
	var dir = NewTempDir()
	var options = launcher.NewGameProcessOptions(dir, "/does/not/exist/openttd")
	var process = launcher.NewGameProcess(options)
	var started, startError = process.Create()
	if startError == nil {
		t.Errorf("starting a missing executable reported no error")
	}
	if started != nil {
		t.Errorf("a process was started anyway")
	}
	if process.GetProcess() != nil {
		t.Errorf("the process field is set")
	}
	process.Destroy()
}

// TestDestroyOnAnUnstartedProcess is destroy() before create(): Java's process
// field is null there, so the only thing that happens is the log line, and the
// port must not write to a console it never opened.
func TestDestroyOnAnUnstartedProcess(t *testing.T) {
	var dir = NewTempDir()
	var capture = NewLogCapture(dir)
	var process = launcher.NewGameProcess(NewOptions(dir, "/bin/cat"))
	process.Destroy()
	if !capture.HasLine("GameProcess already stopped.") {
		t.Errorf("destroy() on an unstarted process logged: %s", capture.Text())
	}
	if process.GetProcess() != nil || process.IsExited() {
		t.Errorf("an unstarted process looks started")
	}
	capture.Close()
}

// TestCreateRoutesTheOutputOfTheServer is run(): the merged stdout and stderr of
// the server, read line by line and logged by level - "dbg: " to the debug log,
// "ERROR: " to the error log, everything else to the info log - which is the
// three way branch of the Java loop. The console line at the end proves the
// stream is the child's console and not just its startup output.
func TestCreateRoutesTheOutputOfTheServer(t *testing.T) {
	var dir = NewTempDir()
	var script = NewServerScript(dir, "server.sh", SCRIPT)
	var capture = NewLogCapture(dir)
	// The launcher passes -d 1 so the server writes its dbg: lines, and the
	// debug log has to be enabled for them to be logged at all.
	util.SetLevel(util.DEBUG)
	var process = launcher.NewGameProcess(NewOptions(dir, script))
	var started, startError = process.Create()
	if startError != nil {
		t.Errorf("create failed: %s", startError.Error())
	}
	if started == nil {
		t.Errorf("no process was started")
	}
	openttd.Await(func() bool {
		return capture.HasLine("INFO GameProcess: Starting dedicated server version 15.3")
	}, 10, "the startup lines")
	// The dbg: line only shows at the debug level, which is what -d 1 buys the
	// launcher, and the ERROR: line is logged as an error, not as info.
	if !capture.HasLine("DEBUG GameProcess: dbg: network.engine.tcp_ip4v.address = 127.0.0.1") {
		t.Errorf("the debug line is not in the log: %s", capture.Text())
	}
	if !capture.HasLine("ERROR GameProcess: ERROR: packet not understood") {
		t.Errorf("the error line is not in the log: %s", capture.Text())
	}
	if !capture.HasLine("INFO GameProcess: Starting dedicated server version 15.3") {
		t.Errorf("the plain line is not in the log: %s", capture.Text())
	}
	// The console of the server is the input stream of the process: what is
	// typed there comes back as one more dbg: line, which is how destroy()
	// will be able to quit it.
	var input = process.GetInput()
	var written, writeError = input.Write([]byte("hello\r\n"))
	if writeError != nil {
		t.Errorf("the console write failed: %s", writeError.Error())
	}
	if written != 7 {
		t.Errorf("the console took %d bytes", written)
	}
	openttd.Await(func() bool {
		return capture.HasLine("dbg: console: hello")
	}, 10, "the server to answer the console")
	process.Destroy()
	capture.Close()
}

// TestDestroyQuitsTheServerThroughItsConsole is destroy(): running = false, the
// quit command typed into the console, and the exit value the server left with
// once the waitForTimeout thread came back. The reader loop stops as soon as
// running is false, exactly as it does in Java, so the echo the server writes
// back is not in the log - what is observable is that the quit reached the
// console, which is that the server left by itself with status 0.
func TestDestroyQuitsTheServerThroughItsConsole(t *testing.T) {
	var dir = NewTempDir()
	var script = NewServerScript(dir, "server.sh", SCRIPT)
	var capture = NewLogCapture(dir)
	util.SetLevel(util.DEBUG)
	var process = launcher.NewGameProcess(NewOptions(dir, script))
	var started, startError = process.Create()
	if startError != nil {
		t.Errorf("create failed: %s", startError.Error())
	}
	if started == nil {
		t.Errorf("no process was started")
	}
	// Let the server reach its console loop first, so the quit command has a
	// reader waiting for it.
	var input = process.GetInput()
	var written, writeError = input.Write([]byte("hello\r\n"))
	if writeError != nil {
		t.Errorf("the console write failed: %s", writeError.Error())
	}
	if written != 7 {
		t.Errorf("the console took %d bytes", written)
	}
	openttd.Await(func() bool {
		return capture.HasLine("dbg: console: hello")
	}, 10, "the server to answer the console")
	process.Destroy()
	openttd.Await(func() bool {
		return capture.HasLine("GameProcess terminated with value: 0")
	}, 10, "the exit value of the server")
	if !capture.HasLine("Quiting GameProcess...") {
		t.Errorf("destroy() did not log the quit: %s", capture.Text())
	}
	if !capture.HasLine("GameProcess stopped.") {
		t.Errorf("the quit did not stop the server: %s", capture.Text())
	}
	if capture.HasLine("Killing GameProcess...") {
		t.Errorf("a server that quit its console was killed anyway: %s", capture.Text())
	}
	if process.GetExitValue() != 0 {
		t.Errorf("the server left with %d", process.GetExitValue())
	}
	if process.GetProcess() != nil {
		t.Errorf("destroy() must clear the process field")
	}
	if process.IsRunning() {
		t.Errorf("destroy() must stop the reader loop")
	}
	capture.Close()
}

// TestDestroyKillsAServerThatIgnoresQuit is the branch the port added: a server
// that outlives the five second join is killed, because this toolchain has no
// signal package, so Java's process.destroy() - SIGTERM - maps to os.Process
// .Kill(), which is SIGKILL.
func TestDestroyKillsAServerThatIgnoresQuit(t *testing.T) {
	var dir = NewTempDir()
	var script = NewServerScript(dir, "server.sh", SCRIPT_THAT_IGNORES_QUIT)
	var capture = NewLogCapture(dir)
	var process = launcher.NewGameProcess(NewOptions(dir, script))
	var started, startError = process.Create()
	if startError != nil {
		t.Errorf("create failed: %s", startError.Error())
	}
	if started == nil {
		t.Errorf("no process was started")
	}
	process.Destroy()
	if !capture.HasLine("Killing GameProcess...") {
		t.Errorf("the server was not killed: %s", capture.Text())
	}
	openttd.Await(func() bool { return process.IsExited() }, 10, "the killed server to be reaped")
	if process.GetProcess() != nil {
		t.Errorf("destroy() must clear the process field")
	}
	capture.Close()
}
