package com.openttd.robot.rule;

import com.openttd.admin.OpenttdAdmin;
import com.openttd.admin.event.DateEvent;
import com.openttd.admin.event.DateEventListener;
import com.openttd.admin.model.Game;
import com.openttd.robot.ExternalServices;
import java.util.ArrayList;
import java.util.Calendar;
import java.util.Collection;

public class ExposeGameInfo extends AbstractRule implements DateEventListener {

    private final long scenarioId;

    public ExposeGameInfo(OpenttdAdmin openttdAdmin, long scenarioId) {
        super(openttdAdmin);
        this.scenarioId = scenarioId;
    }

    @Override
    protected Collection<Class> listEventTypes() {
        Collection<Class> listEventTypes = new ArrayList();
        // Register DateEvent here
        listEventTypes.add(DateEvent.class);
        return listEventTypes;
    }

    @Override
    public void onDateEvent(DateEvent dateEvent) {
        Game game = dateEvent.getOpenttd();
        Calendar now = game.getDate();
        if (isNewWeek(now)) {
            ExternalServices.getInstance().getExternalGameService().exposeGame(scenarioId, game.getGameInfo());
        }
    }

    private int currentWeek = -1;

    private boolean isNewWeek(Calendar now) {
        int nowWeek = now.get(Calendar.WEEK_OF_YEAR);
        if (nowWeek != currentWeek) {
            currentWeek = nowWeek;
            return true;
        }
        return false;
    }

}
