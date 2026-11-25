// GSJsonRequest mirrors com.openttd.gamescript.GSJsonRequest: a gamescript
// command that is not modelled by its own GSRequest, e.g. the goal commands
// NetworkAdminSender sends. The reply's raw data is the result.
//
// Java's arguments are a gson JsonArray; here they are the already rendered
// JSON items, in order, which is what the wire payload needs and what keeps
// the payload byte identical to the Java port's.

package gamescript

import "com.openttd/network/constant"

type GSJsonRequest struct {
	command constant.GSCommand
	arguments []string
	id string
	response *GSResponse
}

// NewGSJsonRequest is the constructor.
func NewGSJsonRequest(command constant.GSCommand, arguments []string) *GSJsonRequest {
	return &GSJsonRequest{command: command, arguments: arguments}
}

// GetCommand is getCommand.
func (r *GSJsonRequest) GetCommand() constant.GSCommand { return r.command }

// GetArguments is getArguments.
func (r *GSJsonRequest) GetArguments() []string { return r.arguments }

// GetId, SetId, HasResponse and SetResponse are the GSRequest base class.
func (r *GSJsonRequest) GetId() string { return r.id }

func (r *GSJsonRequest) SetId(id string) { r.id = id }

func (r *GSJsonRequest) HasResponse() bool { return r.response != nil }

func (r *GSJsonRequest) SetResponse(response *GSResponse) { r.response = response }

// ToJson is toJson: "i", "c" then "a", the order gson would have written them.
func (r *GSJsonRequest) ToJson() string {
	var fields = []JsonField{
		JsonField{Key: constant.GS_ID, Value: JsonString(r.id)},
		JsonField{Key: constant.GS_CMD, Value: JsonNumber(int(r.command))},
		JsonField{Key: constant.GS_ARGS, Value: JsonList(r.arguments)},
	}
	return JsonObject(fields)
}

// ToString is toString. The wire payload must be pure JSON: AdminCmd parses
// it with JSONDecode, so a command name prefix would make it reply
// "Malformed json.".
func (r *GSJsonRequest) ToString() string { return r.ToJson() }

// GetResult is fromJson, which for this request is the reply's raw data. Java
// throws GSException for an "exception" reply and would have thrown a
// NullPointerException when no reply had arrived; both come back as a GSException
// here, since Go has no exceptions.
func (r *GSJsonRequest) GetResult() (string, *GSException) {
	if r.response == nil {
		return "", NewException("No gamescript reply for request " + r.id)
	}
	if r.response.HasException {
		return "", NewException(r.response.Exception)
	}
	return r.response.Data, nil
}
