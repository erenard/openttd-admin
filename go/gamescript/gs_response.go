// GSResponse mirrors com.openttd.gamescript.GSResponse: the three JSON fields
// AdminCmd's gamescript answers with.
//
// MainClass::HandleEvents in src/squirrel/main.nut replies {id, data} for a
// command that worked, {id, exception} for one that failed, and {exception}
// alone when the request itself was malformed, which is why id may be missing.
// Java leaves the fields null until gson fills them in; the Has* flags stand
// in for those nulls, because an empty string is a value gson can also produce.

package gamescript

type GSResponse struct {
	Id           string
	Data         string
	Exception    string
	HasId        bool
	HasData      bool
	HasException bool
}

// NewGSResponse is the default constructor gson uses.
func NewGSResponse() *GSResponse {
	return &GSResponse{}
}

// Parse is gson.fromJson(jsonString, GSResponse.class). AdminCmd writes the
// fields as strings, but a Squirrel value can reach the wire as a number or a
// bool, so every field is read as its JSON text rather than requiring a string.
// Returns nil for anything that is not a JSON object, which is how a reply gson
// could not map at all shows up.
func Parse(json string) *GSResponse {
	var fields = RawFields(json)
	if fields == nil {
		return nil
	}
	var response = NewGSResponse()
	var id, hasId = lookup(fields, "id")
	if hasId {
		response.Id = id
		response.HasId = true
	}
	var data, hasData = lookup(fields, "data")
	if hasData {
		response.Data = data
		response.HasData = true
	}
	var exception, hasException = lookup(fields, "exception")
	if hasException {
		response.Exception = exception
		response.HasException = true
	}
	return response
}

// lookup finds one field. A JSON null is what gson maps to a null String field,
// so it reads as absent here.
func lookup(fields map[string]string, key string) (string, bool) {
	var value, has = fields[key]
	if !has || value == "null" {
		return "", false
	}
	return value, true
}
