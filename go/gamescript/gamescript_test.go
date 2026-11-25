// The gamescript request/response plumbing, offline.
//
// See test/java/com/openttd/gamescript/GSExecutorTest.java. The Java test drives
// this through AdminHarness and a FakeAdminServer, because it checks that a
// reply survives on the listener thread; here the executor is called directly,
// so the same cases need no waiting, and the "still connected" check becomes
// "the sender still takes the next packet".
//
// The GSExecutorTest cases that need a live admin and a real gamescript reply
// round trip are in network/admin, once NetworkAdminSender can send gamescript.

package gamescript_test

import "strconv"
import "testing"
import "com.openttd/gamescript"
import "com.openttd/network/constant"

// FakeSender is the FakeAdminServer's half: it records the payload the executor
// put on the wire.
type FakeSender struct {
	packets []string
}

func (s *FakeSender) GameScript(json string) bool {
	s.packets = append(s.packets, json)
	return true
}

// FakeAdmin is the OpenttdAdmin half: getSend(), null until connected.
type FakeAdmin struct {
	sender *FakeSender
}

func (a *FakeAdmin) GetSender() gamescript.Sender {
	if a.sender == nil {
		return nil
	}
	return a.sender
}

// Fixture is the harness: a connected admin, its sender and the executor.
type Fixture struct {
	sender *FakeSender
	admin *FakeAdmin
	executor *gamescript.GSExecutor
}

func newFixture() *Fixture {
	var sender = FakeSender{}
	var admin = FakeAdmin{}
	admin.sender = &sender
	var out = Fixture{}
	out.sender = &sender
	out.admin = &admin
	out.executor = gamescript.NewGSExecutor(&admin)
	out.executor.Clear()
	return &out
}

// sentId is awaitRequestOnWire(): the id of the request the executor put on the
// wire, at the given position.
func sentId(fixture *Fixture, index int) string {
	if index >= len(fixture.sender.packets) {
		return ""
	}
	var fields = gamescript.RawFields(fixture.sender.packets[index])
	if fields == nil {
		return ""
	}
	var id, has = fields[constant.GS_ID]
	if !has {
		return ""
	}
	return id
}

// The request shape AdminCmd requires: "i" id, "c" command, "a" arguments, all
// three present, or it answers "Malformed json.".
func TestRequestJsonShape(t *testing.T) {
	var arguments = []string{
		gamescript.JsonNumber(1),
		gamescript.JsonString("deliver cargo"),
		gamescript.JsonNumber(int(constant.GOAL_TOWN)),
		gamescript.JsonNumber(4),
	}
	var commands = constant.GSCommandValues()
	for i := 0; i < len(commands); i++ {
		var command = commands[i]
		var request = gamescript.NewGSJsonRequest(command, arguments)
		request.SetId("7")
		var payload = request.ToString()
		var fields = gamescript.RawFields(payload)
		if fields == nil {
			t.Errorf("%v is parseable JSON", constant.GSCommandName(command))
			continue
		}
		var id, hasId = fields[constant.GS_ID]
		if !hasId || id != "7" {
			t.Errorf("%v needs the request id", constant.GSCommandName(command))
		}
		var cmd, hasCmd = fields[constant.GS_CMD]
		if !hasCmd || cmd != strconv.Itoa(int(command)) {
			t.Errorf("%v needs the command id", constant.GSCommandName(command))
		}
		var args, hasArgs = fields[constant.GS_ARGS]
		if !hasArgs || args != "[1,\"deliver cargo\",3,4]" {
			t.Errorf("%v needs the arguments in order, got %v", constant.GSCommandName(command), args)
		}
		// toString() is what goes on the wire, so it must be pure JSON: the
		// Java version prefixed the command name and AdminCmd rejected it.
		if payload != request.ToJson() {
			t.Errorf("%v wire payload is pure JSON", constant.GSCommandName(command))
		}
	}
	// gson's JsonObject is insertion ordered, and its JsonArray keeps the
	// arguments in order.
	var request = gamescript.NewGSJsonRequest(constant.GS_COMMAND_ADD_GOAL, arguments)
	request.SetId("7")
	if request.ToString() != "{\"i\":\"7\",\"c\":2,\"a\":[1,\"deliver cargo\",3,4]}" {
		t.Errorf("the payload is %v", request.ToString())
	}
}

// GSNewsPaper is the one request modelled as its own class: createNews.
func TestNewsPaperRequestJson(t *testing.T) {
	var request = gamescript.NewGSNewsPaperForCompany(2, gamescript.NT_GENERAL, "hello")
	request.SetId("3")
	if request.ToString() != "{\"i\":\"3\",\"c\":1,\"a\":[14,\"hello\",2]}" {
		t.Errorf("the payload is %v", request.ToString())
	}
	// The single argument list of the two argument constructor is "for every
	// company", which Java writes as -1.
	var everyone = gamescript.NewGSNewsPaper(gamescript.NT_GENERAL, "hello")
	everyone.SetId("3")
	if everyone.ToString() != "{\"i\":\"3\",\"c\":1,\"a\":[14,\"hello\",-1]}" {
		t.Errorf("the payload is %v", everyone.ToString())
	}
}

// A news type the gamescript cannot use is replaced by NT_GENERAL, and the
// replacement is still visible on the request afterwards.
func TestNewsPaperInvalidNewsType(t *testing.T) {
	if gamescript.NT_ACCIDENT.IsValid() {
		t.Errorf("NT_ACCIDENT is not a gamescript news type")
	}
	if !gamescript.NT_GENERAL.IsValid() || !gamescript.NT_ECONOMY.IsValid() || !gamescript.NT_SUBSIDIES.IsValid() {
		t.Errorf("general, economy and subsidies are the usable news types")
	}
	var request = gamescript.NewGSNewsPaperForCompany(1, gamescript.NT_ACCIDENT, "boom")
	request.SetId("4")
	if request.ToString() != "{\"i\":\"4\",\"c\":1,\"a\":[14,\"boom\",1]}" {
		t.Errorf("the payload is %v", request.ToString())
	}
}

// A reply with data completes the request, and the id the executor chose is the
// one on the wire.
func TestReplyWithId(t *testing.T) {
	var fixture = newFixture()
	var request = gamescript.NewGSJsonRequest(constant.GS_COMMAND_COUNT_INDUSTRIE, []string{})
	fixture.executor.Send(request)
	if len(fixture.sender.packets) != 1 {
		t.Errorf("the request reached the wire")
		return
	}
	var id = sentId(fixture, 0)
	if id != request.GetId() {
		t.Errorf("the id on the wire is the one the executor correlated")
	}
	fixture.executor.OnGameScriptReply("{\"id\":\"" + id + "\",\"data\":\"7\"}")
	if !request.HasResponse() {
		t.Errorf("the gamescript reply completed the request")
	}
	var result, error = request.GetResult()
	if error != nil || result != "7" {
		t.Errorf("the result is %v", result)
	}
}

// AdminCmd answers a failing command with {id, exception}, and getResult()
// reports it instead of a value.
func TestReplyWithException(t *testing.T) {
	var fixture = newFixture()
	var request = gamescript.NewGSJsonRequest(constant.GS_COMMAND_SET_TOWN_CARGO_GOAL, []string{})
	fixture.executor.Send(request)
	var id = sentId(fixture, 0)
	fixture.executor.OnGameScriptReply("{\"id\":\"" + id + "\",\"exception\":\"Invalid town\"}")
	if !request.HasResponse() {
		t.Errorf("the gamescript error reply")
		return
	}
	var result, error = request.GetResult()
	if error == nil {
		t.Errorf("getResult reports the exception")
	}
	if error != nil && error.GetExceptionMessage() != "Invalid town" {
		t.Errorf("the exception carries AdminCmd's message")
	}
	if result != "" {
		t.Errorf("no result comes with an exception")
	}
}

// Malformed requests are answered with {exception} alone: without an id there
// is nothing to correlate, and the executor must survive it. In Java this
// threw a NullPointerException on the listener thread, which took the admin
// connection down; the Java test proves the connection survived with a PONG.
func TestReplyWithoutId(t *testing.T) {
	var fixture = newFixture()
	fixture.executor.OnGameScriptReply("{\"exception\":\"Malformed json.\"}")
	assertStillSending(fixture, t)
}

// A reply whose id was never sent must not complete another request.
func TestReplyWithUnknownId(t *testing.T) {
	var fixture = newFixture()
	var request = gamescript.NewGSJsonRequest(constant.GS_COMMAND_COUNT_INDUSTRIE, []string{})
	fixture.executor.Send(request)
	var otherId = strconv.Itoa(atoi(sentId(fixture, 0)) + 1000)
	fixture.executor.OnGameScriptReply("{\"id\":\"" + otherId + "\",\"data\":\"7\"}")
	if request.HasResponse() {
		t.Errorf("a reply for another request must not complete this one")
	}
	assertStillSending(fixture, t)
}

// Requests sent after clear() start from id 0 again.
func TestClearRestartsIds(t *testing.T) {
	var fixture = newFixture()
	var request = gamescript.NewGSJsonRequest(constant.GS_COMMAND_COUNT_INDUSTRIE, []string{})
	fixture.executor.Send(request)
	if request.GetId() != "0" {
		t.Errorf("the first id is %v", request.GetId())
	}
	fixture.executor.Clear()
	var next = gamescript.NewGSJsonRequest(constant.GS_COMMAND_COUNT_INDUSTRIE, []string{})
	fixture.executor.Send(next)
	if next.GetId() != "0" {
		t.Errorf("ids restart after clear(), got %v", next.GetId())
	}
}

// GSExecutor.send() needs a connection: Java's getSend() is null until the admin
// is connected, and a null request is dropped before anything else happens.
func TestSendWithoutConnection(t *testing.T) {
	var fixture = newFixture()
	var admin = FakeAdmin{}
	var disconnected = gamescript.NewGSExecutor(&admin)
	var request = gamescript.NewGSJsonRequest(constant.GS_COMMAND_COUNT_INDUSTRIE, []string{})
	disconnected.Send(request)
	disconnected.Send(nil)
	if len(fixture.sender.packets) != 0 {
		t.Errorf("nothing reaches the wire")
	}
	if request.HasResponse() {
		t.Errorf("a request without a connection stays pending")
	}
}

// The reply payload types AdminCmd's Squirrel can produce: a quoted string, a
// bare number, a bare bool. Java's gson mapped all of them into String fields,
// and a JSON null became a null String, which reads as no value here.
func TestReplyDataOfDifferentTypes(t *testing.T) {
	var fixture = newFixture()
	var payloads = []string{"true", "50000000000", "null", "\"\""}
	var expected = []string{"true", "50000000000", "", ""}
	for i := 0; i < len(payloads); i++ {
		var request = gamescript.NewGSJsonRequest(constant.GS_COMMAND_REMOVE_ALL_GOAL, []string{})
		fixture.executor.Send(request)
		fixture.executor.OnGameScriptReply("{\"id\":\"" + sentId(fixture, i) + "\",\"data\":" + payloads[i] + "}")
		var result, error = request.GetResult()
		if error != nil {
			t.Errorf("%v is a reply, not an exception", payloads[i])
		} else if result != expected[i] {
			t.Errorf("%v reads as %v", payloads[i], result)
		}
	}
}

// A newspaper request's result is Boolean.valueOf of the reply.
func TestNewsPaperResult(t *testing.T) {
	var fixture = newFixture()
	var request = gamescript.NewGSNewsPaper(gamescript.NT_GENERAL, "hello")
	fixture.executor.Send(request)
	fixture.executor.OnGameScriptReply("{\"id\":\"" + sentId(fixture, 0) + "\",\"data\":\"TRUE\"}")
	var created, error = request.GetResult()
	if error != nil || !created {
		t.Errorf("Boolean.valueOf is case insensitive")
	}
	var other = gamescript.NewGSNewsPaper(gamescript.NT_GENERAL, "hello")
	fixture.executor.Send(other)
	fixture.executor.OnGameScriptReply("{\"id\":\"" + sentId(fixture, 1) + "\",\"data\":\"false\"}")
	var rejected, rejectedError = other.GetResult()
	if rejectedError != nil || rejected {
		t.Errorf("a rejected newspaper reads false")
	}
}

// Proof the executor survived a bad reply: it still numbers and sends.
func assertStillSending(fixture *Fixture, t *testing.T) {
	var request = gamescript.NewGSJsonRequest(constant.GS_COMMAND_CREATE_NEWS, []string{})
	fixture.executor.Send(request)
	if len(fixture.sender.packets) == 0 {
		t.Errorf("the executor still sends")
	}
	if request.GetId() == "" {
		t.Errorf("the executor still numbers requests")
	}
}

// atoi is Integer.parseInt, for the ids the executor put on the wire.
func atoi(text string) int {
	var value, error = strconv.Atoi(text)
	if error != nil {
		return 0
	}
	return value
}

// The JSON layer both ways: what gson wrote, and what it read back. AdminCmd's
// JSONDecode is the reader on the other end, and it only needs the three keys.
func TestJsonRoundTrip(t *testing.T) {
	// gson escapes the quote, the backslash and the control characters, and
	// leaves every other byte alone, so UTF-8 reaches the wire unchanged.
	if gamescript.JsonString("say \"hi\"\tpath\\x") != "\"say \\\"hi\\\"\\tpath\\\\x\"" {
		t.Errorf("the escaping is %v", gamescript.JsonString("say \"hi\"\tpath\\x"))
	}
	if gamescript.JsonString("café") != "\"café\"" {
		t.Errorf("UTF-8 passes through, got %v", gamescript.JsonString("café"))
	}
	if gamescript.JsonString("") != "\"\"" {
		t.Errorf("an empty message")
	}
	// A control character becomes \u00XX, the way gson writes it.
	if gamescript.JsonString("\u0001") != "\"\\u0001\"" {
		t.Errorf("a control character is %v", gamescript.JsonString("\u0001"))
	}
	// A number above 2^31 is what AdminCmd's money values look like, and gson
	// kept the text it parsed, so it must not come back as an exponent.
	var fields = gamescript.RawFields("{\"id\":\"9\",\"data\":50000000000}")
	if fields == nil {
		t.Errorf("a reply with a bare number is readable")
		return
	}
	var data, hasData = fields["data"]
	if !hasData || data != "50000000000" {
		t.Errorf("the number reads as %v", data)
	}
	// Whitespace, a nested value and an escaped quote in the data.
	var messy = gamescript.RawFields(" { \"data\" : \"a\\\"b\", \"id\" : 3, \"x\": [1,2] } ")
	if messy == nil {
		t.Errorf("a spaced out reply is readable")
		return
	}
	if messy["data"] != "a\"b" || messy["id"] != "3" {
		t.Errorf("the fields read back as %v and %v", messy["data"], messy["id"])
	}
	if messy["x"] != "[1,2]" {
		t.Errorf("a nested value is kept verbatim, got %v", messy["x"])
	}
	// Not an object at all: gson would have thrown, and the executor treats a
	// nil result as "nothing to correlate".
	if gamescript.RawFields("not json") != nil {
		t.Errorf("a non object reply")
	}
	if gamescript.RawFields("{\"id\":") != nil {
		t.Errorf("a truncated reply")
	}
	if gamescript.RawFields("{}") == nil {
		t.Errorf("an empty object is still an object")
	}
}
