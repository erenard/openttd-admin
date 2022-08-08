package com.openttd.admin.event;

import com.openttd.admin.model.Game;

public class CompanyEvent implements Event<CompanyEventListener> {

    public enum Action {
        CREATE,
        UPDATE,
        DELETE
    };

    private final Game openttd;
    private final short companyId;
    private final Action action;

    public CompanyEvent(Game openttd, short companyId, Action action) {
        this.openttd = openttd;
        this.companyId = companyId;
        this.action = action;
    }

    public short getCompanyId() {
        return companyId;
    }

    public Game getOpenttd() {
        return openttd;
    }

    public Action getAction() {
        return action;
    }

    @Override
    public String toString() {
        return action + ":" + companyId;
    }

    @Override
    public void notify(CompanyEventListener listener) {
        listener.onCompanyEvent(this);
    }
}
