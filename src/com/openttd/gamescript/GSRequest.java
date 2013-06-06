package com.openttd.gamescript;

import com.google.gson.JsonObject;

public abstract class GSRequest<T> {
	private String id;
	private GSResponse response;

	String getId() {
		return this.id;
	}

	void setId(String id) {
		this.id = id;
	}

	public boolean hasResponse() {
		return response != null;
	}

	public T getResult() throws GSException {
		if(response.exception != null) {
			throw new GSException(response.exception);
		}
		return fromJson(response.data);
	}

	void setResponse(GSResponse response) {
		this.response = response;
	}

	protected abstract JsonObject toJson();
	protected abstract T fromJson(String json);

	@Override
	public String toString() {
		return this.toJson().toString();
	}
}
