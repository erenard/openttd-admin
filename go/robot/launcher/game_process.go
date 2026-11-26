// GameProcess is com.openttd.robot.launcher.GameProcess: the thread that owns a
// dedicated server it started itself, routes its output to the log, and shuts it
// down by typing "quit" into its console.
//
// Java uses ProcessBuilder plus Process.getInputStream(); the port uses os
// .StartProcess with two os.Pipe()s, which is the same thing: the child's stdout
// and stderr are merged into one stream the reader loop reads line by line, and
// its stdin is the stream destroy() writes "quit\r\n" to.
//
// There is no Java test for this class, so game_process_test.go is new coverage
// written against this source as the specification.

package launcher

import "io"
import "bufio"
import "os"
import "strconv"
import "strings"
import "sync"
import "time"
import "com.openttd/network/core"
import "com.openttd/util"

var gameProcessLog = util.GetLogger("GameProcess")

// WAIT_TIMEOUT_MILLIS is the t.join(5000) destroy() gives its waitForTimeout
// thread: how long to wait for the server to shut down after "quit".
const WAIT_TIMEOUT_MILLIS = 5000

// WAIT_POLL_INTERVAL_MILLIS is how often that wait re-tests the exit flag.
const WAIT_POLL_INTERVAL_MILLIS = 10

// GameProcess is the Java class: the options it was built from, the process it
// started, the stream it writes the quit command to, and the two flags the
// reader loop and destroy() share.
type GameProcess struct {
	options *GameProcessOptions
	// process is the field destroy() sets to null; nil stands for Java's null.
	process *os.Process
	// input is the write end of the child's stdin, the stream destroy() writes
	// "quit\r\n" to. Java reaches it through process.getOutputStream().
	input *os.File
	// stateLock guards the three flags below, which the reader and waiter
	// threads write while Create() and Destroy() read them.
	stateLock sync.Mutex
	running   bool
	// exited is the port's stand-in for the process being gone: AwaitExit has
	// returned, which is what Java's waitFor() returning means to destroy().
	exited    bool
	exitValue int
}

// NewGameProcess is GameProcess(GameProcessOptions options) plus the Thread
// constructor's name. Java logs before building the ProcessBuilder; the port
// does the same and defers the os.StartProcess call to Create(), which is where
// Java's processBuilder.start() happens.
func NewGameProcess(options *GameProcessOptions) *GameProcess {
	gameProcessLog.Info("Preparing new GameProcess.")
	return &GameProcess{
		options:   options,
		process:   nil,
		input:     nil,
		stateLock: sync.Mutex{},
		running:   false,
		exited:    false,
		exitValue: 0}
}

// Create is create(): running = true, then start the process and the thread that
// reads its output. Java declares `throws IOException`; the port returns the
// error instead, which is what the caller of a launcher wants anyway.
func (p *GameProcess) Create() (*os.Process, error) {
	p.setRunning(true)
	var commandLine = p.options.GetOptions()
	var stdinPipeRead, stdinPipeWrite, stdinError = os.Pipe()
	if stdinError != nil {
		gameProcessLog.Error("GameProcess failed: " + core.Message(stdinError))
		return nil, stdinError
	}
	var stdoutPipeRead, stdoutPipeWrite, stdoutError = os.Pipe()
	if stdoutError != nil {
		gameProcessLog.Error("GameProcess failed: " + core.Message(stdoutError))
		return nil, stdoutError
	}
	// The child gets the read end of its stdin and one pipe for stdout and stderr
	// both, which is what redirectErrorStream(true) does in Java.
	var files = []*os.File{stdinPipeRead, stdoutPipeWrite, stdoutPipeWrite}
	gameProcessLog.Info("Starting GameProcess...")
	var process, startError = os.StartProcess(commandLine[0], commandLine,
		&os.ProcAttr{Dir: p.options.GetOpenttdPath(), Files: files})
	if startError != nil {
		gameProcessLog.Error("GameProcess failed: " + core.Message(startError))
		return nil, startError
	}
	// The parent's copies of the child's own ends have to go: while the parent
	// holds the write end open, the reader never sees the end of the output, and
	// while it holds the read end open, the child never sees an end of input.
	var closeWrite = stdoutPipeWrite.Close()
	if closeWrite != nil {
		gameProcessLog.Warn("GameProcess failed: " + core.Message(closeWrite))
	}
	var closeRead = stdinPipeRead.Close()
	if closeRead != nil {
		gameProcessLog.Warn("GameProcess failed: " + core.Message(closeRead))
	}
	p.setProcess(process, stdinPipeWrite)
	go Run(p, stdoutPipeRead)
	go AwaitExit(p, process)
	return process, nil
}

// Destroy is destroy(): stop running, type quit into the console, wait for the
// server to leave, and kill it when it did not.
func (p *GameProcess) Destroy() {
	p.setRunning(false)
	var process = p.GetProcess()
	if process == nil {
		gameProcessLog.Info("GameProcess already stopped.")
		return
	}
	if p.writeQuit() {
		// Java logs this before the write, so a failed write still gets the line.
		gameProcessLog.Info("Quiting GameProcess...")
		if p.awaitExit(WAIT_TIMEOUT_MILLIS) {
			gameProcessLog.Info("GameProcess stopped.")
		}
	} else {
		gameProcessLog.Info("Quiting GameProcess failed.")
	}
	if !p.IsExited() {
		// DEVIATION FROM JAVA: destroy() only reaches this branch when writing
		// quit threw, so a server that ignores quit and outlives the 5 second
		// join is left running. The port kills it, which is what the branch
		// exists for.
		gameProcessLog.Info("Killing GameProcess...")
		var killError = process.Kill()
		if killError != nil {
			gameProcessLog.Error("GameProcess failed: " + core.Message(killError))
		}
	}
	p.clearProcess()
}

// writeQuit is the OutputStreamWriter of destroy(): the quit command, then the
// CRLF the server's console wants, in one write, which is the flush too.
func (p *GameProcess) writeQuit() bool {
	var input = p.GetInput()
	if input == nil {
		// DEVIATION FROM JAVA: getOutputStream() on a null process throws
		// NullPointerException there; the port reports the failure instead.
		gameProcessLog.Error("GameProcess failed: no console to quit.")
		return false
	}
	var _, writeError = input.Write([]byte("quit\r\n"))
	if writeError != nil {
		gameProcessLog.Error("GameProcess failed: " + core.Message(writeError))
		return false
	}
	return true
}

// awaitExit is the t.join(5000) of destroy(): wait for AwaitExit to report back.
// Java joins the waitForTimeout thread; the port polls the flag it sets, which
// is how every other join(timeout) in this port is written.
func (p *GameProcess) awaitExit(timeoutMillis int) bool {
	var deadline = time.Now().Add(time.Duration(timeoutMillis) * time.Millisecond)
	for !p.IsExited() && time.Now().Before(deadline) {
		time.Sleep(WAIT_POLL_INTERVAL_MILLIS * time.Millisecond)
	}
	return p.IsExited()
}

// Run is run(): the loop that turns the server's output into log lines. Java
// reads it with a BufferedReader over process.getInputStream().
func Run(p *GameProcess, output *os.File) {
	var reader = bufio.NewReader(output)
	for p.IsRunning() {
		var line, more, readError = reader.ReadLine()
		if readError != nil {
			if readError == io.EOF {
				// DEVIATION FROM JAVA: readLine() returns null at the end of
				// the stream, which makes Java's loop spin on it forever. The
				// port ends the loop, and reaching the end of the output of a
				// server that exited is not a failure, so it is logged at info.
				gameProcessLog.Info("GameProcess output ended.")
			} else if p.IsExited() || !p.IsRunning() {
				gameProcessLog.Info("GameProcess output ended: " + core.Message(readError))
			} else {
				gameProcessLog.Error("GameProcess failed: " + core.Message(readError))
				p.stopProcess()
			}
			return
		}
		if len(line) > 0 || more {
			// Java's `if (line != null)`: an empty line is still a line, and
			// the port skips the ones the stream ends with.
			Show(p, string(line))
		}
	}
}

// Show is the three way branch of the loop: the server's own log levels, routed
// to the logger of this class.
func Show(p *GameProcess, line string) {
	if strings.HasPrefix(line, "dbg: ") {
		gameProcessLog.Debug(line)
	} else if strings.HasPrefix(line, "ERROR: ") {
		gameProcessLog.Error(line)
	} else {
		gameProcessLog.Info(line)
	}
}

// AwaitExit is the waitForTimeout Runnable: wait for the child and log the value
// it left with.
//
// DEVIATION FROM JAVA: Java starts that thread from destroy(), so a server that
// exits on its own is never waited for and stays a zombie process. The port starts
// it with the process, so the child is always reaped, and destroy() only has to
// look at the flag it sets.
func AwaitExit(p *GameProcess, process *os.Process) {
	var state, waitError = process.Wait()
	if waitError != nil {
		gameProcessLog.Info("GameProcess terminated: " + core.Message(waitError))
		p.setExited(0)
		return
	}
	p.setExited(state.ExitCode())
	gameProcessLog.Info("GameProcess terminated with value: " + strconv.Itoa(state.ExitCode()))
}

// stopProcess is the process.destroy() of run()'s catch block.
func (p *GameProcess) stopProcess() {
	var process = p.GetProcess()
	if process == nil {
		return
	}
	gameProcessLog.Info("Stopping GameProcess...")
	var killError = process.Kill()
	if killError != nil {
		gameProcessLog.Error("GameProcess failed: " + core.Message(killError))
	}
	p.clearProcess()
}

// IsRunning is the running flag the loop condition reads.
func (p *GameProcess) IsRunning() bool {
	p.stateLock.Lock()
	var running = p.running
	p.stateLock.Unlock()
	return running
}

// IsExited is the flag AwaitExit sets when the child is gone.
func (p *GameProcess) IsExited() bool {
	p.stateLock.Lock()
	var exited = p.exited
	p.stateLock.Unlock()
	return exited
}

// GetExitValue is the value the process left with, which waitForTimeout only
// logs in Java and the port keeps for the caller.
func (p *GameProcess) GetExitValue() int {
	p.stateLock.Lock()
	var exitValue = p.exitValue
	p.stateLock.Unlock()
	return exitValue
}

// GetProcess is the process field, which is null once the process is gone.
func (p *GameProcess) GetProcess() *os.Process {
	p.stateLock.Lock()
	var process = p.process
	p.stateLock.Unlock()
	return process
}

// GetInput is the console stream, which destroy() writes to.
func (p *GameProcess) GetInput() *os.File {
	p.stateLock.Lock()
	var input = p.input
	p.stateLock.Unlock()
	return input
}

// setRunning is the running = true / false of create() and destroy().
func (p *GameProcess) setRunning(running bool) {
	p.stateLock.Lock()
	p.running = running
	p.stateLock.Unlock()
}

// setProcess is the process = ... of create().
func (p *GameProcess) setProcess(process *os.Process, input *os.File) {
	p.stateLock.Lock()
	p.process = process
	p.input = input
	p.stateLock.Unlock()
}

// clearProcess is the process = null of destroy() and of run()'s catch block.
func (p *GameProcess) clearProcess() {
	p.stateLock.Lock()
	p.process = nil
	p.input = nil
	p.stateLock.Unlock()
}

// setExited is what AwaitExit does when the child is gone.
func (p *GameProcess) setExited(exitValue int) {
	p.stateLock.Lock()
	p.exited = true
	p.exitValue = exitValue
	p.stateLock.Unlock()
}
