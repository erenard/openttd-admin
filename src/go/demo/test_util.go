// TestUtil is com.openttd.demo.TestUtil: the two fake services the demos put
// into the ExternalServices singleton so the robot rules have something to call
// without an embedding application behind them. Anyone can log in - the token
// becomes the user's name - and a token that starts with "admin" gets the
// moderation powers the Administration rule gates its commands on.
//
// Java prints with System.out.println(); the port logs at info, which is the
// only stream the port has that a test can point at a file. There is no Java
// test for this class, so demo_test.go is new coverage written against this
// source.

package demo

import "strings"
import "com.openttd/robot"
import "com.openttd/robot/model"
import "com.openttd/network/admin"
import "com.openttd/util"

var demoLog = util.GetLogger("TestUtil")

// FakeUserService is the lambda Java passes to setExternalUserService: the token
// is the name, and the "admin" prefix is the whole of its idea of a moderator.
type FakeUserService struct{}

func NewFakeUserService() *FakeUserService {
	return &FakeUserService{}
}

// IdentifyUser is the body of that lambda.
func (s *FakeUserService) IdentifyUser(token string) *model.ExternalUser {
	var externalUser = model.NewExternalUser()
	externalUser.SetName(token)
	if strings.HasPrefix(token, "admin") {
		externalUser.SetAdmin(true)
	}
	return externalUser
}

// FakeGameService is the anonymous ExternalGameService Java builds with new:
// it prints what the rules hand it and forgets everything else.
type FakeGameService struct {
	// ended is the port's stand-in for the System.exit(0) of endGame: the
	// callback records that the game is over instead of killing the process,
	// which is what the demo's Main does when it sees this flag.
	ended bool
}

func NewFakeGameService() *FakeGameService {
	return &FakeGameService{ended: false}
}

// GetGame is getGame(long scenarioId): the fake has no leaderboard to read, so
// it returns null, which is why the rules that call it check for it.
func (s *FakeGameService) GetGame(scenarioId int64) *admin.GameInfo {
	return nil
}

// ExposeGame is exposeGame(long, GameInfo): System.out.println(gameInfo). A
// null prints as "null" in Java, which is what the port's empty-string stand-in
// would otherwise hide.
func (s *FakeGameService) ExposeGame(scenarioId int64, gameState *admin.GameInfo) {
	if gameState == nil {
		demoLog.Info("null")
		return
	}
	demoLog.Info(gameState.ToString())
}

// SaveGame is saveGame(long, Collection<GamePlayer>): the scenario id, then one
// line per player.
func (s *FakeGameService) SaveGame(scenarioId int64, gamePlayers []model.GamePlayer) {
	demoLog.Info(util.FormatInt(scenarioId))
	for _, gamePlayer := range gamePlayers {
		demoLog.Info(gamePlayer.ToString())
	}
}

// EndGame is endGame(long): Java leaves the process. The port records it, so
// the caller can stop the way the demo's Main does.
func (s *FakeGameService) EndGame(scenarioId int64) {
	s.ended = true
}

// IsEnded tells the caller what EndGame recorded.
func (s *FakeGameService) IsEnded() bool { return s.ended }

// FakeExternalUserService is fakeExternalUserService(): put both fakes into the
// singleton. The rules capture the services when they are constructed, so this
// has to run before the rules are built, exactly as the demos call it.
func FakeExternalUserService() {
	var services = robot.GetInstance()
	services.SetExternalUserService(NewFakeUserService())
	services.SetExternalGameService(NewFakeGameService())
}
