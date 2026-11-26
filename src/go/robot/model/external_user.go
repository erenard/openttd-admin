// ExternalUser is com.openttd.robot.model.ExternalUser: who the application
// says a chat token or a client belongs to. See model/GamePlayer.java for the
// companion record.

package model

import "strconv"

type ExternalUser struct {
	id    int64
	name  string
	admin bool
}

// NewExternalUser is the implicit no-argument constructor; Java's setters fill
// the rest, so the port starts from the empty user.
func NewExternalUser() *ExternalUser {
	return &ExternalUser{}
}

// GetId is getId.
func (u *ExternalUser) GetId() int64 { return u.id }

// SetId is setId.
func (u *ExternalUser) SetId(id int64) { u.id = id }

// GetName is getName.
func (u *ExternalUser) GetName() string { return u.name }

// SetName is setName.
func (u *ExternalUser) SetName(name string) { u.name = name }

// IsAdmin is isAdmin.
func (u *ExternalUser) IsAdmin() bool { return u.admin }

// SetAdmin is setAdmin.
func (u *ExternalUser) SetAdmin(admin bool) { u.admin = admin }

// Equals is equals: two users are the same user when their ids match, whatever
// their names say. Java's hashCode() agrees with it, and the port's maps key on
// the id directly instead.
func (u *ExternalUser) Equals(other *ExternalUser) bool {
	if other == nil {
		return false
	}
	return u.id == other.id
}

// ToString is what the rules log around a user. Java has no toString() here, so
// string concatenation would show the identity hash; the port spells out the id
// and name instead, which is the useful part of that log line.
func (u *ExternalUser) ToString() string {
	return strconv.FormatInt(u.id, 10) + ":" + u.name
}
