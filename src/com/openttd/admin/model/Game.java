package com.openttd.admin.model;

import java.util.ArrayList;
import java.util.Calendar;
import java.util.Collection;
import java.util.HashMap;
import java.util.Map;
import java.util.Map.Entry;

import com.openttd.admin.event.ChatEvent;
import com.openttd.admin.event.ClientEvent;
import com.openttd.admin.event.CompanyEvent;
import com.openttd.admin.event.DateEvent;
import com.openttd.admin.event.EventDispatcher;
import com.openttd.network.admin.NetworkEvent;
import com.openttd.network.admin.NetworkModel;
import com.openttd.network.admin.NetworkClient.Send;
import com.openttd.network.constant.TcpAdmin.AdminUpdateFrequency;
import com.openttd.network.constant.TcpAdmin.AdminUpdateType;
import com.openttd.util.Convert;

public class Game {

	private final EventDispatcher eventDispatcher;

	public Game(EventDispatcher eventDispatcher) {
		this.eventDispatcher = eventDispatcher;
	}

	public void reset() {
		date = null;
		clientById.clear();
		companyById.clear();
	}

	private Calendar date;

	public Calendar getDate() {
		return date;
	}

	public void setDate(Calendar date) {
		this.date = date;
	}

	private Map<Integer, Client> clientById = new HashMap<Integer, Client>();

	public Client getClient(Integer clientId) {
		if (!clientById.containsKey(clientId)) clientById.put(clientId, new Client());
		return clientById.get(clientId);
	}

	public void removeClient(Integer clientId) {
		clientById.remove(clientId);
	}

	public Collection<Client> getClients() {
		return clientById.values();
	}

	public Collection<Client> getClients(Integer companyId) {
		Collection<Client> clients = new ArrayList<Client>();
		for (Entry<Integer, Client> entry : clientById.entrySet()) {
			Client client = entry.getValue();
			if (client.getCompanyId() == companyId) {
				clients.add(client);
			}
		}
		return clients;
	}

	private Map<Integer, Company> companyById = new HashMap<Integer, Company>();

	public Company getCompany(Integer companyId) {
		if (!companyById.containsKey(companyId)) companyById.put(companyId, new Company());
		return companyById.get(companyId);
	}

	public void removeCompany(Integer companyId) {
		companyById.remove(companyId);
	}

	public Collection<Company> getCompanies() {
		return companyById.values();
	}

	public void update(NetworkEvent event, NetworkModel networkModel, Send send) {
		switch (event.getPacketServerType()) {
		case ADMIN_PACKET_SERVER_CHAT: {
			eventDispatcher.dispatch(new ChatEvent(this, event.getClientId(), event.getMessage()));
			break;
		}
		case ADMIN_PACKET_SERVER_DATE: {
			date = Convert.dayToCalendar(event.getDate());
			eventDispatcher.dispatch(new DateEvent(this));
			break;
		}
		case ADMIN_PACKET_SERVER_RCON: {
			break;
		}
		case ADMIN_PACKET_SERVER_WELCOME: {
			send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_CLIENT_INFO, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC);
			send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_COMPANY_INFO, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC);
			send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_COMPANY_STATS, AdminUpdateFrequency.ADMIN_FREQUENCY_WEEKLY);
			send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_COMPANY_ECONOMY, AdminUpdateFrequency.ADMIN_FREQUENCY_WEEKLY);
			send.updateFrequency(AdminUpdateType.ADMIN_UPDATE_CHAT, AdminUpdateFrequency.ADMIN_FREQUENCY_AUTOMATIC);
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
			com.openttd.network.admin.Client c = networkModel.retreiveClient(clientId);
			Client client = getClient((int) clientId);
			client.setId((int) c.getId());
			client.setIp(c.getIp());
			client.setName(c.getName());
			client.setCompanyId((int) c.getCompanyId());
			eventDispatcher.dispatch(new ClientEvent(this, clientId, ClientEvent.Action.UPDATE));
			break;
		}
		case ADMIN_PACKET_SERVER_CLIENT_ERROR:
		case ADMIN_PACKET_SERVER_CLIENT_QUIT: {
			long clientId = event.getClientId();
			removeClient((int) clientId);
			eventDispatcher.dispatch(new ClientEvent(this, clientId, ClientEvent.Action.DELETE));
			break;
		}
		case ADMIN_PACKET_SERVER_COMPANY_NEW: {
			short companyId = event.getCompanyId();
			Company company = getCompany((int) companyId);
			company.setInauguration(getDate());
			eventDispatcher.dispatch(new CompanyEvent(this, companyId, CompanyEvent.Action.CREATE));
			break;
		}
		case ADMIN_PACKET_SERVER_COMPANY_INFO:
		case ADMIN_PACKET_SERVER_COMPANY_UPDATE:
		case ADMIN_PACKET_SERVER_COMPANY_ECONOMY: {
			short companyId = event.getCompanyId();
			com.openttd.network.admin.Company c = networkModel.retreiveCompany(companyId);
			Company company = getCompany((int) companyId);
			company.setId((int) companyId);
			company.setName(c.getName());
			company.setUsePassword(c.isUsePassword());
			company.setValue(c.getLastValue() != null ? c.getLastValue().longValue() : 0);
			company.setPerformance(c.getLastPerformance());
			company.setPreviousPerformance(c.getPreviousPerformance());
			eventDispatcher.dispatch(new CompanyEvent(this, companyId, CompanyEvent.Action.UPDATE));
			break;
		}
		case ADMIN_PACKET_SERVER_COMPANY_REMOVE: {
			short companyId = event.getCompanyId();
			removeCompany((int) companyId);
			eventDispatcher.dispatch(new CompanyEvent(this, companyId, CompanyEvent.Action.DELETE));
			break;
		}
		}
	}
}
