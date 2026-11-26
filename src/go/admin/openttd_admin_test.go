// OpenttdAdmin facade tests: the wiring the library promises its callers -
// startup/shutdown, the listeners registered by AddListener, and the gamescript
// round trip through the executor.

package admin_test

import "testing"
import "time"
import "com.openttd/admin/event"
import "com.openttd/gamescript"
import "com.openttd/network/constant"
import "com.openttd/server"
import "com.openttd"

const SETTLE_MILLIS = 200

func TestStartupAndShutdown(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	if !harness.GetAdmin().IsConnected() {
		t.Errorf("the admin is not connected after startup")
	}
	var game = harness.GetAdmin().GetGame()
	if game == nil {
		t.Errorf("no game model after startup")
	}
	if game.GetGameInfo().GetServerName() != "Strategyboard test server" {
		t.Errorf("server name %s", game.GetGameInfo().GetServerName())
	}
	if harness.GetAdmin().GetSend() == nil {
		t.Errorf("no admin sender after startup")
	}
	if harness.GetAdmin().GetGSExecutor() == nil {
		t.Errorf("no gamescript executor after startup")
	}
	if harness.GetAdmin().GetSender() == nil {
		t.Errorf("GetSender() did not find the admin sender")
	}
	harness.Close()
	if harness.GetAdmin().IsConnected() {
		t.Errorf("still connected after shutdown")
	}
}

// A CHAT packet reaches the listeners the harness registered.
func TestChatReachesTheListeners(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var recorder = harness.GetRecorder()
	harness.GetServer().Send(chatPacket("hello"))
	openttd.Await(func() bool { return len(recorder.GetChatEvents()) == 1 }, 5, "the chat event")
	var chats = recorder.GetChatEvents()
	var chat = chats[0]
	if chat.GetClientId() != 7 || chat.GetMessage() != "hello" {
		t.Errorf("chat %d/%s", chat.GetClientId(), chat.GetMessage())
	}
	if chat.GetOpenttd() == nil {
		t.Errorf("the chat event has no Openttd")
	}
}

// AddListener and RemoveListener drive the dispatcher's listener map.
func TestAddAndRemoveListener(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var listener = &ConsoleListener{}
	harness.GetAdmin().AddListener(event.EVENT_CONSOLE, listener)
	harness.GetServer().Send(consolePacket("say hi"))
	openttd.Await(func() bool { return len(listener.messages) == 1 }, 5, "the console event")
	harness.GetAdmin().RemoveListener(event.EVENT_CONSOLE, listener)
	harness.GetServer().Send(consolePacket("say again"))
	time.Sleep(SETTLE_MILLIS * time.Millisecond)
	if len(listener.messages) != 1 {
		t.Errorf("listener saw %d events, expected 1", len(listener.messages))
	}
}

// The executor numbers requests, sends them as ADMIN_PACKET_ADMIN_GAMESCRIPT
// and matches the server's reply by id. Startup() clears the ids, so the first
// request of a test is "0".
func TestGamescriptRoundTrip(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var executor = harness.GetAdmin().GetGSExecutor()
	var request = gamescript.NewGSJsonRequest(constant.GS_COMMAND_COUNT_INDUSTRIE, []string{"0"})
	executor.Send(request)
	if request.GetId() != "0" {
		t.Errorf("request id %s, expected 0", request.GetId())
	}
	var sent = harness.GetServer().AwaitPacket(constant.ADMIN_PACKET_ADMIN_GAMESCRIPT, 5)
	if sent == nil {
		return
	}
	var payload = sent.GetPacket().ReadString()
	if payload != request.ToString() {
		t.Errorf("sent %s, expected %s", payload, request.ToString())
	}
	if request.HasResponse() {
		t.Errorf("the reply arrived before the server sent one")
	}
	harness.GetServer().Send(gamescriptReply("0", "42"))
	openttd.Await(func() bool { return request.HasResponse() }, 5, "the gamescript reply")
	var result, exception = request.GetResult()
	if result != "42" {
		t.Errorf("result %s, expected 42", result)
	}
	if exception != nil {
		t.Errorf("unexpected exception %s", exception.ToString())
	}
}

// A reply for an id nobody asked about is logged and dropped, as Java's
// "Ignoring gamescript reply" branch does, and the event still reaches the
// listeners.
func TestGamescriptReplyWithoutRequest(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var recorder = harness.GetRecorder()
	harness.GetServer().Send(gamescriptReply("99", "42"))
	time.Sleep(SETTLE_MILLIS * time.Millisecond)
	if len(recorder.GetGameScriptEvents()) != 1 {
		t.Errorf("the gamescript event was not dispatched")
	}
}

type ConsoleListener struct {
	messages []string
}

func (l *ConsoleListener) OnConsoleEvent(consoleEvent *event.ConsoleEvent) {
	l.messages = append(l.messages, consoleEvent.GetMessage())
}

func newHarness(t *testing.T) *server.AdminHarness {
	var harness, err = server.NewAdminHarness()
	if err != nil {
		t.Errorf("harness: %s", err)
		return nil
	}
	return harness
}

func chatPacket(message string) *server.PacketWriter {
	var packet = server.New(constant.ADMIN_PACKET_SERVER_CHAT)
	packet.WriteUint8(0) // ChatAction
	packet.WriteUint8(0) // DestType
	packet.WriteUint32(7)
	packet.WriteString(message)
	packet.WriteUint64(0) // money
	return packet
}

func consolePacket(message string) *server.PacketWriter {
	var packet = server.New(constant.ADMIN_PACKET_SERVER_CONSOLE)
	packet.WriteString("::1")
	packet.WriteString(message)
	return packet
}

// The server's reply for a gamescript request: ADMIN_PACKET_SERVER_GAMESCRIPT
// carrying {"id":..,"data":..}.
func gamescriptReply(id string, data string) *server.PacketWriter {
	var packet = server.New(constant.ADMIN_PACKET_SERVER_GAMESCRIPT)
	var payload = "{\"id\":" + gamescript.JsonString(id) + ",\"data\":" + gamescript.JsonString(data) + "}"
	packet.WriteString(payload)
	return packet
}
