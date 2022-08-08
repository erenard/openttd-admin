package com.openttd.demo;

import java.util.ArrayList;
import java.util.Calendar;
import java.util.Collection;

import com.openttd.admin.OpenttdAdmin;
import com.openttd.admin.event.DateEvent;
import com.openttd.admin.event.DateEventListener;
import com.openttd.network.core.Configuration;
import com.openttd.robot.rule.AbstractRule;
import com.openttd.robot.rule.Administration;
import com.openttd.robot.rule.ExternalUsers;

public class AdministrationDemo extends OpenttdAdmin {

    /**
     * Ingame help rule. Show the help message every monday. See HelloWorldTest
     * for more info on rules
     */
    private static class IngameInfo extends AbstractRule implements DateEventListener {

        public IngameInfo(OpenttdAdmin openttdAdmin) {
            super(openttdAdmin);
        }

        @SuppressWarnings("rawtypes")
        @Override
        protected Collection<Class> listEventTypes() {
            Collection<Class> classes = new ArrayList<>();
            classes.add(DateEvent.class);
            return classes;
        }

        @Override
        public void onDateEvent(DateEvent dateEvent) {
            Calendar date = dateEvent.getOpenttd().getDate();
            if (date.get(Calendar.DAY_OF_WEEK) == Calendar.MONDAY) {
                this.getSend().chatBroadcast("Welcome to this login demo.");
                this.getSend().chatBroadcast("Type '!login username' to log on (any name will work).");
                this.getSend().chatBroadcast("If your username starts with 'admin', you will have moderation powers.");
                this.getSend().chatBroadcast("Once logged, type '$help' to know more about moderation commands.");
            }
        }
    }

    public AdministrationDemo(Configuration configuration) {
        super(configuration);
        //External user login rule
        ExternalUsers externalUsers = new ExternalUsers(this);
        //Administration commands rule
        new Administration(this, externalUsers);
        //Ingame info rule
        new IngameInfo(this);
    }

    public static void main(String[] args) {
        Configuration configuration = new Configuration();
        CLIUtil.parseArguments(args, configuration);
        TestUtil.fakeExternalUserService();
        //Create the robot
        HelloWorldDemo robot = new HelloWorldDemo(configuration);
        //Start the robot and connect it to OpenTTD
        robot.startup();
        //Wait 3 minutes
        try {
            Thread.sleep(180000);
        } catch (InterruptedException e) {
            e.printStackTrace();
        }
        //Disconnect the robot and shut it down
        robot.shutdown();
    }

}
