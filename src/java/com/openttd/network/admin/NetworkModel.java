package com.openttd.network.admin;

import java.util.Collection;
import java.util.HashMap;
import java.util.Map;
import java.util.Map.Entry;
import java.util.concurrent.locks.Lock;
import java.util.concurrent.locks.ReentrantLock;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class NetworkModel implements Cloneable {

    private static final Logger log = LoggerFactory.getLogger(NetworkModel.class);

    private GameInfo gameInfo = new GameInfo();

    public GameInfo getGameInfo() {
        gameInfo.clientsCount = clientById.size();
        gameInfo.companiesCount = companyById.size();
        return gameInfo;
    }

    public void setGameInfo(GameInfo gameInfo) {
        this.gameInfo = gameInfo;
    }

    private final Map<Short, Company> companyById = new HashMap();

    public Company retreiveCompany(short companyId) {
        if (!companyById.containsKey(companyId)) {
            companyById.put(companyId, new Company(companyId));
        }
        return companyById.get(companyId);
    }

    public Collection<Company> retreiveCompanies() {
        return companyById.values();
    }

    public void deleteCompany(short companyId) {
        companyById.remove(companyId);
    }

    private final Map<Long, Client> clientById = new HashMap();

    public Client retreiveClient(long clientId) {
        if (!clientById.containsKey(clientId)) {
            clientById.put(clientId, new Client(clientId));
        }
        return clientById.get(clientId);
    }

    public Collection<Client> retreiveClients() {
        return clientById.values();
    }

    public void deleteClient(long clientId) {
        clientById.remove(clientId);
    }

    /**
     * DoCommand names announced by ADMIN_PACKET_SERVER_CMD_NAMES. Not stable
     * across OpenTTD versions, used for logging purposes only.
     */
    private final Map<Integer, String> commandNameById = new HashMap();

    public void putCommandName(int commandId, String name) {
        commandNameById.put(commandId, name);
    }

    public Map<Integer, String> retreiveCommandNames() {
        return commandNameById;
    }

    private Lock lock = new ReentrantLock();

    public void lock() {
        lock.lock();
    }

    public void unlock() {
        lock.unlock();
    }

    public NetworkModel snapshot() {
        try {
            lock.lock();
            return (NetworkModel) this.clone();
        } catch (CloneNotSupportedException ignore) {
            return null;
        } finally {
            lock.unlock();
        }
    }

    @Override
    protected Object clone() throws CloneNotSupportedException {
        NetworkModel clone = new NetworkModel();
        clone.setGameInfo((GameInfo) gameInfo.clone());
        for (Entry<Short, Company> entry : companyById.entrySet()) {
            clone.companyById.put(entry.getKey(), (Company) entry.getValue().clone());
        }
        for (Entry<Long, Client> entry : clientById.entrySet()) {
            clone.clientById.put(entry.getKey(), (Client) entry.getValue().clone());
        }
        clone.commandNameById.putAll(commandNameById);
        return clone;
    }

    public NetworkModel copy() {
        try {
            return (NetworkModel) this.clone();
        } catch (CloneNotSupportedException e) {
            log.error(e.getMessage(), e);
        }
        return null;
    }
}
