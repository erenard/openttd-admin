package com.openttd.server;

/**
 * Reference implementation of OpenTTD 15.3's calendar arithmetic, copied from
 * `git show 15.3:src/timer/timer_game_common.cpp` (CalendarConvertDateToYMD) and
 * `timer_game_common.h` (CalendarTime::DAYS_IN_YEAR / DAYS_IN_LEAP_YEAR).
 *
 * Used to pin Convert.dayToCalendar(): an admin gets game dates as a uint32 day
 * counter, so a date that lands in the wrong year makes DateEvent consumers
 * (strategyboard's objectives, openttd-robot's timers) act on the wrong day.
 *
 * Note OpenTTD counts year 0, so a date of `year * 365 + 1` is not that year:
 * leap days have to be subtracted, which is exactly what this function does.
 */
public class UpstreamCalendar {

    private static final int DAYS_IN_YEAR = 365;
    private static final int DAYS_IN_LEAP_YEAR = 366;

    /** git show 15.3:src/timer/timer_game_common.h, CalendarTime::IsLeapYear */
    public static boolean isLeapYear(int year) {
        return (year % 4 == 0 && (year % 100 != 0 || year % 400 == 0)) || (year % 4000 == 0);
    }

    /**
     * Year part of CalendarConvertDateToYMD(date): large steps first (400 years
     * hold 97 leap years), then centuries, then the 4 year cycles.
     *
     * @param date game date, days since year 0
     * @return year
     */
    public static int yearOf(int date) {
        int year = 400 * (date / (DAYS_IN_YEAR * 400 + 97));
        int remaining = date % (DAYS_IN_YEAR * 400 + 97);

        if (remaining >= DAYS_IN_YEAR * 100 + 25) {
            year += 100;
            remaining -= DAYS_IN_YEAR * 100 + 25;
            year += 100 * (remaining / (DAYS_IN_YEAR * 100 + 24));
            remaining = remaining % (DAYS_IN_YEAR * 100 + 24);
        }

        if (!isLeapYear(year) && remaining >= DAYS_IN_YEAR * 4) {
            year += 4;
            remaining -= DAYS_IN_YEAR * 4;
        }

        year += 4 * (remaining / (DAYS_IN_YEAR * 4 + 1));
        remaining = remaining % (DAYS_IN_YEAR * 4 + 1);

        while (remaining >= (isLeapYear(year) ? DAYS_IN_LEAP_YEAR : DAYS_IN_YEAR)) {
            remaining -= isLeapYear(year) ? DAYS_IN_LEAP_YEAR : DAYS_IN_YEAR;
            year++;
        }
        return year;
    }
}
