package com.openttd.admin.model;

import java.util.ArrayList;
import java.util.Calendar;
import java.util.Collection;

import com.openttd.admin.event.ChatEvent;
import com.openttd.admin.event.ClientEvent;
import com.openttd.admin.event.CompanyEvent;
import com.openttd.admin.event.ConsoleEvent;
import com.openttd.admin.event.DateEvent;
import com.openttd.admin.event.EventDispatcher;
import com.openttd.admin.event.GameScriptEvent;
import com.openttd.admin.event.PongEvent;
import com.openttd.admin.event.RConEndEvent;
import com.openttd.admin.event.RConEvent;
import com.openttd.network.admin.Client;
import com.openttd.network.admin.Company;
import com.openttd.network.admin.GameInfo;
import com.openttd.network.admin.NetworkAdminEvent;
import com.openttd.network.admin.NetworkAdminSender;
import com.openttd.network.admin.NetworkModel;
import com.openttd.network.constant.TcpAdmin.AdminUpdateFrequency;
import com.openttd.network.constant.TcpAdmin.AdminUpdateType;
import com.openttd.util.Convert;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

/**
 * Game is the state of the openttd's game. This class fill two purpose: 1.
 * updating its own state, 2. translating technical events (NetworkEvent) into
 * functionnal events (CompanyEvent, ClientEvent...)
 */
public class Game {

    private static final Logger log = LoggerFactory.getLogger(Game.class);

    private final EventDispatcher eventDispatcher;

    public Game(EventDispatcher eventDispatcher) {
        this.eventDispatcher = eventDispatcher;
    }

    private NetworkModel networkModel;

    public GameInfo getGameInfo() {
        return networkModel.getGameInfo();
    }

    public Calendar getDate() {
        return Convert.dayToCalendar(networkModel.getGameInfo().getCurrentDate());
    }

    public Client getClient(long clientId) {
        return networkModel.retreiveClient(clientId);
    }

    public Collection<Client> getClients() {
        return networkModel.retreiveClients();
    }

    public Collection<Client> getClients(short companyId) {
        Collection<Client> clients = new ArrayList();
        for (Client client : getClients()) {
            if (client.getCompanyId() == companyId) {
                clients.add(client);
            }
        }
        return clients;
    }

    public Company getCompany(short companyId) {
        return networkModel.retreiveCompany(companyId);
    }

    public Collection<Company> getCompanies() {
        return networkModel.retreiveCompanies();
    }

    public void updateAdmin(NetworkAdminEvent event, NetworkModel networkModel, NetworkAdminSender send) {
        log.debug("Game.updateAdmin " + event);
        this.networkModel = networkModel;
        switch (event.getPacketServerType()) {
            case ADMIN_PACKET_SERVER_CHAT: {
                eventDispatcher.dispatch(new ChatEvent(this, event.getClientId(), event.getMessage()));
                break;
            }
            case ADMIN_PACKET_SERVER_DATE: {
                eventDispatcher.dispatch(new DateEvent(this));
                break;
            }
            case ADMIN_PACKET_SERVER_NEWGAME: {
                break;
            }
            case ADMIN_PACKET_SERVER_RCON: {
                eventDispatcher.dispatch(new RConEvent(this, event.getColor(), event.getMessage()));
                break;
            }
            case ADMIN_PACKET_SERVER_WELCOME: {
                send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_CLIENT_INFO, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC);
                send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_COMPANY_INFO, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC);
                send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_COMPANY_STATS, AdminUpdateFrequency.ADMIN_FREQUENCY_WEEKLY);
                send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_COMPANY_ECONOMY, AdminUpdateFrequency.ADMIN_FREQUENCY_WEEKLY);
                send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_CHAT, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC);
                send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_GAMESCRIPT, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC);
                send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_CONSOLE, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC);
                send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_DATE, AdminUpdateFrequency.ADMIN_FREQUENCY_DAILY);
                send.pollDate();
                send.pollClientInfo((long) -1);
                send.pollCompanyInfo((short) -1);
                send.pollCompanyEconomy((short) -1);
                break;
            }
            case ADMIN_PACKET_SERVER_CLIENT_JOIN: {
                long clientId = event.getClientId();
                eventDispatcher.dispatch(new ClientEvent(this, clientId, ClientEvent.Action.CREATE));
                break;
            }
            case ADMIN_PACKET_SERVER_CLIENT_INFO:
            case ADMIN_PACKET_SERVER_CLIENT_UPDATE: {
                long clientId = event.getClientId();
                eventDispatcher.dispatch(new ClientEvent(this, clientId, ClientEvent.Action.UPDATE));
                break;
            }
            case ADMIN_PACKET_SERVER_CLIENT_ERROR:
            case ADMIN_PACKET_SERVER_CLIENT_QUIT: {
                long clientId = event.getClientId();
                eventDispatcher.dispatch(new ClientEvent(this, clientId, ClientEvent.Action.DELETE));
                break;
            }
            case ADMIN_PACKET_SERVER_COMPANY_NEW: {
                short companyId = event.getCompanyId();
                eventDispatcher.dispatch(new CompanyEvent(this, companyId, CompanyEvent.Action.CREATE));
                break;
            }
            case ADMIN_PACKET_SERVER_COMPANY_INFO:
            case ADMIN_PACKET_SERVER_COMPANY_UPDATE:
            case ADMIN_PACKET_SERVER_COMPANY_ECONOMY: {
                short companyId = event.getCompanyId();
                eventDispatcher.dispatch(new CompanyEvent(this, companyId, CompanyEvent.Action.UPDATE));
                break;
            }
            case ADMIN_PACKET_SERVER_COMPANY_REMOVE: {
                short companyId = event.getCompanyId();
                eventDispatcher.dispatch(new CompanyEvent(this, companyId, CompanyEvent.Action.DELETE));
                break;
            }
            case ADMIN_PACKET_SERVER_GAMESCRIPT: {
                eventDispatcher.dispatch(new GameScriptEvent(this, event.getMessage()));
            }
                break;
            case ADMIN_PACKET_SERVER_RCON_END: {
                eventDispatcher.dispatch(new RConEndEvent(this, event.getMessage()));
            }
            case ADMIN_PACKET_SERVER_CONSOLE: {
                eventDispatcher.dispatch(new ConsoleEvent(this, event.getOrigin(), event.getMessage()));
                break;
            }
            case ADMIN_PACKET_SERVER_PONG: {
                eventDispatcher.dispatch(new PongEvent(this, event.getPingId()));
                break;
            }
        }
    }
}
