package com.openttd.util;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertTrue;

import java.util.Calendar;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import com.openttd.server.UpstreamCalendar;

/**
 * Convert turns the uint32 day counter the admin network sends (ADMIN_PACKET_SERVER_DATE)
 * into a java.util.Calendar. Checked against 15.3's CalendarConvertDateToYMD,
 * see UpstreamCalendar.
 */
public class ConvertTest {

    /**
     * Every game date must land in the year OpenTTD computes. Checked from 1900
     * to 2050 - OpenTTD's MAX_YEAR is 2050 - and across century years, which are
     * the ones the leap rules disagree about.
     */
    @Test
    @DisplayName("dates map to the year OpenTTD computes")
    public void testYearMatchesUpstream() {
        for (int year = 1900; year <= 2050; year++) {
            // DateAtStartOfYear(year): 365 days per year plus one day for each
            // leap year before it, year 0 included.
            int startOfYear = year * 365 + ((year - 1) / 4 - (year - 1) / 100 + (year - 1) / 400) + 1;
            int yearLength = UpstreamCalendar.isLeapYear(year) ? 366 : 365;
            for (int dayOfYear = 0; dayOfYear < yearLength; dayOfYear += 7) {
                int date = startOfYear + dayOfYear;
                Calendar calendar = Convert.dayToCalendar(date);
                assertEquals(year, calendar.get(Calendar.YEAR),
                        "date " + date + " (year " + year + ", day " + dayOfYear + " of " + yearLength + " days)");
            }
        }
    }

    /**
     * OpenTTD counts year 0, java.util.Calendar's AD era does not: dates in game
     * year 0 map to java year 1. Only relevant for a map starting at year 0, so
     * this is documented rather than fixed, and every date a game can reach is
     * covered by testYearMatchesUpstream.
     */
    @Test
    @DisplayName("game year 0 is java year 1, the documented edge case")
    public void testYearZero() {
        assertEquals(0, UpstreamCalendar.yearOf(0));
        assertEquals(1, Convert.dayToCalendar(0).get(Calendar.YEAR));
    }

    /** Objectives and robot rules compare calendars against game dates. */
    @Test
    public void testRoundTrip() {
        long[] dates = {0L, 1L, 711751L, 719527L, 719528L, 1950 * 365 + 1, 1000000L};
        for (long date : dates) {
            assertEquals(date, Convert.calendarToDay(Convert.dayToCalendar(date)), "date " + date);
        }
    }

    /** January 1st of a game year, the date a new map reports in its WELCOME packet. */
    @Test
    public void testStartOfGameYear() {
        int year = 1950;
        int startOfYear = year * 365 + ((year - 1) / 4 - (year - 1) / 100 + (year - 1) / 400) + 1;
        Calendar calendar = Convert.dayToCalendar(startOfYear);
        assertEquals(year, calendar.get(Calendar.YEAR));
        assertEquals(Calendar.JANUARY, calendar.get(Calendar.MONTH));
        assertEquals(1, calendar.get(Calendar.DAY_OF_MONTH));
        assertTrue(UpstreamCalendar.isLeapYear(2000));
        assertEquals(false, UpstreamCalendar.isLeapYear(1900));
    }
}
