/*
 * To change this license header, choose License Headers in Project Properties.
 * To change this template file, choose Tools | Templates
 * and open the template in the editor.
 */
package com.openttd.demo;

import com.openttd.network.admin.GameInfo;
import com.openttd.robot.ExternalServices;
import com.openttd.robot.model.ExternalUser;
import com.openttd.robot.model.GamePlayer;
import java.util.Collection;

/**
 *
 * @author erenard
 */
public class TestUtil {

    public static void fakeExternalUserService() {
        //Fake user identification
        ExternalServices.getInstance().setExternalUserService((String token) -> {
            ExternalUser externalUser = new ExternalUser();
            externalUser.setName(token);
            if (token.startsWith("admin")) {
                externalUser.setAdmin(true);
            }
            return externalUser;
        });
        //Fake game service
        ExternalServices.getInstance().setExternalGameService(new ExternalServices.ExternalGameService() {
            @Override
            public void exposeGame(long scenarioId, GameInfo gameInfo) {
                System.out.println(gameInfo);
            }

            @Override
            public void saveGame(long scenarioId, Collection<GamePlayer> gamePlayers) {
                System.out.println(scenarioId);
                for (GamePlayer gamePlayer : gamePlayers) {
                    System.out.println(gamePlayer.toString());
                }
            }

            @Override
            public void endGame(long scenarioId) {
                System.exit(0);
            }

            @Override
            public GameInfo getGame(long scenarioId) {
                return null;
            }
        });
    }

}
