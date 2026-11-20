package com.openttd.gamescript;

import com.google.gson.JsonArray;
import com.google.gson.JsonObject;
import com.google.gson.JsonPrimitive;
import com.openttd.network.constant.GameScript;
import com.openttd.network.constant.GameScript.GSCommand;

/**
 * Gamescript command that is not modelled by its own GSRequest, e.g. the goal
 * commands sent by NetworkAdminSender. The reply's raw data is the result.
 */
public class GSJsonRequest extends GSRequest<String> {

    private final GSCommand command;
    private final JsonArray arguments;

    public GSJsonRequest(GSCommand command, JsonArray arguments) {
        this.command = command;
        this.arguments = arguments;
    }

    public GSCommand getCommand() {
        return command;
    }

    public JsonArray getArguments() {
        return arguments;
    }

    @Override
    protected JsonObject toJson() {
        JsonObject json = new JsonObject();
        json.add(GameScript.ID, new JsonPrimitive(this.getId()));
        json.add(GameScript.CMD, new JsonPrimitive(command.ordinal()));
        json.add(GameScript.ARGS, arguments);
        return json;
    }

    @Override
    protected String fromJson(String json) {
        return json;
    }

    @Override
    public String toString() {
        // The wire payload must be pure JSON: AdminCmd parses it with
        // JSONDecode, so a command name prefix would make it reply "Malformed json.".
        return super.toString();
    }
}
