package com.openttd.robot.rule;

import java.util.ArrayList;
import java.util.Collection;
import java.util.HashMap;
import java.util.Map;

import com.openttd.admin.OpenttdAdmin;
import com.openttd.admin.event.ChatEvent;
import com.openttd.admin.event.ChatEventListener;
import com.openttd.network.admin.Client;
import com.openttd.admin.model.Game;
import com.openttd.network.admin.NetworkAdminSender;

/**
 * Take care of the admins commands Rule #1: Handle $clients, $cls Rule #2:
 * Handle $kick, $k Rule #3: Handle $ban, $b //TODO Test Rule #4: Handle $warn,
 * $w Rule #5: $help, $h Rule #6: $pause, $p Rule #7: $unpause, $u
 */
public class Administration extends AbstractRule implements ChatEventListener {

    private final ExternalUsers externalUsers;

    public Administration(OpenttdAdmin openttdAdmin, ExternalUsers externalUsers) {
        super(openttdAdmin);
        this.externalUsers = externalUsers;
    }

    @Override
    public void onChatEvent(ChatEvent chatEvent) {
        int clientId = chatEvent.getClientId();
        String message = chatEvent.getMessage();
        if (message != null) {
            message = message.trim().toLowerCase();
            if (externalUsers.getExternalUser(clientId) != null
                    && externalUsers.getExternalUser(clientId).isAdmin()) {
                if (message.equals("$clients") || message.equals("$cls")) {
                    //Rule #1
                    showClients(chatEvent.getOpenttd(), clientId);
                } else if (message.startsWith("$kick") || message.startsWith("$k")) {
                    //Rule #2
                    try {
                        String argument = message.split(" ")[1].trim();
                        Integer kickedClientId = Integer.parseInt(argument);
                        kick(kickedClientId);
                    } catch (NumberFormatException e) {
                        showMessage(clientId, "Usage: $kick clientId (try $clients to find clientIds)");
                    }
                } else if (message.startsWith("$ban") || message.startsWith("$b")) {
                    //Rule #3
                    try {
                        String argument = message.split(" ")[1].trim();
                        Integer bannedClientId = Integer.parseInt(argument);
                        ban(bannedClientId);
                    } catch (NumberFormatException ignore) {
                        showMessage(clientId, "Usage: $ban clientId (try $clients to find clientIds)");
                    }
                } else if (message.startsWith("$warn") || message.startsWith("$w")) {
                    //Rule #4
                    try {
                        message = chatEvent.getMessage().trim();
                        String[] args = message.split(" ");
                        int warnedClientId = Integer.parseInt(args[1]);
                        String warningMessage = "";
                        if (args.length > 1) {
                            warningMessage = message.substring(message.indexOf(args[2]));
                        }
                        warn(warnedClientId, warningMessage);
                    } catch (NumberFormatException ignore) {
                        showMessage(clientId, "Usage: $warn clientId [message] (try $clients to find clientIds)");
                    }
                } else if (message.startsWith("$help") || message.equals("$h")) {
                    //Rule #5
                    showHelpAdmin(clientId);
                } else if (message.startsWith("$pause") || message.equals("$p")) {
                    //Rule #6
                    pause();
                } else if (message.startsWith("$unpause") || message.equals("$u")) {
                    //Rule #7
                    unpause();
                }
            }
        }
    }

    @SuppressWarnings("rawtypes")
    @Override
    public Collection<Class> listEventTypes() {
        Collection<Class> listEventTypes = new ArrayList();
        listEventTypes.add(ChatEvent.class);
        return listEventTypes;
    }

    private void showClients(Game openttd, int clientId) {
        NetworkAdminSender send = super.getSend();
        send.chatClient(clientId, "#ClientId, Name, IpAddress, CompanyId");
        for (Client client : openttd.getClients()) {
            send.chatClient(clientId, "#" + client.getId() + ", " + client.getName() + ", " + client.getIp() + ", " + client.getCompanyId());
        }
    }

    private void showMessage(int clientId, String message) {
        NetworkAdminSender send = super.getSend();
        send.chatClient(clientId, message);
    }

    private void kick(int clientId) {
        NetworkAdminSender send = super.getSend();
        send.rcon("kick " + clientId);
    }

    private void ban(int clientId) {
        NetworkAdminSender send = super.getSend();
        send.rcon("ban " + clientId);
    }

    private Map<Integer, Integer> warningCountByClientId = new HashMap<>();

    private void warn(int clientId, String warning) {
        NetworkAdminSender send = super.getSend();
        Integer warningCount = warningCountByClientId.get(clientId);
        if (warningCount == null) {
            warningCount = 0;
        }
        if (warningCount == 0) {
            send.chatClient(clientId, "Be warned! No rule breaking!" + (warning != null ? " (" + warning + ")" : ""));
        } else if (warningCount == 1) {
            send.chatClient(clientId, "Last warning! No rule breaking!" + (warning != null ? " (" + warning + ")" : ""));
        } else if (warningCount == 2) {
            kick(clientId);
        } else if (warningCount > 2) {
            ban(clientId);
        }
        warningCountByClientId.put(clientId, warningCount + 1);
    }

    private void showHelpAdmin(int clientId) {
        NetworkAdminSender send = super.getSend();
        send.chatClient(clientId, "Valid commands are $clients, $companies, $warn, $kick, $ban, $reset, $pause, $unpause");
        send.chatClient(clientId, "Short commands are $cls, $cps, $w, $k, $b, $r, $p, $u");
    }

    private void pause() {
        NetworkAdminSender send = super.getSend();
        send.rcon("pause");
    }

    private void unpause() {
        NetworkAdminSender send = super.getSend();
        send.rcon("unpause");
    }

}
