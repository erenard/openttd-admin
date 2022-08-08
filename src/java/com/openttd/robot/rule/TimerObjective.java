package com.openttd.robot.rule;

import java.text.DateFormat;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.Collection;
import java.util.Collections;
import java.util.Formatter;
import java.util.List;
import java.util.Locale;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.openttd.admin.OpenttdAdmin;
import com.openttd.admin.event.ChatEvent;
import com.openttd.admin.event.ChatEventListener;
import com.openttd.admin.event.DateEvent;
import com.openttd.admin.event.DateEventListener;
import com.openttd.network.admin.Company;
import com.openttd.admin.model.Game;
import com.openttd.gamescript.GSNewsPaper;
import com.openttd.network.admin.NetworkAdminSender;
import com.openttd.robot.ExternalServices;
import com.openttd.robot.ExternalServices.ExternalGameService;
import com.openttd.robot.model.ExternalUser;
import com.openttd.robot.model.GamePlayer;

/**
 *
 * Rule #1: Handle !goal, !g Rule #2: Handle !score, !cv, !cp Rule #3: Handle
 * $end
 */
public class TimerObjective extends AbstractRule implements DateEventListener, ChatEventListener {

    private static final Logger log = LoggerFactory.getLogger(TimerObjective.class);

    private final ExternalGameService externalGameService = ExternalServices.getInstance().getExternalGameService();
    private final ExternalUsers externalUsers;
    private final long scenarioId;
    private final int nbYear;
    /* Game State */
    private boolean gameEnded;
    private Calendar endGame;
    private Calendar restartServer;
    private Collection<ScoreBean> finalScores;
    /* Utility Variables */
    private int currentWeek;
    private int currentTrimester;

    public TimerObjective(OpenttdAdmin openttdAdmin, ExternalUsers externalUsers, long scenarioId, int nbYear) {
        super(openttdAdmin);
        this.externalUsers = externalUsers;
        this.scenarioId = scenarioId;
        this.nbYear = nbYear;
        this.gameEnded = false;
        this.currentWeek = 0;
        this.currentTrimester = 0;
    }

    private boolean isNewTrimester(Calendar now) {
        int nowTrimester = now.get(Calendar.MONTH) / 3 + 4 * now.get(Calendar.YEAR);
        if (nowTrimester > currentTrimester) {
            currentTrimester = nowTrimester;
            return true;
        }
        return false;
    }

    private boolean isNewWeek(Calendar now) {
        int nowWeek = now.get(Calendar.WEEK_OF_YEAR);
        if (nowWeek != currentWeek) {
            currentWeek = nowWeek;
            return true;
        }
        return false;
    }

    @Override
    public Collection<Class> listEventTypes() {
        Collection<Class> listEventTypes = new ArrayList();
        listEventTypes.add(ChatEvent.class);
        listEventTypes.add(DateEvent.class);
        return listEventTypes;
    }

    @Override
    public void onDateEvent(DateEvent dateEvent) {
        Calendar now = dateEvent.getOpenttd().getDate();
        if (endGame == null) {
            endGame = (Calendar) now.clone();
            endGame.add(Calendar.YEAR, nbYear);
        }
        if (isNewWeek(now)) {
            Game openttd = dateEvent.getOpenttd();
            if (!gameEnded) {
                // End game condition
                if (now.after(endGame)) {
                    finalScores = calculateScores(openttd);
                    saveGame();
                    restartServer = (Calendar) now.clone();
                    restartServer.add(Calendar.MONTH, 1);
                    gameEnded = true;
                }
            }
            if (gameEnded) {
                if (now.after(restartServer)) {
                    endGame();
                } else {
                    broadcastVictory(dateEvent.getOpenttd());
                }
            }
        }
        //Show game progression
        if (!gameEnded && isNewTrimester(now)) {
            Game openttd = dateEvent.getOpenttd();
            showScore(-1, openttd);
        }
    }

    private Collection<GamePlayer> saveGame() {
        Collection<GamePlayer> gamePlayers = new ArrayList<>();
        boolean winnerSet = false;
        for (ScoreBean scoreBean : finalScores) {
            log.info(scoreBean.companyId + "(" + scoreBean.score + "): " + scoreBean.externalUser);
            if (scoreBean.externalUser != null) {
                GamePlayer gamePlayer = new GamePlayer();
                gamePlayer.setExternalUser(scoreBean.externalUser);
                if (!winnerSet) {
                    gamePlayer.setWinner(true);
                    winnerSet = true;
                }
                gamePlayer.setAccomplishment(scoreBean.score);
                gamePlayer.setDuration(1);
                gamePlayer.setCompanyId(scoreBean.companyId);
                gamePlayers.add(gamePlayer);
            }
        }
        externalGameService.saveGame(scenarioId, gamePlayers);
        return gamePlayers;
    }

    @Override
    public void onChatEvent(ChatEvent chatEvent) {
        int clientId = chatEvent.getClientId();
        String message = chatEvent.getMessage();
        if (message != null) {
            message = message.trim().toLowerCase();
            switch (message) {
                case "!goal": //Rule #1
                case "!g": //Rule #1
                    showGoal(clientId);
                    break;
                case "!score": //Rule #2
                case "!cv": //Rule #2
                case "!cp": //Rule #2
                    showScore(clientId, chatEvent.getOpenttd());
                    break;
                case "$end": {
                    //Rule #3
                    ExternalUser user = externalUsers.getExternalUser(clientId);
                    if (user != null && user.isAdmin()) {
                        NetworkAdminSender send = super.getSend();
                        send.chatClient(clientId, "Game ending now !");
                        endGame = chatEvent.getOpenttd().getDate();
                    }
                    break;
                }
                default: {
                    break;
                }
            }
        }
    }

    private void showGoal(int clientId) {
        NetworkAdminSender send = super.getSend();
        send.chatClient(clientId, "Goal ***");
        if (endGame != null) {
            DateFormat dateFormat = DateFormat.getDateInstance(DateFormat.LONG, Locale.UK);
            send.chatClient(clientId, "Reach the best company performance the " + dateFormat.format(endGame.getTime()) + " to win the game.");
        } else {
            send.chatClient(clientId, "Reach the best company performance in " + nbYear + "years to win the game.");
        }
        send.chatClient(clientId, "Your score will be your company's performance.");
    }

    private void broadcastVictory(Game openttd) {
        NetworkAdminSender send = super.getSend();
        send.chatBroadcast("Game Over ***");
        send.rcon("setting min_active_clients 0");
        if (finalScores != null) {
            boolean winner = true;
            for (ScoreBean scoreBean : finalScores) {
                StringBuilder stringBuilder = new StringBuilder();
                try (Formatter formatter = new Formatter(stringBuilder)) {
                    if (scoreBean.externalUser != null) {
                        if (winner) {
                            formatter.format("%s wins the game and get %d points on the leaderboard.",
                                    scoreBean.externalUser.getName(), scoreBean.score);
                            GSNewsPaper newsPaper = new GSNewsPaper(GSNewsPaper.NewsType.NT_GENERAL, stringBuilder.toString());
                            openttdAdmin.getGSExecutor().send(newsPaper);
                            winner = false;
                        } else {
                            formatter.format("%s get %d points on the leaderboard.",
                                    scoreBean.externalUser.getName(), scoreBean.score);
                        }
                    } else {
                        Company company = openttd.getCompany(scoreBean.companyId);
                        formatter.format("%s: %d points wasted, player not logged.",
                                company.getName(), scoreBean.score);
                    }
                    send.chatBroadcast(stringBuilder.toString());
                }
            }
        }
        if (restartServer != null) {
            DateFormat dateFormat = DateFormat.getDateInstance(DateFormat.LONG, Locale.UK);
            send.chatBroadcast("Server will restart the " + dateFormat.format(restartServer.getTime()) + " ***");
        } else {
            send.chatBroadcast("Server will restart ***");
        }
    }

    private void showScore(long clientId, Game openttd) {
        NetworkAdminSender send = super.getSend();
        send.chatClient(clientId, "Score ***");
        Collection<ScoreBean> scoreBeans = calculateScores(openttd);
        for (ScoreBean scoreBean : scoreBeans) {
            StringBuilder stringBuilder = new StringBuilder();
            try ( Formatter formatter = new Formatter(stringBuilder)) {
                formatter.format("%d (%+d) - %s - %s.",
                        scoreBean.score,
                        scoreBean.score - scoreBean.lastScore,
                        scoreBean.companyName,
                        scoreBean.externalUser != null ? scoreBean.externalUser.getName() : "Nobody");
                send.chatClient(clientId, stringBuilder.toString());
            }
        }
    }

    private void endGame() {
        externalGameService.endGame(scenarioId);
    }

    private class ScoreBean implements Comparable<ScoreBean> {

        public int score;
        public int lastScore;
        public short companyId;
        public String companyName;
        public ExternalUser externalUser;

        @Override
        public int compareTo(ScoreBean o) {
            if (score != o.score) {
                return score - o.score;
            } else {
                if (lastScore != o.lastScore) {
                    return lastScore - o.lastScore;
                } else {
                    if (externalUser != null && o.externalUser == null) {
                        return 1;
                    }
                    if (externalUser == null && o.externalUser != null) {
                        return -1;
                    }
                    return 0;
                }
            }
        }
    }

    private Collection<ScoreBean> calculateScores(Game openttd) {
        List<ScoreBean> scoreBeans = new ArrayList<>();
        for (Company company : openttd.getCompanies()) {
            ScoreBean scoreBean = new ScoreBean();
            scoreBean.score = company.getLastPerformance();
            scoreBean.lastScore = company.getPreviousPerformance();
            scoreBean.companyName = company.getName();
            var companyId = company.getId();
            scoreBean.companyId = companyId;
            scoreBean.externalUser = externalUsers.getOwnerOf(companyId);
            scoreBeans.add(scoreBean);
        }
        Collections.sort(scoreBeans);
        Collections.reverse(scoreBeans);
        return scoreBeans;
    }
}
