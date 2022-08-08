package com.openttd.admin.event;

public interface CompanyEventListener extends EventListener {

    void onCompanyEvent(CompanyEvent companyEvent);
}
