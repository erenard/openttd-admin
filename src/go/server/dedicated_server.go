// DedicatedServer is com.openttd.server.OpenttdServerWrapper: the admin-port
// probe the integration ladder uses, plus the dedicated server the Java wrapper
// spawns when 3977 turns out to be free.
//
// The Go ladder does not spawn a server: `go test` has no @BeforeAll/@AfterAll
// pair to own the process for the whole run, and the Java notes already say
// pre-starting the server is the reliable option. The integration tests ask
// CheckAdminPort() and skip when nothing listens, so this file's job is to keep
// the wrapper's behaviour ported and available: CheckAdminPort is the question
// they ask, AwaitAdminPort the wait the Java startServer() does after spawning.
//
// There is no Java test for this class, so it has no Go test either: the probe
// is exercised by the integration ladder, which needs a real server.

package server

import "bufio"
import "net"
import "os"
import "strconv"
import "strings"
import "sync"
import "time"
import "com.openttd"

// ADMIN_PORT is the admin port the tests connect to, see openttd/openttd.cfg.
const ADMIN_PORT = 3977

// CONNECT_TIMEOUT_MILLIS bounds one probe of the port, the way Java's
// `new Socket("localhost", ADMIN_PORT)` gives up on a connection that never
// completes.
const CONNECT_TIMEOUT_MILLIS = 1000

// DESTROY_TIMEOUT_MILLIS is how long Destroy() waits for the console reader to
// notice the killed server, standing in for Java's consoleStream.join().
const DESTROY_TIMEOUT_MILLIS = 5000

// DESTROY_POLL_INTERVAL_MILLIS is how often that wait re-tests the flag.
const DESTROY_POLL_INTERVAL_MILLIS = 10

// DedicatedServer is the inner OpenttdDedicatedServer: the process, and the two
// flags its reader loop and Destroy() share.
type DedicatedServer struct {
	process *os.Process
	// stateLock guards the two flags below.
	stateLock sync.Mutex
	running   bool
	// stopped is the port's stand-in for the reader thread having joined.
	stopped   bool
}

// NewDedicatedServer is the inner class's constructor: `openttd -D -d
// 9,console=1,sprite=0,grf=0,sl=0 -c openttd/openttd.cfg`, stderr merged into
// stdout and echoed line by line, exactly as the Java wrapper starts it.
//
// Java's constructor takes nothing because Gradle runs from the repo root, which
// is OpenTTD's search path; `go test` runs from the tested package's directory,
// so the search path is an argument here instead.
func NewDedicatedServer(searchPath string) (*DedicatedServer, error) {
	var server = &DedicatedServer{running: true}
	var commandLine = []string{
		"openttd",
		"-D",
		"-d",
		"9,console=1,sprite=0,grf=0,sl=0",
		"-c",
		"openttd/openttd.cfg",
	}
	var outputRead, outputWrite, pipeError = os.Pipe()
	if pipeError != nil {
		return nil, pipeError
	}
	// No stdin: the wrapper never types into the server, it only reads it.
	// The same pipe for stdout and stderr is redirectErrorStream(true).
	var files = []*os.File{outputRead, outputWrite, outputWrite}
	var process, startError = os.StartProcess(commandLine[0], commandLine,
		&os.ProcAttr{Dir: searchPath, Files: files})
	if startError != nil {
		return nil, startError
	}
	server.process = process
	go ReadConsole(server, outputRead)
	go AwaitExit(process)
	// The parent's copy of the write end has to go, or the reader never sees
	// the end of the console output.
	outputWrite.Close()
	return server, nil
}

// ReadConsole is the consoleStream thread: print the server's lines until the
// wrapper stops wanting them. Java prints them to System.out; the port writes
// them to stderr, which is where `go test` shows them too.
//
// ReadString('\n') rather than ReadLine(): Java's BufferedReader.readLine() hands
// back a whole line, and ReadLine would report a long line as a prefix and let the
// rest of it start the next iteration.
func ReadConsole(server *DedicatedServer, output *os.File) {
	var reader = bufio.NewReader(output)
	for (server.IsRunning()) {
		var line, err = reader.ReadString('\n')
		if line != "" {
			var text = strings.ReplaceAll(line, "\r", "")
			if text == "\n" {
				continue
			}
			os.Stderr.WriteString(text)
		}
		if err != nil {
			break
		}
	}
	server.setStopped(true)
}

// AwaitExit reaps the child, the way the launcher's waiter goroutine does:
// os.Process.Wait may only be called once, so it happens here rather than in
// Destroy().
func AwaitExit(process *os.Process) {
	process.Wait()
}

// Destroy is destroy(): stop reading the console and take the server down.
//
// Java joins the reader thread, closes the readers, then calls destroy()
// (SIGTERM) and destroyForcibly() (SIGKILL). This toolchain has no signal
// package, so only the SIGKILL half exists here: the port kills the process and
// then waits for the reader to see the closed stream, which is the same end
// state.
func (s *DedicatedServer) Destroy() {
	s.setRunning(false)
	if s.process != nil {
		s.process.Kill()
	}
	var waited = 0
	for (waited < DESTROY_TIMEOUT_MILLIS) && (!s.IsStopped()) {
		time.Sleep(DESTROY_POLL_INTERVAL_MILLIS * time.Millisecond)
		waited = waited + DESTROY_POLL_INTERVAL_MILLIS
	}
}

// IsRunning is the reader loop's flag.
func (s *DedicatedServer) IsRunning() bool {
	s.stateLock.Lock()
	var running = s.running
	s.stateLock.Unlock()
	return running
}

// IsStopped reports that the console reader has finished.
func (s *DedicatedServer) IsStopped() bool {
	s.stateLock.Lock()
	var stopped = s.stopped
	s.stateLock.Unlock()
	return stopped
}

func (s *DedicatedServer) setRunning(running bool) {
	s.stateLock.Lock()
	s.running = running
	s.stateLock.Unlock()
}

func (s *DedicatedServer) setStopped(stopped bool) {
	s.stateLock.Lock()
	s.stopped = stopped
	s.stateLock.Unlock()
}

// CheckAdminPort is checkAdminPort(): true when something already listens on
// ADMIN_PORT. Java probes by opening a ServerSocket, so the port being free is
// the success path and returns false; the port being taken is the IOException
// and returns true.
func CheckAdminPort() bool {
	var listener, err = net.Listen("tcp", "localhost:"+strconv.Itoa(ADMIN_PORT))
	if err != nil {
		return true
	}
	listener.Close()
	return false
}

// CanConnectAdminPort is the awaited condition on its own: the TCP handshake
// Java's `new Socket("localhost", ADMIN_PORT)` performs.
func CanConnectAdminPort() bool {
	var conn, err = net.DialTimeout("tcp", "localhost:"+strconv.Itoa(ADMIN_PORT),
		time.Duration(CONNECT_TIMEOUT_MILLIS)*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// AwaitAdminPort is awaitAdminPort(): wait for the port to accept a connection.
// A server that was just spawned needs a few seconds to generate its map, and
// NetworkAdmin only retries every 5 s, so joining before the listener is up
// makes the first test fail with "Connection refused".
func AwaitAdminPort(timeoutSeconds int) {
	openttd.Await(func() bool { return CanConnectAdminPort() }, timeoutSeconds,
		"an admin listener on port "+strconv.Itoa(ADMIN_PORT))
}