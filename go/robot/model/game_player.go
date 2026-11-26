// GamePlayer is com.openttd.robot.model.GamePlayer: one player's entry on the
// leaderboard the external game service is told about when a game ends. See
// model/ExternalUser.java for the companion record.

package model

import "strconv"
import "com.openttd/util"

type GamePlayer struct {
	externalUser   *ExternalUser
	accomplishment float64
	duration       float64
	winner         bool
	companyId      int
	score          int
}

// NewGamePlayer is the implicit no-argument constructor.
func NewGamePlayer() *GamePlayer {
	return &GamePlayer{}
}

// GetExternalUser is getExternalUser.
func (p *GamePlayer) GetExternalUser() *ExternalUser { return p.externalUser }

// SetExternalUser is setExternalUser.
func (p *GamePlayer) SetExternalUser(user *ExternalUser) { p.externalUser = user }

// GetAccomplishment is getAccomplishment.
func (p *GamePlayer) GetAccomplishment() float64 { return p.accomplishment }

// SetAccomplishment is setAccomplishment.
func (p *GamePlayer) SetAccomplishment(accomplishment float64) { p.accomplishment = accomplishment }

// GetDuration is getDuration.
func (p *GamePlayer) GetDuration() float64 { return p.duration }

// SetDuration is setDuration.
func (p *GamePlayer) SetDuration(duration float64) { p.duration = duration }

// IsWinner is isWinner.
func (p *GamePlayer) IsWinner() bool { return p.winner }

// SetWinner is setWinner.
func (p *GamePlayer) SetWinner(winner bool) { p.winner = winner }

// GetCompanyId is getCompanyId: Java's short.
func (p *GamePlayer) GetCompanyId() int { return p.companyId }

// SetCompanyId is setCompanyId.
func (p *GamePlayer) SetCompanyId(companyId int) { p.companyId = companyId }

// GetScore is getScore.
func (p *GamePlayer) GetScore() int { return p.score }

// SetScore is setScore.
func (p *GamePlayer) SetScore(score int) { p.score = score }

// ToString is toString: the line the rules log for every player they hand over.
func (p *GamePlayer) ToString() string {
	var id = ""
	var name = ""
	if p.externalUser != nil {
		id = strconv.FormatInt(p.externalUser.GetId(), 10)
		name = p.externalUser.GetName()
	}
	return id + ":" + name +
		" accomplishment:" + util.FormatDouble(p.accomplishment) +
		" duration:" + util.FormatDouble(p.duration) +
		" score:" + strconv.Itoa(p.score) +
		" winner:" + BoolText(p.winner)
}

// BoolText is the "true"/"false" StringBuilder.append(boolean) writes.
func BoolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
