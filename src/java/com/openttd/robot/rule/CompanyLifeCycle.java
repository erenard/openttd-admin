package com.openttd.robot.rule;

import java.util.ArrayList;
import java.util.Collection;
import java.util.Formatter;
import java.util.HashMap;
import java.util.Iterator;
import java.util.Map;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.openttd.admin.OpenttdAdmin;
import com.openttd.admin.event.ChatEvent;
import com.openttd.admin.event.ChatEventListener;
import com.openttd.admin.event.CompanyEvent;
import com.openttd.admin.event.CompanyEvent.Action;
import com.openttd.admin.event.CompanyEventListener;
import com.openttd.network.admin.Client;
import com.openttd.network.admin.Company;
import com.openttd.admin.model.Game;
import com.openttd.network.admin.NetworkAdminSender;
import com.openttd.robot.model.ExternalUser;

/**
 * Company life-cycle rule : Rule #1: only logged player can create Rule #2:
 * only one company by ip address can exist Rule #3: handle !resetme Rule #4:
 * handle $companies, $cps Rule #5: handle !info Rule #6: handle $reset
 */
public class CompanyLifeCycle extends AbstractRule implements CompanyEventListener, ChatEventListener {

    private static final Logger log = LoggerFactory.getLogger(CompanyLifeCycle.class);

    private final ExternalUsers externalUsers;

    public CompanyLifeCycle(OpenttdAdmin openttdAdmin, ExternalUsers externalUsers) {
        super(openttdAdmin);
        this.externalUsers = externalUsers;
    }

    //Company id stored by the ip address of the creator
    private final Map<String, Short> companyIdByIpAddress = new HashMap<>();

    //Configuration
    public boolean checkLogin = true;
    public boolean checkIpAddress = true;

    @SuppressWarnings("rawtypes")
    @Override
    public Collection<Class> listEventTypes() {
        Collection<Class> listEventTypes = new ArrayList();
        listEventTypes.add(CompanyEvent.class);
        listEventTypes.add(ChatEvent.class);
        return listEventTypes;
    }

    @Override
    public void onCompanyEvent(CompanyEvent companyEvent) {
        if (log.isDebugEnabled()) {
            log.debug("CompanyLifeCycle.onCompanyEvent " + companyEvent.toString());
        }
        short companyId = companyEvent.getCompanyId();
        Action action = companyEvent.getAction();
        Game game = companyEvent.getOpenttd();
        switch (action) {
            case CREATE: {
                if (applyRules(companyId, game)) {
                    // Find the company owner
                    ExternalUser companyOwner = null;
                    {
                        Collection<Client> clients = game.getClients(companyId);
                        if (clients != null) {
                            Client creator = clients.iterator().next();
                            if (creator != null) {
                                //Client is logged
                                companyOwner = externalUsers.getExternalUser(creator.getId());
                            }
                        }
                    }

                    if (companyOwner != null) {
                        //First remove ownership of every other company
                        Collection<Company> companies = game.getCompanies();
                        for (Company company : companies) {
                            ExternalUser externalUser = externalUsers.getOwnerOf(company.getId());
                            if (companyOwner.equals(externalUser)) {
                                externalUsers.removeOwnerOf(company.getId());
                            }
                        }
                        //Then own a company
                        externalUsers.setOwnerOf(companyId, companyOwner);
                    }
                } else {
                    deleteCompany(game, companyId);
                }
                break;
            }
            case UPDATE: {
                break;
            }
            case DELETE: {
                //Keep the local model up to date
                externalUsers.removeOwnerOf(companyId);
                for (Iterator<Map.Entry<String, Short>> iterator = companyIdByIpAddress.entrySet().iterator(); iterator.hasNext();) {
                    Map.Entry<String, Short> entry = iterator.next();
                    if (entry.getValue().equals(companyId)) {
                        iterator.remove();
                    }
                }
                break;
            }
        }
    }

    private boolean applyRules(short companyId, Game game) {
        NetworkAdminSender send = super.getSend();
        // Find the company owner
        ExternalUser companyOwner = null;
        Client creator = null;
        String ipAddress = null;
        {
            Collection<Client> clients = game.getClients(companyId);
            if (clients != null) {
                creator = clients.iterator().next();
                if (creator != null) {
                    ipAddress = creator.getIp();
                    //Client is logged
                    companyOwner = externalUsers.getExternalUser(creator.getId());
                }
            }
        }
        // Verify Rule #1
        if (checkLogin && companyOwner == null) {
            if (creator != null) {
                send.chatClient(creator.getId(), "Company " + companyId + " creation rejected: Please loggin before start a new company.");
                externalUsers.showHowtoLogin(creator.getId());
            }
            return false;
        }
        // Verify Rule #2
        if (checkIpAddress && ipAddress != null) {
            if (companyOwner == null || !companyOwner.isAdmin()) {
                Short previousCompanyId = companyIdByIpAddress.get(ipAddress);
                if (previousCompanyId != null) {
                    Company previousCompany = game.getCompany(previousCompanyId);
                    if (previousCompany != null) {
                        String companyName = game.getCompany(companyId).getName();
                        String previousCompanyName = game.getCompany(previousCompanyId).getName();
                        StringBuffer sb = new StringBuffer();
                        try ( Formatter formatter = new Formatter(sb)) {
                            formatter.format("Company '%s' creation rejected: You (%s) already own '%s'.", companyName, ipAddress, previousCompanyName);
                            if (creator != null) {
                                send.chatClient(creator.getId(), sb.toString());
                            }
                        }
                        return false;
                    }
                }
            }
            // Add the company
            companyIdByIpAddress.put(ipAddress, companyId);
        }
        return true;
    }

    private void deleteCompany(Game game, short companyId) {
        NetworkAdminSender send = super.getSend();
        Collection<Client> clients = game.getClients(companyId);
        for (Client client : clients) {
            send.rcon("move " + client.getId() + " 255");
        }
        send.rcon("reset_company " + (companyId + 1));
    }

    @Override
    public void onChatEvent(ChatEvent chatEvent) {
        if (log.isDebugEnabled()) {
            log.debug("CompanyLifeCycle.onChatEvent " + chatEvent.toString());
        }
        int clientId = chatEvent.getClientId();
        String message = chatEvent.getMessage();
        Game openttd = chatEvent.getOpenttd();
        if (message != null) {
            message = message.trim().toLowerCase();
            if (message.equals("!resetme")) {
                //Rule #3
                Client client = openttd.getClient(clientId);
                short companyId = client.getCompanyId();
                if (companyId >= 0 && companyId < 255) {
                    deleteCompany(openttd, companyId);
                }
            } else if (message.equals("!info")) {
                //Rule #5
                showInfo(openttd, clientId);
            } else if (externalUsers.getExternalUser(clientId) != null
                    && externalUsers.getExternalUser(clientId).isAdmin()) {
                if (message.equals("$companies") || message.equals("$cps")) {
                    //Rule #4
                    showCompanies(openttd, clientId);
                } else if (message.startsWith("$reset") || message.startsWith("$r")) {
                    //Rule #6
                    try {
                        String argument = message.split(" ")[1].trim();
                        short companyId = Short.parseShort(argument);
                        deleteCompany(openttd, companyId);
                    } catch (NumberFormatException ignore) {
                        showMessage(clientId, "Usage: $reset companyId (try $companies to find companyIds)");
                    }
                }
            }
        }
    }

    private void showMessage(int clientId, String message) {
        NetworkAdminSender send = super.getSend();
        send.chatClient(clientId, message);
    }

    private void showCompanies(Game openttd, Integer clientId) {
        NetworkAdminSender send = super.getSend();
        send.chatClient(clientId, "#Id, Name, CEO, Inauguration.");
        for (Company company : openttd.getCompanies()) {
            if (company != null) {
                StringBuilder sb = new StringBuilder("#");
                try ( Formatter formatter = new Formatter(sb)) {
                    formatter.format("# %3d, %255s, %255s, %4d",
                            company.getId(),
                            company.getName(),
                            externalUsers.getOwnerOf(company.getId()).getName(),
                            company.getInauguratedYear());
                    send.chatClient(clientId, sb.toString());
                }
            }
        }
    }

    private void showInfo(Game openttd, Integer clientId) {
        NetworkAdminSender send = super.getSend();
        Client client = openttd.getClient(clientId);
        if (client != null) {
            String userName = "Anonymous";
            {
                ExternalUser user = externalUsers.getExternalUser(clientId);
                if (user != null) {
                    userName = user.getName();
                }
            }
            send.chatClient(clientId, "User: " + userName + ", play as (" + client.getName() + ").");
            short companyId = client.getCompanyId();
            if (companyId == 255) {
                send.chatClient(clientId, "Spectator.");
            } else {
                Company company = openttd.getCompany(companyId);
                ExternalUser owner = externalUsers.getOwnerOf(companyId);
                if (company != null && owner != null) {
                    send.chatClient(clientId, "Company: " + company.getName() + ", owned by " + owner.getName() + ".");
                }
            }
        }
    }
}
