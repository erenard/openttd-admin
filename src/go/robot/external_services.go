// ExternalServices mirrors com.openttd.robot.ExternalServices: the process-wide
// holder for the two services the robot rules call out to - the leaderboard and
// the identity provider of the embedding application. Java makes it a singleton
// with a private constructor; the port keeps one package level instance behind
// GetInstance().
//
// The rules capture the service in a field when they are constructed, exactly as
// Java's field initializers do, so an application (or a test) has to inject the
// services before building the rules.

package robot

import "com.openttd/robot/model"
import "com.openttd/network/admin"

// ExternalGameService is ExternalServices.ExternalGameService: the leaderboard
// side, what a game is, what it produced, and when it ended.
type ExternalGameService interface {
	GetGame(scenarioId int64) *admin.GameInfo
	ExposeGame(scenarioId int64, gameState *admin.GameInfo)
	SaveGame(scenarioId int64, gamePlayers []model.GamePlayer)
	EndGame(scenarioId int64)
}

// ExternalUserService is ExternalServices.ExternalUserService: turning the token
// a player pastes into the chat into the user behind it.
type ExternalUserService interface {
	IdentifyUser(token string) *model.ExternalUser
}

// ExternalServices is ExternalServices: the holder of both services.
type ExternalServices struct {
	externalGameService ExternalGameService
	externalUserService ExternalUserService
}

// instance is ExternalServices.instance: the one and only instance.
var instance = &ExternalServices{}

// GetInstance is ExternalServices.getInstance().
func GetInstance() *ExternalServices {
	return instance
}

// GetExternalGameService is getExternalGameService. It is nil until the
// application injects one, which is why the rules that use it check.
func (s *ExternalServices) GetExternalGameService() ExternalGameService {
	return s.externalGameService
}

// SetExternalGameService is setExternalGameService.
func (s *ExternalServices) SetExternalGameService(service ExternalGameService) {
	s.externalGameService = service
}

// GetExternalUserService is getExternalUserService.
func (s *ExternalServices) GetExternalUserService() ExternalUserService {
	return s.externalUserService
}

// SetExternalUserService is setExternalUserService.
func (s *ExternalServices) SetExternalUserService(service ExternalUserService) {
	s.externalUserService = service
}
