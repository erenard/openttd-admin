package com.openttd.robot;

import com.openttd.network.admin.GameInfo;
import java.util.Collection;

import com.openttd.robot.model.ExternalUser;
import com.openttd.robot.model.GamePlayer;

public class ExternalServices {

    private static final ExternalServices instance = new ExternalServices();

    private ExternalServices() {
    }

    public static ExternalServices getInstance() {
        return instance;
    }

    public interface ExternalGameService {
        GameInfo getGame(long scenarioId);
        void exposeGame(long scenarioId, GameInfo gameState);
        void saveGame(long scenarionId, Collection<GamePlayer> gamePlayers);
        void endGame(long scenarioId);
    }

    public interface ExternalUserService {

        ExternalUser identifyUser(String token);
    }

    private ExternalGameService externalGameService;
    private ExternalUserService externalUserService;

    public ExternalGameService getExternalGameService() {
        return externalGameService;
    }

    public void setExternalGameService(ExternalGameService externalGameService) {
        this.externalGameService = externalGameService;
    }

    public ExternalUserService getExternalUserService() {
        return externalUserService;
    }

    public void setExternalUserService(ExternalUserService externalUserService) {
        this.externalUserService = externalUserService;
    }
}
