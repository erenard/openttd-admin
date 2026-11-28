// GSRequest mirrors com.openttd.gamescript.GSRequest, the abstract base class
// GSExecutor keeps in its map of pending requests.
//
// Go has no inheritance, so the base class becomes the interface GSExecutor
// handles, and each concrete request carries its own id and response. Java's
// getId, setId and setResponse are package private; they are exported here
// because GSExecutor reaches them through the interface. getResult() stays
// per request, because its return type is Java's type parameter T.

package gamescript

type GSRequest interface {
	GetId() string
	SetId(id string)
	HasResponse() bool
	SetResponse(response *GSResponse)
	ToString() string
}
