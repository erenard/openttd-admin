// EventDispatcher mirrors com.openttd.admin.event.EventDispatcher: the thread
// that takes dispatched events off a queue and hands them to the listeners
// registered for their class, so a listener's work never runs on the thread
// that decodes packets.
//
// Two Java constructs need a stand-in here. Its listener map is keyed by
// Class, which this dialect has no equivalent of, so the key is EventKind and
// each event reports its own. Its queue is a LinkedBlockingQueue<EventMessage>
// polled with poll(5, TimeUnit.SECONDS) and woken by thread.interrupt(): the
// dialect's sync has only Mutex, and an unbuffered channel cannot be polled
// with a timeout, so EventQueue below is hand-rolled the way core.PacketQueue
// is, with Close() standing in for the interrupt. Java's inner EventMessage pair
// is not needed either: its Class member is what Event.Kind() now answers.

package event

import "sync"
import "time"
import "com.openttd/util"

var eventLog = util.GetLogger("EventDispatcher")

// DISPATCH_TIMEOUT_MILLIS is Java's queue.poll(5, TimeUnit.SECONDS): how long
// the dispatcher waits for the next event before going back to its condition.
const DISPATCH_TIMEOUT_MILLIS = 5000

// DISPATCH_POLL_INTERVAL_MILLIS is how often Poll looks again. Java blocks in
// the queue instead of polling.
const DISPATCH_POLL_INTERVAL_MILLIS = 10

// DISPATCH_JOIN_TIMEOUT_MILLIS is shutdown()'s thread.join(5000).
const DISPATCH_JOIN_TIMEOUT_MILLIS = 5000

type EventQueue struct {
	items []Event
	lock sync.Mutex
	closed bool
}

// NewEventQueue is new LinkedBlockingQueue<EventMessage>().
func NewEventQueue() *EventQueue {
	return &EventQueue{lock: sync.Mutex{}}
}

// Offer is queue.offer(event): never blocks, the queue is unbounded.
func (q *EventQueue) Offer(dispatched Event) {
	q.lock.Lock()
	q.items = append(q.items, dispatched)
	q.lock.Unlock()
}

// Poll is queue.poll(5, TimeUnit.SECONDS): the oldest event, or nil when the
// wait ran out or the queue was closed by shutdown().
func (q *EventQueue) Poll(timeoutMillis int) Event {
	var deadline = time.Now().Add(time.Duration(timeoutMillis) * time.Millisecond)
	for (true) {
		q.lock.Lock()
		if len(q.items) > 0 {
			var dispatched = q.items[0]
			var rest []Event
			for i := 1; i < len(q.items); i++ {
				rest = append(rest, q.items[i])
			}
			q.items = rest
			q.lock.Unlock()
			return dispatched
		}
		var closed = q.closed
		q.lock.Unlock()
		if closed || !time.Now().Before(deadline) {
			return nil
		}
		time.Sleep(DISPATCH_POLL_INTERVAL_MILLIS * time.Millisecond)
	}
	return nil
}

// Clear is queue.clear(): startup() and shutdown() both drop what is queued, so
// events from a previous connection are not replayed.
func (q *EventQueue) Clear() {
	q.lock.Lock()
	q.items = []Event{}
	q.lock.Unlock()
}

// Close is the port's stand-in for Thread.interrupt(): a Poll waiting on this
// queue returns nil, so the dispatcher's loop goes back to its condition.
func (q *EventQueue) Close() {
	q.lock.Lock()
	q.closed = true
	q.lock.Unlock()
}

// Open undoes Close, so a dispatcher started again after a shutdown can wait
// for events again. Java needs no such step: its shutdown interrupts a thread
// that startup() replaces with a new one.
func (q *EventQueue) Open() {
	q.lock.Lock()
	q.closed = false
	q.lock.Unlock()
}

type EventDispatcher struct {
	queue *EventQueue
	// lock is Java's ReentrantReadWriteLock: addListener and removeListener hold
	// its write lock, run() its read lock. One mutex covers both here.
	lock sync.Mutex
	listeners map[EventKind][]EventListener
	// stateLock guards the three flags below, which Run reads on the dispatcher
	// thread while startup() and shutdown() write them on the caller's.
	stateLock sync.Mutex
	running bool
	// started and done together are Java's thread != null && thread.isAlive().
	started bool
	done bool
}

// NewEventDispatcher is EventDispatcher().
func NewEventDispatcher() *EventDispatcher {
	return &EventDispatcher{
		queue: NewEventQueue(),
		lock: sync.Mutex{},
		listeners: map[EventKind][]EventListener{},
		stateLock: sync.Mutex{},
	}
}

// AddListener is addListener(Class, EventListener).
func (d *EventDispatcher) AddListener(kind EventKind, listener EventListener) {
	d.lock.Lock()
	var existing, has = d.listeners[kind]
	if !has {
		existing = []EventListener{}
		d.listeners[kind] = existing
	}
	existing = append(existing, listener)
	d.listeners[kind] = existing
	d.lock.Unlock()
}

// RemoveListener is removeListener(Class, EventListener). Java creates the
// list first when there is none, which is harmless, and removes by equals,
// which for these listeners is identity: the port compares the registered
// values, so a listener has to be removed by the same pointer it was added
// with.
func (d *EventDispatcher) RemoveListener(kind EventKind, listener EventListener) {
	d.lock.Lock()
	var existing, has = d.listeners[kind]
	if !has {
		existing = []EventListener{}
		d.listeners[kind] = existing
	}
	var kept []EventListener
	for i := 0; i < len(existing); i++ {
		if existing[i] != listener {
			kept = append(kept, existing[i])
		}
	}
	d.listeners[kind] = kept
	d.lock.Unlock()
}

// Dispatch is dispatch(Event<L> event): Java reads event.getClass() and queues
// the pair of event and class; the port queues the event alone, since Kind()
// answers the class question.
func (d *EventDispatcher) Dispatch(dispatched Event) {
	d.queue.Offer(dispatched)
}

// Startup is startup(): log, run = true, clear the queue, start the thread.
// Java names that thread after the dispatcher's class.
func (d *EventDispatcher) Startup() {
	eventLog.Debug("startup...")
	d.setRunning(true)
	d.queue.Clear()
	d.queue.Open()
	d.setStarted()
	go Run(d)
	eventLog.Debug("started")
}

// Shutdown is shutdown(): stop running, drop the queued events, and stop the
// thread. run() blocks in Poll(5 s), so it has to be woken to stop at once,
// which is what Close() does.
func (d *EventDispatcher) Shutdown() {
	eventLog.Debug("shutdown...")
	d.setRunning(false)
	d.queue.Clear()
	d.queue.Close()
	d.awaitStopped(DISPATCH_JOIN_TIMEOUT_MILLIS)
	eventLog.Debug("stopped")
}

// IsAlive is isAlive(): the dispatcher thread exists and has not finished.
func (d *EventDispatcher) IsAlive() bool {
	d.stateLock.Lock()
	var alive = d.started && !d.done
	d.stateLock.Unlock()
	return alive
}

// QueueSize is queue.size(), which the tests use to see that nothing is queued.
func (d *EventDispatcher) QueueSize() int {
	d.queue.lock.Lock()
	var size = len(d.queue.items)
	d.queue.lock.Unlock()
	return size
}

// Run is run(): while running, take the next event and notify the listeners
// registered for its kind. A Poll that times out yields nil, which is how the
// loop re-checks the flag.
func Run(d *EventDispatcher) {
	for (d.IsRunning()) {
		var dispatched = d.queue.Poll(DISPATCH_TIMEOUT_MILLIS)
		if dispatched != nil {
			d.notifyListeners(dispatched)
		}
	}
	d.setDone()
}

// notifyListeners is the read-locked half of run(): look the kind up and tell
// every listener registered for it. Java wraps this in a try/catch that logs
// whatever notify() threw, typically a ClassCastException for a listener that
// is not of the right kind; the events log that themselves, where it happens.
func (d *EventDispatcher) notifyListeners(dispatched Event) {
	d.lock.Lock()
	var listeners, has = d.listeners[dispatched.Kind()]
	if has {
		for i := 0; i < len(listeners); i++ {
			dispatched.Notify(listeners[i])
		}
	}
	d.lock.Unlock()
}

func (d *EventDispatcher) IsRunning() bool {
	d.stateLock.Lock()
	var running = d.running
	d.stateLock.Unlock()
	return running
}

func (d *EventDispatcher) setRunning(running bool) {
	d.stateLock.Lock()
	d.running = running
	d.stateLock.Unlock()
}

func (d *EventDispatcher) setStarted() {
	d.stateLock.Lock()
	d.started = true
	d.done = false
	d.stateLock.Unlock()
}

func (d *EventDispatcher) setDone() {
	d.stateLock.Lock()
	d.done = true
	d.stateLock.Unlock()
}

// awaitStopped is shutdown()'s thread.join(5000): wait for the dispatcher thread
// to leave its loop.
func (d *EventDispatcher) awaitStopped(timeoutMillis int) {
	var deadline = time.Now().Add(time.Duration(timeoutMillis) * time.Millisecond)
	for (d.IsAlive() && time.Now().Before(deadline)) {
		time.Sleep(DISPATCH_POLL_INTERVAL_MILLIS * time.Millisecond)
	}
}
