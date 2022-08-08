package com.openttd.util;

import java.util.Calendar;

public class Convert {

    private static final long openttdToJavaOffset = -62167392000000l;

    private static final long openttdToJavaMultiplier = 24l * 60l * 60l * 1000l;

    /**
     * Return a Calendar reflecting the ingame day
     *
     * @param day
     * @return a Calendar reflecting the ingame day
     */
    public static Calendar dayToCalendar(long day) {
        Calendar calendar = Calendar.getInstance();
        calendar.setTimeInMillis(openttdToJavaOffset + (day + 2) * openttdToJavaMultiplier);
        return calendar;
    }

    public static long calendarToDay(Calendar calendar) {
        long millis = calendar.getTimeInMillis();
        millis -= openttdToJavaOffset;
        return (millis / openttdToJavaMultiplier) - 2;
    }
}
