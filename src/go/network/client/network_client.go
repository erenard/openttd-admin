// NetworkClient mirrors com.openttd.network.client.NetworkClient: the listener
// thread of the game port (3979). It owns a NetworkClientSender, decodes the
// server's packets into a NetworkClientModel and answers the handshake the way
// network_client.cpp does. See network_client.cpp for updates.
package client

import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/util"
import "crypto/md5"
import "encoding/hex"
import "sync"
import "time"

// clientLog is the class logger. Java's NetworkClient and NetworkClientSender
// each have their own; the sender's is the plain "log" of this package, so this
// one is named to keep the two apart.
var clientLog = util.GetLogger("NetworkClient")

// RECONNECT_DELAY_MILLIS is Thread.sleep(5000) between two join attempts.
const RECONNECT_DELAY_MILLIS = 5000

// SLEEP_POLL_MILLIS is how often the reconnect sleep checks IsRunning(), the
// stand-in for Java's thread interrupt.
const SLEEP_POLL_MILLIS = 10

type NetworkClient struct {
	configuration *core.Configuration

	// stateLock guards running, which run() reads on its own thread while
	// shutdown() writes it. Java leaves the plain boolean field unsynchronized.
	stateLock sync.Mutex
	running   bool

	// sendLock guards send, which GetSend() hands to the application while
	// run() replaces it on every reconnect. Java's field is volatile.
	sendLock sync.Mutex
	send     *NetworkClientSender

	// socket is Java's field: the current connection, closed on shutdown() to
	// unblock a receive() that is waiting for data.
	socketLock sync.Mutex
	socket     *core.Socket

	// model is Java's "client" field: the state decoded out of the server's
	// packets, rebuilt for every connection.
	modelLock sync.Mutex
	model     *NetworkClientModel
}

// NewNetworkClient is the constructor: it keeps the configuration. The thread
// is started separately, as Java's start() does.
func NewNetworkClient(configuration *core.Configuration) *NetworkClient {
	return &NetworkClient{
		configuration: configuration,
		stateLock:     sync.Mutex{},
		sendLock:      sync.Mutex{},
		socketLock:    sync.Mutex{},
		modelLock:     sync.Mutex{},
		model:         NewNetworkClientModel(),
	}
}

// Startup is Thread.start(): the listener thread. Java sets the running flag
// inside run(); here it is set before the thread is started, as in the admin
// port, so a shutdown() that races the startup still sees a live client and
// IsConnected() is true as soon as Startup() returns.
func (c *NetworkClient) Startup() {
	c.setRunning(true)
	go c.run()
}

// Shutdown is shutdown: stop the client, quit and shut the sender down, close
// the socket so a blocked receive() returns, then wait for the thread. Java
// leaves the sender unguarded, which is only safe because its caller checks
// isConnected() first; here a never-connected client may be shut down, so the
// sender and the socket are both nil-checked.
func (c *NetworkClient) Shutdown() {
	c.setRunning(false)
	var send = c.GetSend()
	if send != nil {
		send.Quit()
		send.Shutdown()
	}
	// receive() blocks this thread, so the socket has to be closed here.
	var socket = c.getSocket()
	if socket != nil {
		socket.Close()
	}
	c.awaitStopped(RECONNECT_DELAY_MILLIS)
}

// IsRunning is the loop condition of run().
func (c *NetworkClient) IsRunning() bool {
	c.stateLock.Lock()
	var value = c.running
	c.stateLock.Unlock()
	return value
}

func (c *NetworkClient) setRunning(running bool) {
	c.stateLock.Lock()
	c.running = running
	c.stateLock.Unlock()
}

// run is run: reconnect until asked to stop. Unlike NetworkAdmin there is no
// "if (!running) break;" in Java's loop, so a failed join still waits out the
// reconnect delay before trying again. Java's "running = true" is Startup()'s
// job here (see Startup).
func (c *NetworkClient) run() {
	clientLog.Debug("started")
	for c.IsRunning() {
		c.attempt()
		if !c.sleep() {
			break
		}
	}
	clientLog.Debug("stopped")
}

// attempt is one turn of Java's while loop: the try, the receive loop and the
// finally block. Every failure Java catches is logged here and swallowed, so
// the reconnect loop carries on, exactly as the catch clauses do.
func (c *NetworkClient) attempt() {
	clientLog.Debug("connecting...")
	var socket, err = core.NewTcpSocket(c.configuration.Host, c.configuration.ClientPort)
	if err != nil {
		// Java's ConnectException clause logs unconditionally, the IOException
		// one only while running: a refused connection is worth reporting even
		// during a shutdown.
		clientLog.Error(core.Message(err))
		c.disconnect(nil)
		return
	}
	c.setSocket(socket)

	var send = NewNetworkClientSender(socket)
	c.setSend(send)
	send.Startup()
	c.setModel(NewNetworkClientModel())
	send.Join(c.configuration)

	// Java's inner loop is while (this.running). handlePacket never throws here,
	// so the only way out of the loop is the running flag or a failed read.
	for c.IsRunning() {
		var packets, readError = socket.Receive()
		if readError != nil {
			if c.IsRunning() {
				clientLog.Error(core.Message(readError))
			}
			break
		}
		for i := 0; i < len(packets); i++ {
			c.handlePacket(packets[i])
		}
	}

	c.disconnect(socket)
}

// disconnect is the finally block: log, tear the sender down, close the socket.
// Java does not send QUIT here, only shutdown() does.
func (c *NetworkClient) disconnect(socket *core.Socket) {
	clientLog.Debug("disconneting...")
	c.setSocket(nil)
	var send = c.GetSend()
	if send != nil {
		send.Shutdown()
	}
	if socket != nil {
		socket.Close()
	}
}

// sleep is the Thread.sleep(5000) between attempts. Java interrupts it from
// shutdown(); here the flag is checked every SLEEP_POLL_MILLIS ms, so shutdown
// returns promptly instead of waiting out the 5 s.
func (c *NetworkClient) sleep() bool {
	var deadline = time.Now().Add(time.Duration(RECONNECT_DELAY_MILLIS) * time.Millisecond)
	for time.Now().Before(deadline) {
		if !c.IsRunning() {
			return false
		}
		time.Sleep(time.Duration(SLEEP_POLL_MILLIS) * time.Millisecond)
	}
	return true
}

// awaitStopped is Thread.join(5000): wait for the listener thread to leave its
// reconnect loop.
func (c *NetworkClient) awaitStopped(timeoutMillis int) {
	var deadline = time.Now().Add(time.Duration(timeoutMillis) * time.Millisecond)
	for time.Now().Before(deadline) {
		if !c.IsRunning() {
			return
		}
		time.Sleep(time.Duration(SLEEP_POLL_MILLIS) * time.Millisecond)
	}
}

// GetSend is getSend: the speaking thread of the current connection, null until
// one has been made.
func (c *NetworkClient) GetSend() *NetworkClientSender {
	c.sendLock.Lock()
	var value = c.send
	c.sendLock.Unlock()
	return value
}

func (c *NetworkClient) setSend(send *NetworkClientSender) {
	c.sendLock.Lock()
	c.send = send
	c.sendLock.Unlock()
}

func (c *NetworkClient) getSocket() *core.Socket {
	c.socketLock.Lock()
	var value = c.socket
	c.socketLock.Unlock()
	return value
}

func (c *NetworkClient) setSocket(socket *core.Socket) {
	c.socketLock.Lock()
	c.socket = socket
	c.socketLock.Unlock()
}

// GetModel is getClient: the state decoded out of the server's packets.
func (c *NetworkClient) GetModel() *NetworkClientModel {
	c.modelLock.Lock()
	var value = c.model
	c.modelLock.Unlock()
	return value
}

func (c *NetworkClient) setModel(model *NetworkClientModel) {
	c.modelLock.Lock()
	c.model = model
	c.modelLock.Unlock()
}

// handlePacket is handlePacket: decode one of the server's packets. Java
// declares "throws NetworkException" but never throws: every case either reads
// the payload or answers the handshake, so the port returns nothing. The
// locals Java keeps for the unfinished cases are read and dropped here, which
// is what keeps the cursor where the next packet needs it.
func (c *NetworkClient) handlePacket(packet *core.Packet) {
	var packetTypeId = packet.ReadUint8()
	var packetType = constant.PacketGameTypeOf(packetTypeId)
	clientLog.Debug("Rcv " + constant.PacketGameTypeName(packetType))
	var send = c.GetSend()
	var model = c.GetModel()

	switch packetType {
	case constant.PACKET_SERVER_FULL:
		// We try to join a server which is full.
	case constant.PACKET_SERVER_BANNED:
		// We try to join a server where we are banned.
	case constant.PACKET_SERVER_CLIENT_INFO:
		packet.ReadUint32() // clientId
		packet.ReadUint8()  // companyId
		packet.ReadString() // name
	case constant.PACKET_SERVER_ERROR:
		var errorId = packet.ReadUint8()
		clientLog.Error(DescribeError(errorId))
	case constant.PACKET_SERVER_CHECK_NEWGRFS:
		var grfCount = packet.ReadUint8()
		for grfIndex := 0; grfIndex < grfCount; grfIndex++ {
			packet.ReadUint32() // grfId
			packet.ReadUint32() // md5, TODO incomplete as in Java (game_info.cpp:377)
		}
		send.NewGRFsOk()
	case constant.PACKET_SERVER_NEED_GAME_PASSWORD:
		send.GamePassword(c.configuration.Password)
	case constant.PACKET_SERVER_NEED_COMPANY_PASSWORD:
		model.PasswordGameSeed = int64(packet.ReadUint32())
		model.PasswordServerId = packet.ReadString()
		send.CompanyPassword(HashPassword(c.configuration.Password, model.PasswordServerId, model.PasswordGameSeed))
	case constant.PACKET_SERVER_WELCOME:
		model.ClientId = int64(packet.ReadUint32())
		model.PasswordGameSeed = int64(packet.ReadUint32())
		model.PasswordServerId = packet.ReadString()
		send.GetMap()
	case constant.PACKET_SERVER_WAIT:
		packet.ReadUint8() // waiting
	case constant.PACKET_SERVER_MAP_BEGIN:
		packet.ReadUint32() // frame
	case constant.PACKET_SERVER_MAP_SIZE:
		packet.ReadUint32() // mapSize
	case constant.PACKET_SERVER_MAP_DATA:
		// TODO incomplete, as in Java.
	case constant.PACKET_SERVER_MAP_DONE:
		send.MapOk()
	case constant.PACKET_SERVER_FRAME:
		model.FrameCounterServer = int64(packet.ReadUint32())
		packet.ReadUint32() // frame_counter_max
		// Java: position() + 1 < length(), i.e. two bytes still to read.
		if packet.Remaining() > 1 {
			packet.ReadUint32() // sync_seed1
			if packet.Remaining() > 1 {
				packet.ReadUint32() // sync_seed2
			}
		}
		if packet.Remaining() != 0 {
			model.Token = packet.ReadUint8()
		}
		if model.LastAckFrame < model.FrameCounterServer {
			model.LastAckFrame = model.FrameCounterServer + DAY_TICKS
			send.Ack(model.FrameCounterServer, model.Token)
		}
	case constant.PACKET_SERVER_SYNC:
		packet.ReadUint32() // sync_frame
		packet.ReadUint32() // sync_seed1
		if packet.Remaining() != 0 {
			packet.ReadUint32() // sync_seed2
		}
	case constant.PACKET_SERVER_COMMAND:
		packet.ReadUint32() // cmdFrame
		packet.ReadBool8()  // myCmd
	case constant.PACKET_SERVER_CHAT:
		packet.ReadUint8()  // networkActionId
		packet.ReadUint32() // client_id
		packet.ReadBool8()  // selfSend
		packet.ReadString() // message
		packet.ReadUint64() // data
	case constant.PACKET_SERVER_EXTERNAL_CHAT:
		packet.ReadString() // source
		packet.ReadUint16() // colour
		packet.ReadString() // user
		packet.ReadString() // message
	case constant.PACKET_SERVER_ERROR_QUIT, constant.PACKET_SERVER_QUIT, constant.PACKET_SERVER_JOIN:
		packet.ReadUint32() // clientId
	case constant.PACKET_SERVER_SHUTDOWN, constant.PACKET_SERVER_NEWGAME:
		// Nothing to do, as in Java's empty fall-through group.
	case constant.PACKET_SERVER_RCON:
		packet.ReadUint16() // colour
		packet.ReadString() // rcon
	case constant.PACKET_SERVER_MOVE:
		packet.ReadUint32() // clientId
		packet.ReadUint8()  // companyId
	case constant.PACKET_SERVER_CONFIG_UPDATE:
		packet.ReadUint8()  // maxCompanies
		packet.ReadString() // serverName
	case constant.PACKET_SERVER_COMPANY_UPDATE:
		packet.ReadUint16() // passworded
	}
}

// DescribeError is NetworkErrorCode.valueOf(id).toString(), the way Java logs
// the server's error code.
func DescribeError(errorId int) string {
	return constant.NetworkErrorCodeName(constant.NetworkErrorCodeOf(errorId))
}

// HashPassword is hashPassword, "see network.cpp, method:
// GenerateCompanyPasswordHash": the company password is sent as the MD5 of itself
// salted with the server's id and the game seed.
//
// DEVIATION: Java's version cannot return at all. It writes the hex digest with
// a PrintWriter it never flushes, so the StringWriter is still empty, and the
// following write(new char[]{'\0'}, length - 1, 1) is write(chars, -1, 1): an
// IndexOutOfBoundsException that escapes handlePacket and kills the listener
// thread. The port implements the upstream intent instead; the oracle values
// network_client_test.go's TestHashPassword uses are the output of the Java
// source's own fill and salt loops, run with only the broken PrintWriter part
// replaced (/tmp/opencode/hash/HashRule.java).
func HashPassword(password string, passwordServerId string, passwordGameSeed int64) string {
	// NETWORK_SERVER_ID_LENGTH is passwordServerId.length(). Java counts UTF-16
	// code units; a server id is ASCII, so bytes and code units agree here.
	var saltLength = len(passwordServerId)
	// if (StrEmpty(password)) return password;
	if password == "" {
		return password
	}
	if saltLength == 0 {
		// Java would have thrown: new byte[0] then update(0, -1).
		return ""
	}

	// char salted_password[NETWORK_SERVER_ID_LENGTH]; memset(..., 0, ...) then
	// snprintf(salted_password, ..., "%s", password): the password, padded with
	// spaces to the length of the server id.
	//
	// DEVIATION FROM UPSTREAM, FAITHFUL TO JAVA: Java's if (password.length() <
	// NETWORK_SERVER_ID_LENGTH) sits inside the fill loop, so a password shorter
	// than the server id is replaced by spaces entirely - the password itself never
	// reaches the buffer. Upstream's snprintf truncates to n-1 chars and NUL-pads
	// instead, so a real 15.3 server computes a different hash for that case. The
	// port keeps Java's rule, which is what the oracle values below come from.
	var serverId = []byte(passwordServerId)
	var salted = make([]byte, saltLength)
	var passwordBytes = []byte(password)
	if len(passwordBytes) < saltLength {
		for i := 0; i < saltLength; i++ {
			salted[i] = byte(' ')
		}
	} else {
		for i := 0; i < saltLength; i++ {
			salted[i] = passwordBytes[i]
		}
	}

	// /* Add the game seed and the server's ID as the salt. */ Java's "^" mixes
	// a char and a long, then truncates to a byte, so only the low 8 bits of the
	// shifted seed matter.
	for i := 0; i < saltLength-1; i++ {
		salted[i] = salted[i] ^ serverId[i] ^ byte(passwordGameSeed>>(i%32))
	}

	// Md5 checksum; checksum.update(salted_password, 0, NETWORK_SERVER_ID_LENGTH
	// - 1), then the 16 bytes as "%02x".
	var digest = md5.Sum(salted[0 : saltLength-1])
	return hex.EncodeToString(digest[0:])
}
