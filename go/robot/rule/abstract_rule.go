// rule mirrors com.openttd.robot.rule: the objectives the strategyboard app
// installs on top of the admin facade. Each rule is a listener that turns the
// functional events the facade dispatches into rcon commands, chat messages and
// gamescript calls, and keeps its own small state next to the game model.
//
// There is no Java test for any of these classes, so the *_test.go files in this
// package are new coverage written against the Java source as the specification,
// not ports of existing tests.
//
// AbstractRule is com.openttd.robot.rule.AbstractRule. Java's register() runs in
// its constructor and calls the abstract listEventTypes() virtually; the port
// passes the kinds and the owning rule to NewAbstractRule, which registers them
// the same way. Java's listener map is keyed by Class objects; the port keys it
// by event.EventKind, so listEventTypes() becomes a list of kinds.

package rule

import "com.openttd/admin"
import "com.openttd/admin/model"
import netadmin "com.openttd/network/admin"
import "com.openttd/admin/event"
import "com.openttd/util"

var ruleLog = util.GetLogger("AbstractRule")

// AbstractRule is AbstractRule: the facade a rule talks to, the listener it
// registered itself as, and the kinds it registered for.
type AbstractRule struct {
	openttd *admin.OpenttdAdmin
	self    event.EventListener
	kinds   []event.EventKind
}

// NewAbstractRule is AbstractRule(OpenttdAdmin openttdAdmin) followed by the
// register() call that constructor makes. self is the rule that owns this base:
// Java passes `this`, and the port has to be told, because the base is a field
// of the rule rather than its superclass.
func NewAbstractRule(openttd *admin.OpenttdAdmin, self event.EventListener, kinds []event.EventKind) *AbstractRule {
	var base = &AbstractRule{openttd: openttd, self: self, kinds: kinds}
	base.Register()
	return base
}

// ListEventKinds is listEventTypes(): the kinds this rule listens for. Java
// declares it abstract; the port just reads back what the constructor was given.
func (b *AbstractRule) ListEventKinds() []event.EventKind {
	return b.kinds
}

// Register is register(): addListener(eventClass, this) for every kind.
func (b *AbstractRule) Register() {
	for _, kind := range b.kinds {
		b.openttd.AddListener(kind, b.self)
	}
}

// Unregister is unregister(): removeListener(eventClass, this) for every kind.
//
// DEVIATION FROM JAVA, WHICH IS ALSO A FIX: Java keeps no list of what it
// registered, so unregister() calls listEventTypes() again; that is equivalent
// here, and both remove every kind the rule listens for.
func (b *AbstractRule) Unregister() {
	for _, kind := range b.kinds {
		b.openttd.RemoveListener(kind, b.self)
	}
}

// GetSend is getSend(): the admin sender the rules send their rcon and chat
// commands through.
func (b *AbstractRule) GetSend() *netadmin.NetworkAdminSender {
	return b.openttd.GetSend()
}

// GetOpenttd is the port's accessor for Java's protected openttdAdmin field: a
// rule that needs the facade itself, not just the sender, goes through this.
func (b *AbstractRule) GetOpenttd() *admin.OpenttdAdmin {
	return b.openttd
}

// GetGame is openttdAdmin.getGame(): the game model the rules read.
func (b *AbstractRule) GetGame() *model.Game {
	return b.openttd.GetGame()
}
