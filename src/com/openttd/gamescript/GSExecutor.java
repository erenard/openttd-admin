package com.openttd.gamescript;

import com.google.gson.Gson;
import com.openttd.admin.OpenttdAdmin;
import com.openttd.admin.event.GameScriptEvent;
import com.openttd.admin.event.GameScriptEventListener;
import com.openttd.network.admin.NetworkAdminSender;

import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ConcurrentMap;

public class GSExecutor implements GameScriptEventListener {

	private final OpenttdAdmin admin;
	private static int nextRequestId = 0;
	private final Gson gson = new Gson();
	private final ConcurrentMap<String, GSRequest> requestById = new ConcurrentHashMap<String, GSRequest>();

	public GSExecutor(OpenttdAdmin admin) {
		this.admin = admin;
	}

	public void clear() {
		GSExecutor.nextRequestId = 0;
		requestById.clear();
	}

	public <T> void send(GSRequest<T> request) {
		NetworkAdminSender send = this.admin.getSend();
		if(request != null && send != null) {
			String id = "" + GSExecutor.nextRequestId++;
			request.setId(id);
			this.requestById.put(id, request);
			send.gameScript(request.toString());
		}
	}

	@Override
	public void onGameScriptEvent(GameScriptEvent gameScriptEvent) {
		String jsonString = gameScriptEvent.getMessage();
		GSResponse response = gson.fromJson(jsonString, GSResponse.class);
		this.requestById.remove(response.id).setResponse(response);
	}
}
