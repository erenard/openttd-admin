package com.openttd.robot.rule;

import java.util.ArrayList;
import java.util.Calendar;
import java.util.Collection;

import com.openttd.admin.OpenttdAdmin;
import com.openttd.admin.event.DateEvent;
import com.openttd.admin.event.DateEventListener;
import com.openttd.network.admin.Company;
import com.openttd.admin.model.Game;
import com.openttd.network.admin.NetworkAdminSender;

/**
 * Warn non-passworded companies, once a month.
 */
public class CompanyPasswordRemainder extends AbstractRule implements DateEventListener {

    public CompanyPasswordRemainder(OpenttdAdmin openttdAdmin) {
        super(openttdAdmin);
    }

    private int currentMonth;

    private boolean isNewMonth(Calendar now) {
        int nowMonth = now.get(Calendar.MONTH) + 12 * now.get(Calendar.YEAR);
        if (nowMonth > currentMonth) {
            currentMonth = nowMonth;
            return true;
        }
        return false;
    }

    @Override
    public void onDateEvent(DateEvent dateEvent) {
        Game openttd = dateEvent.getOpenttd();
        Calendar now = openttd.getDate();
        if (isNewMonth(now)) {
            //NetworkAdminSender message to non-passworded companies
            for (Company company : openttd.getCompanies()) {
                if (!company.isUsePassword()) {
                    warningNoPassword(company.getId());
                }
            }
        }
    }

    @SuppressWarnings("rawtypes")
    @Override
    public Collection<Class> listEventTypes() {
        Collection<Class> listEventTypes = new ArrayList();
        listEventTypes.add(DateEvent.class);
        return listEventTypes;
    }

    private void warningNoPassword(int companyId) {
        NetworkAdminSender send = super.getSend();
        send.chatCompany((short) companyId, "Warning: your company has NO PASSWORD.");
    }
}
