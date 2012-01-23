package com.openttd.admin;

import com.openttd.admin.event.ChatEvent;
import com.openttd.admin.event.ChatEventListener;
import com.openttd.admin.event.ClientEvent;
import com.openttd.admin.event.ClientEventListener;
import com.openttd.admin.event.CompanyEvent;
import com.openttd.admin.event.CompanyEventListener;
import com.openttd.network.core.Configuration;

public class SimpleAdmin extends OpenttdAdmin implements ClientEventListener, ChatEventListener, CompanyEventListener {

	public SimpleAdmin(Configuration configuration) {
		super(configuration);
	}

	@Override
	public void onClientEvent(ClientEvent clientEvent) {
		System.out.println("Client event " + clientEvent);
	}

	@Override
	public void onChatEvent(ChatEvent chatEvent) {
		System.out.print("Chat event : " + chatEvent);
	}

	@Override
	public void onCompanyEvent(CompanyEvent companyEvent) {
		System.out.print("Company event : " + companyEvent);
	}

	/**
	 * Console output demo
	 * 
	 * @param args
	 */
	public static void main(String[] args) {
		Configuration configuration = new Configuration();

		try {
			String url = configuration.password = args[0];
			configuration.host = url.split(":")[0];
			configuration.adminPort = new Integer(url.split(":")[1]);
			configuration.password = args[1];
		} catch (Exception e) {
			System.err.println(e.getMessage());
			System.err.println("Usage: java -jar openttd-admin.jar localhost:3979 admin_password");
			System.err.println("See openttd.cfg to set your server's admin_password first !");
		}

		SimpleAdmin simpleAdmin = new SimpleAdmin(configuration);
		simpleAdmin.addListener(simpleAdmin);
		simpleAdmin.startup();
		System.out.println("Openttd admin started");

		// simpleAdmin.shutdown();
	}
}
