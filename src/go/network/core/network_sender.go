// NetworkSender mirrors com.openttd.network.core.NetworkSender: the sender
// thread every Network* class has. Java makes it abstract, with the subclass
// supplying isQuitPacket() and debugPacket(); this port has no inheritance, so
// those two arrive as function fields set by the constructor instead, and the
// subclass (NetworkAdminSender) keeps one of these and delegates startup() and
// shutdown() to it.
//
// The queue is Java's unbounded LinkedBlockingQueue<Packet>, and Poll is its
// poll(5, TimeUnit.SECONDS): it waits for the next packet and gives up after
// 5 s, which is what lets the thread notice a dead socket. Java's shutdown()
// interrupts that wait; Go has no interrupt, so Close() ends it instead.
package core

import "sync"
import "time"
import "com.openttd/util"

var senderLog = util.GetLogger("NetworkSender")

// SEND_TIMEOUT_MILLIS is Java's queue.poll(5, TimeUnit.SECONDS).
const SEND_TIMEOUT_MILLIS = 5000

// SEND_DELAY_MILLIS is the Thread.sleep(100) between two packets that are not
// the quit one: the admin port is a low rate protocol, and OpenTTD reads one
// packet per game tick.
const SEND_DELAY_MILLIS = 100

// POLL_INTERVAL_MILLIS is how often Poll looks again. Java blocks in the queue
// instead; polling every 10 ms costs nothing measurable and keeps the port
// free of condition variables, which this toolchain's sync module has none of.
const POLL_INTERVAL_MILLIS = 10

type PacketQueue struct {
	items  []*Packet
	lock   sync.Mutex
	closed bool
}

// NewPacketQueue is new LinkedBlockingQueue<Packet>().
func NewPacketQueue() *PacketQueue {
	return &PacketQueue{lock: sync.Mutex{}}
}

// Offer is queue.offer(packet): never blocks, the queue is unbounded.
func (q *PacketQueue) Offer(packet *Packet) {
	q.lock.Lock()
	q.items = append(q.items, packet)
	q.lock.Unlock()
}

// Poll is queue.poll(timeout): the oldest packet, or nil when the wait ran out
// or the queue was closed by NetworkSender.shutdown().
func (q *PacketQueue) Poll(timeoutMillis int) *Packet {
	var deadline = time.Now().Add(time.Duration(timeoutMillis) * time.Millisecond)
	for true {
		q.lock.Lock()
		if len(q.items) > 0 {
			var packet = q.items[0]
			var rest []*Packet
			for i := 1; i < len(q.items); i++ {
				rest = append(rest, q.items[i])
			}
			q.items = rest
			q.lock.Unlock()
			return packet
		}
		var closed = q.closed
		q.lock.Unlock()
		if closed || !time.Now().Before(deadline) {
			return nil
		}
		time.Sleep(POLL_INTERVAL_MILLIS * time.Millisecond)
	}
	return nil
}

// Size is queue.size(), which the tests use to see that nothing was queued.
func (q *PacketQueue) Size() int {
	q.lock.Lock()
	var size = len(q.items)
	q.lock.Unlock()
	return size
}

// Close is the port's stand-in for Thread.interrupt(): a Poll waiting on this
// queue returns nil, so the run loop goes back to its condition.
func (q *PacketQueue) Close() {
	q.lock.Lock()
	q.closed = true
	q.lock.Unlock()
}

type NetworkSender struct {
	socket *Socket
	// Queue is the field the subclass offers packets to.
	Queue *PacketQueue
	// stateLock guards the two flags below, which Run reads on the sender thread
	// while startup()/shutdown() write them on the caller's. Java leaves them
	// unsynchronized; Go's memory model makes that a race worth locking.
	stateLock sync.Mutex
	running   bool
	// done is what Join waits for: Run's loop has exited, i.e. Java's
	// !thread.isAlive().
	done         bool
	isQuitPacket func(packet *Packet) bool
	debugPacket  func(packet *Packet)
}

// NewNetworkSender is NetworkSender(Socket socket) plus the two abstract
// methods, which the subclass passes in.
func NewNetworkSender(socket *Socket, isQuitPacket func(packet *Packet) bool, debugPacket func(packet *Packet)) *NetworkSender {
	return &NetworkSender{
		socket:       socket,
		Queue:        NewPacketQueue(),
		stateLock:    sync.Mutex{},
		isQuitPacket: isQuitPacket,
		debugPacket:  debugPacket,
	}
}

// Startup is startup(): running = true, then the loop on its own thread. Java
// names that thread "OpenttdPacketSender".
func (s *NetworkSender) Startup() {
	s.stateLock.Lock()
	s.running = true
	s.stateLock.Unlock()
	go Run(s)
}

// Shutdown is shutdown(): stop running and release the waiting Poll.
func (s *NetworkSender) Shutdown() {
	s.stateLock.Lock()
	s.running = false
	s.stateLock.Unlock()
	s.Queue.Close()
}

// IsRunning is the running flag, which the subclass's loop condition and an
// embedding app's reconnect logic read.
func (s *NetworkSender) IsRunning() bool {
	s.stateLock.Lock()
	var running = s.running
	s.stateLock.Unlock()
	return running
}

// AwaitStopped is shutdown()'s join(timeoutMillis): wait for the sender thread
// to stop. Java overloads join(long) with the subclass's join(Configuration),
// which sends the JOIN packet; Go has no overloading, so the two get distinct
// names here and in NetworkAdminSender. Java's join() waits forever; the timeout
// the caller passes is what keeps a sender stuck in a socket write from hanging
// the shutdown.
func (s *NetworkSender) AwaitStopped(timeoutMillis int) {
	var deadline = time.Now().Add(time.Duration(timeoutMillis) * time.Millisecond)
	for !s.IsDone() && time.Now().Before(deadline) {
		time.Sleep(POLL_INTERVAL_MILLIS * time.Millisecond)
	}
}

// IsDone is the port's stand-in for !thread.isAlive().
func (s *NetworkSender) IsDone() bool {
	s.stateLock.Lock()
	var done = s.done
	s.stateLock.Unlock()
	return done
}

// GetSocket is the socket field: NetworkAdmin closes it when the sender stops.
func (s *NetworkSender) GetSocket() *Socket { return s.socket }

// Run is run(): while the admin is running and its socket is open, take the
// next packet, send it, log it when debug logging is on, and stop after the
// quit packet. A Poll that times out yields null, which is how the loop
// re-checks the socket.
func Run(s *NetworkSender) {
	for s.IsRunning() && s.socket != nil && s.socket.IsOpen() {
		var packet = s.Queue.Poll(SEND_TIMEOUT_MILLIS)
		if packet != nil && s.socket != nil {
			s.socket.Send(packet)
			if senderLog.IsDebugEnabled() {
				s.debugPacket(packet)
			}
			// Clean shutdown: ADMIN_PACKET_ADMIN_QUIT is the last packet.
			if s.isQuitPacket(packet) {
				s.stateLock.Lock()
				s.running = false
				s.stateLock.Unlock()
			} else {
				time.Sleep(SEND_DELAY_MILLIS * time.Millisecond)
			}
		}
	}
	s.stateLock.Lock()
	s.done = true
	s.stateLock.Unlock()
}
