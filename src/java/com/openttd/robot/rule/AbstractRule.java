package com.openttd.robot.rule;

import java.util.Collection;

import com.openttd.admin.OpenttdAdmin;
import com.openttd.admin.event.EventListener;
import com.openttd.network.admin.NetworkAdminSender;

public abstract class AbstractRule implements EventListener {

    protected final OpenttdAdmin openttdAdmin;


    public AbstractRule(OpenttdAdmin openttdAdmin) {
        this.openttdAdmin = openttdAdmin;
        this.register();
    }

    protected NetworkAdminSender getSend() {
        return openttdAdmin.getSend();
    }

    abstract protected Collection<Class> listEventTypes();

    public final void register() {
        for (Class eventClass : listEventTypes()) {
            this.openttdAdmin.addListener(eventClass, this);
        }
    }

    public final void unregister() {
        for (Class eventClass : listEventTypes()) {
            this.openttdAdmin.removeListener(eventClass, this);
        }
    }
}
