package com.openttd.robot.rule;

import java.util.ArrayList;
import java.util.Collection;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Map.Entry;
import java.util.Set;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.openttd.admin.OpenttdAdmin;
import com.openttd.admin.event.ChatEvent;
import com.openttd.admin.event.ChatEventListener;
import com.openttd.network.admin.Company;
import com.openttd.admin.model.Game;
import com.openttd.network.admin.NetworkAdminSender;
import com.openttd.robot.ExternalServices;
import com.openttd.robot.ExternalServices.ExternalUserService;
import com.openttd.robot.model.ExternalUser;

/**
 * Store the users info Rule #1: Handle !login Rule #2: Handle !howto Rule #3:
 * Handle !rename
 */
public class ExternalUsers extends AbstractRule implements ChatEventListener {

    private static final Logger log = LoggerFactory.getLogger(ExternalUsers.class);

    private final ExternalUserService externalUserService = ExternalServices.getInstance().getExternalUserService();

    public ExternalUsers(OpenttdAdmin openttdAdmin) {
        super(openttdAdmin);
    }

    private final Map<Long, ExternalUser> externalUserByClientId = new HashMap<>();
    private final Map<Short, ExternalUser> externalUserByCompanyId = new HashMap<>();

    @Override
    public void onChatEvent(ChatEvent chatEvent) {
        long clientId = chatEvent.getClientId();
        String message = chatEvent.getMessage();
        if (message != null) {
            message = message.trim();
            String command = message.toLowerCase();
            if (command.startsWith("!login ")) {
                //Rule #1
                String loginToken = message.substring("!login ".length());
                ExternalUser user = externalUserService.identifyUser(loginToken);
                if (user != null) {
                    boolean logged = true;
                    if (externalUserByClientId.containsKey(clientId)) {
                        if (externalUserByClientId.get(clientId).equals(user)) {
                            showClientAlreadyLogged(clientId);
                        } else {
                            logged = false;
                        }
                    } else {
                        showLoginSucceed(clientId);
                        externalUserByClientId.put(clientId, user);
                        renameClient(clientId, user.getName());
                    }
                    Game openttd = chatEvent.getOpenttd();
                    short companyId = openttd.getClient(clientId).getCompanyId();
                    if (logged && companyId != 255) {
                        ExternalUser companyOwner = getOwnerOf(companyId);
                        Company company = openttd.getCompany(companyId);
                        String companyName = company.getName();
                        if (companyOwner != null) {
                            if (companyOwner.equals(user)) {
                                showCompanyAlreadyOwned(clientId, companyName);
                            } else {
                                showCompanyAlreadyOwned(clientId, companyName, companyOwner);
                            }
                        } else {
                            Set<Entry<Short, ExternalUser>> copy = new HashSet<>(externalUserByCompanyId.entrySet());
                            for (Entry<Short, ExternalUser> entry : copy) {
                                log.info(entry.getKey() + " " + entry.getValue());
                                if (user.equals(entry.getValue())) {
                                    removeOwnerOf(entry.getKey());
                                }
                            }
                            setOwnerOf(companyId, user);
                            showCompanyOwned(clientId, companyName);
                        }
                    }
                } else {
                    showLoginFailed(clientId);
                    showHowtoLogin(clientId);
                }
            } else if (command.equals("!howto")) {
                //Rule #2
                showHowtoLogin(clientId);
            } else if (command.startsWith("!rename ")) {
                //Rule #3
                String newName = message.substring("!rename ".length()).trim();
                renameClient(clientId, newName);
            }
        }
    }

    @Override
    public Collection<Class> listEventTypes() {
        Collection<Class> listEventTypes = new ArrayList();
        listEventTypes.add(ChatEvent.class);
        return listEventTypes;
    }

    private void showCompanyOwned(long clientId, String companyName) {
        NetworkAdminSender send = super.getSend();
        send.chatClient(clientId, "Congratulation ! You now own " + companyName + ".");
    }

    private void showCompanyAlreadyOwned(long clientId, String companyName) {
        NetworkAdminSender send = super.getSend();
        send.chatClient(clientId, "You already own " + companyName + ".");
    }

    private void showCompanyAlreadyOwned(long clientId, String companyName, ExternalUser user) {
        NetworkAdminSender send = super.getSend();
        send.chatClient(clientId, companyName + "is already owned by " + user.getName());
    }

    private void renameClient(long clientId, String name) {
        NetworkAdminSender send = super.getSend();
        if (name.length() > 1) {
            if (name.startsWith("Player")) {
                send.chatClient(clientId, "Player... names are not allowed, try again...");
            } else {
                while (name.indexOf(' ') != -1) {
                    name = name.replace(' ', '_');
                }
                send.rcon("client_name " + clientId + " " + name);
            }
        } else {
            send.chatClient(clientId, "Try : !rename NewName");
        }
    }

    public ExternalUser getExternalUser(long clientId) {
        return externalUserByClientId.get(clientId);
    }

    public ExternalUser getOwnerOf(short companyId) {
        return externalUserByCompanyId.get(companyId);
    }

    public void setOwnerOf(short companyId, ExternalUser owner) {
        externalUserByCompanyId.put(companyId, owner);
    }

    public void removeOwnerOf(short companyId) {
        externalUserByCompanyId.remove(companyId);
    }

    private void showLoginFailed(long clientId) {
        NetworkAdminSender send = super.getSend();
        send.chatClient(clientId, "Login failed.");
    }

    private void showLoginSucceed(long clientId) {
        NetworkAdminSender send = super.getSend();
        send.chatClient(clientId, "Login succeed.");
    }

    public void showHowtoLogin(long clientId) {
        NetworkAdminSender send = super.getSend();
        send.chatClient(clientId, "How to login ***");
        send.chatClient(clientId, "1. Goto www.strategyboard.net and register there,");
        send.chatClient(clientId, "2. Login there and click 'In game login',");
        send.chatClient(clientId, "3. Copy and paste the whole !login command line into the chat window of the game.");
    }

    private void showClientAlreadyLogged(long clientId) {
        NetworkAdminSender send = super.getSend();
        send.chatClient(clientId, "You are already logged.");
    }

}
