// NetworkClientTest has no Java counterpart: the only test of this layer in the
// Java tree is OpenttdClientIntegrationTest, which needs a real dedicated server
// and is excluded from gradle's integrationTest because the Java PacketGameType
// ids are stale for 15.3 (the join is refused with NETWORK_ERROR_WRONG_REVISION).
// So the assertions below are the wire format the Java source itself states:
// every write* in NetworkClientSender, every read* in NetworkClient.handlePacket,
// and the hashPassword oracle, pinned against a scripted game server.
package client_test

import "testing"
import "time"
import "com.openttd"
import "com.openttd/network/client"
import "com.openttd/network/constant"
import "com.openttd/network/core"
import "com.openttd/server"

// SETTLE_MILLIS is how long to wait before asserting that the client sent nothing
// more, the same idiom the event-mapping table uses for its "none" rows: the
// await only proves a packet arrived, not that no second one did.
const SETTLE_MILLIS = 200

// newHarness is the @BeforeEach of the admin tests' twin: a fake game server on
// an ephemeral port, a real NetworkClient connected to it, and the join already
// answered with SERVER_WELCOME, so the model has its client id and server id.
func newHarness(t *testing.T) *server.ClientHarness {
	var harness, harnessError = server.NewClientHarness()
	if harnessError != nil {
		t.Errorf("ClientHarness: %s", core.Message(harnessError))
		return nil
	}
	return harness
}

// awaitPackets waits until the fake server has seen count packets of one type
// from the client, then hands them back oldest first.
func awaitPackets(srv *server.FakeGameServer, packetType constant.PacketGameType, count int, timeoutSeconds int) []*server.ReceivedGamePacket {
	var name = "PACKET_"+constant.PacketGameTypeName(packetType)
	openttd.Await(func() bool { return len(srv.Packets(packetType)) >= count }, timeoutSeconds, name)
	return srv.Packets(packetType)
}

// PACKET_CLIENT_JOIN: string version, uint32 newgrf checksum, string name,
// uint8 PlayAs, uint8 language.
func TestJoin(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	var join = srv.AwaitPacket(constant.PACKET_CLIENT_JOIN, 10)
	if join == nil {
		t.Errorf("no PACKET_CLIENT_JOIN within 10s")
		return
	}
	var payload = join.GetPacket()
	if payload.ReadString() != constant.OPENTTD {
		t.Errorf("join: expected the version string %s", constant.OPENTTD)
	}
	if payload.ReadUint32() != uint32(constant.NEWGRF) {
		t.Errorf("join: expected the newgrf checksum %d", uint32(constant.NEWGRF))
	}
	if payload.ReadString() != harness.GetConfiguration().Name {
		t.Errorf("join: expected the client name %s", harness.GetConfiguration().Name)
	}
	if payload.ReadUint8() != 255 {
		t.Errorf("join: expected the PlayAs mask 255")
	}
	if payload.ReadUint8() != 0 {
		t.Errorf("join: expected the retired language field 0")
	}
}

// PACKET_SERVER_WELCOME: uint32 client_id, uint32 map_seed, string server_id.
// The client keeps all three and answers with PACKET_CLIENT_GETMAP.
func TestWelcomeIsTheHandshake(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()
	var model = harness.GetModel()

	// The harness's handshake is the only welcome so far, so the model holds
	// exactly what FakeGameServer.GameWelcomePacket wrote.
	if model.ClientId != int64(1) {
		t.Errorf("welcome: expected client id 1, got %d", model.ClientId)
	}
	if model.PasswordGameSeed != int64(server.GAME_MAP_SEED) {
		t.Errorf("welcome: expected the map seed %d, got %d", server.GAME_MAP_SEED, model.PasswordGameSeed)
	}
	if model.PasswordServerId != server.GAME_SERVER_ID {
		t.Errorf("welcome: expected the server id %s, got %s", server.GAME_SERVER_ID, model.PasswordServerId)
	}

	var asked = awaitPackets(srv, constant.PACKET_CLIENT_GETMAP, 1, 10)
	if len(asked) != 1 {
		t.Errorf("getmap: expected 1, got %d", len(asked))
	}
	if asked[0].GetPacket().Remaining() != 0 {
		t.Errorf("getmap: expected an empty payload")
	}
}

// PACKET_SERVER_CHECK_NEWGRFS: uint8 count, then count x (uint32 grf_id,
// uint32 md5). The client has every NewGRF, so it answers
// PACKET_CLIENT_NEWGRFS_CHECKED with an empty payload.
func TestCheckNewgrfsAnswersChecked(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	var list = server.New(constant.PACKET_SERVER_CHECK_NEWGRFS)
	list.WriteUint8(2)
	list.WriteUint32(uint32(0x12345678))
	list.WriteUint32(uint32(0x9ABCDEF0))
	list.WriteUint32(uint32(0x00000001))
	list.WriteUint32(uint32(0x00000002))
	srv.Send(list)

	var checked = awaitPackets(srv, constant.PACKET_CLIENT_NEWGRFS_CHECKED, 1, 10)
	if len(checked) != 1 {
		t.Errorf("newgrfs checked: expected 1, got %d", len(checked))
	}
	if checked[0].GetPacket().Remaining() != 0 {
		t.Errorf("newgrfs checked: expected an empty payload")
	}
}

// PACKET_SERVER_NEED_GAME_PASSWORD has no payload: the client sends back
// PACKET_CLIENT_GAME_PASSWORD with the password from its configuration, unhashed
// (upstream hashes it in the game, not here).
func TestNeedGamePasswordSendsTheConfiguredOne(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	srv.Send(server.New(constant.PACKET_SERVER_NEED_GAME_PASSWORD))

	var sent = awaitPackets(srv, constant.PACKET_CLIENT_GAME_PASSWORD, 1, 10)
	if len(sent) != 1 {
		t.Errorf("game password: expected 1, got %d", len(sent))
	}
	if sent[0].GetPacket().ReadString() != harness.GetConfiguration().Password {
		t.Errorf("game password: expected %s", harness.GetConfiguration().Password)
	}
}

// PACKET_SERVER_NEED_COMPANY_PASSWORD: uint32 game_seed, string server_id. The
// client keeps both and answers PACKET_CLIENT_COMPANY_PASSWORD with the salted
// MD5 of its own password, which is what upstream's GenerateCompanyPasswordHash
// produces.
func TestNeedCompanyPasswordSendsTheHash(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	var ask = server.New(constant.PACKET_SERVER_NEED_COMPANY_PASSWORD)
	ask.WriteUint32(uint32(123456789))
	ask.WriteString("server-id-0123456789abcdefghij")
	srv.Send(ask)

	var sent = awaitPackets(srv, constant.PACKET_CLIENT_COMPANY_PASSWORD, 1, 10)
	if len(sent) != 1 {
		t.Errorf("company password: expected 1, got %d", len(sent))
	}
	var model = harness.GetModel()
	if model.PasswordGameSeed != int64(123456789) {
		t.Errorf("company password: expected the game seed 123456789, got %d", model.PasswordGameSeed)
	}
	if model.PasswordServerId != "server-id-0123456789abcdefghij" {
		t.Errorf("company password: expected the server id, got %s", model.PasswordServerId)
	}
	var hashed = sent[0].GetPacket().ReadString()
	var expected = client.HashPassword(harness.GetConfiguration().Password,
		"server-id-0123456789abcdefghij", int64(123456789))
	if hashed != expected {
		t.Errorf("company password: expected the hash %s, got %s", expected, hashed)
	}
}

// PACKET_SERVER_MAP_DONE has no payload: the client says the map arrived with
// PACKET_CLIENT_MAP_OK.
func TestMapDoneAcceptsTheMap(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	srv.Send(server.New(constant.PACKET_SERVER_MAP_DONE))

	var ok = awaitPackets(srv, constant.PACKET_CLIENT_MAP_OK, 1, 10)
	if len(ok) != 1 {
		t.Errorf("map ok: expected 1, got %d", len(ok))
	}
	if ok[0].GetPacket().Remaining() != 0 {
		t.Errorf("map ok: expected an empty payload")
	}
}

// PACKET_SERVER_FRAME: uint32 frame_counter_server, uint32 frame_counter_max,
// then optionally the sync seeds and a uint8 token. The client acks the frame
// only when it has not already acked one that far ahead: last_ack_frame becomes
// frame + DAY_TICKS, which is the whole point of the rule.
func TestFrameAcksOncePerDay(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()
	var model = harness.GetModel()

	var first = server.New(constant.PACKET_SERVER_FRAME)
	first.WriteUint32(uint32(1000))
	first.WriteUint32(uint32(1000))
	var behind = server.New(constant.PACKET_SERVER_FRAME)
	behind.WriteUint32(uint32(1001))
	behind.WriteUint32(uint32(1001))
	var ahead = server.New(constant.PACKET_SERVER_FRAME)
	ahead.WriteUint32(uint32(1075))
	ahead.WriteUint32(uint32(1075))
	ahead.WriteUint8(3)
	srv.SendAll([]*server.PacketWriter{first, behind, ahead})

	// 1000 is acked; 1001 is inside the DAY_TICKS window, so it is not; 1075	// is past 1000 + 74, so it is.
	awaitPackets(srv, constant.PACKET_CLIENT_ACK, 2, 10)
	time.Sleep(SETTLE_MILLIS * time.Millisecond)
	var acks = srv.Packets(constant.PACKET_CLIENT_ACK)
	if len(acks) != 2 {
		t.Errorf("acks: expected 2, got %d", len(acks))
		return
	}
	// GetPacket re-reads the header, so the payload has to be read from one
	// local: calling it again would start the fields over.
	var firstAck = acks[0].GetPacket()
	if firstAck.ReadUint32() != uint32(1000) {
		t.Errorf("first ack: expected frame 1000")
	}
	if firstAck.ReadUint8() != 0 {
		t.Errorf("first ack: expected the token 0, since FRAME carried none")
	}
	var secondAck = acks[1].GetPacket()
	if secondAck.ReadUint32() != uint32(1075) {
		t.Errorf("second ack: expected frame 1075")
	}
	if secondAck.ReadUint8() != 3 {
		t.Errorf("second ack: expected the token 3 from the packet")
	}
	if model.LastAckFrame != int64(1075+client.DAY_TICKS) {
		t.Errorf("last_ack_frame: expected %d, got %d", 1075+client.DAY_TICKS, model.LastAckFrame)
	}
	if model.FrameCounterServer != int64(1075) {
		t.Errorf("frame_counter_server: expected 1075, got %d", model.FrameCounterServer)
	}
	if model.Token != 3 {
		t.Errorf("token: expected 3, got %d", model.Token)
	}
}

// The optional FRAME fields are read only when the packet still has room for
// them: two sync seeds and the token come after the two frame counters, and a
// packet that stops short must not consume the token byte of the next field.
func TestFrameWithSyncSeeds(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()
	var model = harness.GetModel()

	var both = server.New(constant.PACKET_SERVER_FRAME)
	both.WriteUint32(uint32(3000))
	both.WriteUint32(uint32(3000))
	both.WriteUint32(uint32(11))
	both.WriteUint32(uint32(22))
	both.WriteUint8(9)
	var one = server.New(constant.PACKET_SERVER_FRAME)
	one.WriteUint32(uint32(4000))
	one.WriteUint32(uint32(4000))
	one.WriteUint32(uint32(33))
	srv.SendAll([]*server.PacketWriter{both, one})

	// Both frames are far enough ahead to ack, and both must carry the token
	// read from the first packet: the second has no token of its own, so the
	// model keeps the old one.
	awaitPackets(srv, constant.PACKET_CLIENT_ACK, 2, 10)
	time.Sleep(SETTLE_MILLIS * time.Millisecond)
	var acks = srv.Packets(constant.PACKET_CLIENT_ACK)
	if len(acks) != 2 {
		t.Errorf("acks: expected 2, got %d", len(acks))
		return
	}
	var firstAck = acks[0].GetPacket()
	if firstAck.ReadUint32() != uint32(3000) {
		t.Errorf("first ack: expected frame 3000")
	}
	if firstAck.ReadUint8() != 9 {
		t.Errorf("first ack: expected the token 9 after both sync seeds")
	}
	var secondAck = acks[1].GetPacket()
	if secondAck.ReadUint32() != uint32(4000) {
		t.Errorf("second ack: expected frame 4000")
	}
	if secondAck.ReadUint8() != 9 {
		t.Errorf("second ack: expected the token 9 kept from the previous packet")
	}
	if model.Token != 9 {
		t.Errorf("token: expected 9, got %d", model.Token)
	}
}

// Every packet whose case only decodes the payload: the client must survive all
// of them, which it proves by still answering the MAP_DONE that closes the batch
// (an over-read would mark the packet broken, and a panic would take the
// listener thread with it).
func TestClientPackets(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()

	var batch = []*server.PacketWriter{
		server.New(constant.PACKET_SERVER_FULL),
		server.New(constant.PACKET_SERVER_BANNED),
		server.New(constant.PACKET_SERVER_CLIENT_INFO).
			WriteUint32(uint32(7)).WriteUint8(0).WriteString("Client"),
		server.New(constant.PACKET_SERVER_ERROR).WriteUint8(10),
		server.New(constant.PACKET_SERVER_WAIT).WriteUint8(1),
		server.New(constant.PACKET_SERVER_MAP_BEGIN).WriteUint32(uint32(100)),
		server.New(constant.PACKET_SERVER_MAP_SIZE).WriteUint32(uint32(10000)),
		server.New(constant.PACKET_SERVER_MAP_DATA),
		server.New(constant.PACKET_SERVER_SYNC).WriteUint32(uint32(100)).WriteUint32(uint32(11)),
		server.New(constant.PACKET_SERVER_COMMAND).WriteUint32(uint32(100)).WriteBool8(false),
		server.New(constant.PACKET_SERVER_CHAT).
			WriteUint8(2).WriteUint32(uint32(7)).WriteBool8(false).WriteString("hello").WriteUint64(uint64(30000000000)),
		server.New(constant.PACKET_SERVER_EXTERNAL_CHAT).
			WriteString("server").WriteUint16(0x800B).WriteString("user").WriteString("hi"),
		server.New(constant.PACKET_SERVER_JOIN).WriteUint32(uint32(8)),
		server.New(constant.PACKET_SERVER_SHUTDOWN),
		server.New(constant.PACKET_SERVER_NEWGAME),
		server.New(constant.PACKET_SERVER_RCON).WriteUint16(0x800B).WriteString("getdate"),
		server.New(constant.PACKET_SERVER_MOVE).WriteUint32(uint32(7)).WriteUint8(0),
		server.New(constant.PACKET_SERVER_CONFIG_UPDATE).WriteUint8(32).WriteString("name"),
		server.New(constant.PACKET_SERVER_COMPANY_UPDATE).WriteUint16(1),
		server.New(constant.PACKET_SERVER_MAP_DONE),
	}
	srv.SendAll(batch)

	var ok = awaitPackets(srv, constant.PACKET_CLIENT_MAP_OK, 1, 10)
	if len(ok) != 1 {
		t.Errorf("map ok: expected 1 after the batch, got %d", len(ok))
	}
	if !harness.GetClient().IsConnected() {
		t.Errorf("the client should still be connected after the batch")
	}
}

// The sender's own API, which no Java test ever called: the chat helpers pick
// the action and destination type upstream documents, and rcon puts the password
// on the wire before the command.
func TestChatAndRconWireFormat(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()
	var send = harness.GetSend()

	send.ChatClient(int64(7), "private")
	send.ChatCompany(3, "team")
	send.ChatBroadcast("all")
	send.ChatClient(int64(-1), "fallback")
	send.RCon("getdate", "secret")

	var chats = awaitPackets(srv, constant.PACKET_CLIENT_CHAT, 4, 10)
	if len(chats) != 4 {
		t.Errorf("chat: expected 4, got %d", len(chats))
		return
	}
	// chatClient(7): one client, so DESTTYPE_CLIENT.
	var first = chats[0].GetPacket()
	if first.ReadUint8() != int(constant.NETWORK_ACTION_SERVER_MESSAGE) {
		t.Errorf("chat 0: expected the server-message action")
	}
	if first.ReadUint8() != int(constant.DESTTYPE_CLIENT) {
		t.Errorf("chat 0: expected DESTTYPE_CLIENT")
	}
	if first.ReadUint32() != uint32(7) {
		t.Errorf("chat 0: expected the client id 7")
	}
	if first.ReadString() != "private" {
		t.Errorf("chat 0: expected the message private")
	}
	// chatCompany(3): DESTTYPE_TEAM, and the id is a company id.
	var second = chats[1].GetPacket()
	if second.ReadUint8() != int(constant.NETWORK_ACTION_SERVER_MESSAGE) {
		t.Errorf("chat 1: expected the server-message action")
	}
	if second.ReadUint8() != int(constant.DESTTYPE_TEAM) {
		t.Errorf("chat 1: expected DESTTYPE_TEAM")
	}
	if second.ReadUint32() != uint32(3) {
		t.Errorf("chat 1: expected the company id 3")
	}
	if second.ReadString() != "team" {
		t.Errorf("chat 1: expected the message team")
	}
	// chatBroadcast: DESTTYPE_BROADCAST to nobody in particular.
	var third = chats[2].GetPacket()
	if third.ReadUint8() != int(constant.NETWORK_ACTION_SERVER_MESSAGE) {
		t.Errorf("chat 2: expected the server-message action")
	}
	if third.ReadUint8() != int(constant.DESTTYPE_BROADCAST) {
		t.Errorf("chat 2: expected DESTTYPE_BROADCAST")
	}
	if third.ReadUint32() != uint32(0) {
		t.Errorf("chat 2: expected the destination 0")
	}
	if third.ReadString() != "all" {
		t.Errorf("chat 2: expected the message all")
	}
	// chatClient(-1): no client, so the broadcast fallback.
	var fourth = chats[3].GetPacket()
	if fourth.ReadUint8() != int(constant.NETWORK_ACTION_SERVER_MESSAGE) {
		t.Errorf("chat 3: expected the server-message action")
	}
	if fourth.ReadUint8() != int(constant.DESTTYPE_BROADCAST) {
		t.Errorf("chat 3: expected the broadcast fallback")
	}
	if fourth.ReadUint32() != uint32(0) {
		t.Errorf("chat 3: expected the fallback destination 0")
	}
	if fourth.ReadString() != "fallback" {
		t.Errorf("chat 3: expected the message fallback")
	}

	var rcons = awaitPackets(srv, constant.PACKET_CLIENT_RCON, 1, 10)
	if len(rcons) != 1 {
		t.Errorf("rcon: expected 1, got %d", len(rcons))
		return
	}
	var rcon = rcons[0].GetPacket()
	if rcon.ReadString() != "secret" {
		t.Errorf("rcon: expected the password first on the wire")
	}
	if rcon.ReadString() != "getdate" {
		t.Errorf("rcon: expected the command second on the wire")
	}
}

// IsQuitPacket: PACKET_CLIENT_QUIT is the last packet of a clean shutdown, so
// the sender thread stops after sending it, which is what Shutdown's join waits
// for in Java.
func TestQuitStopsTheSender(t *testing.T) {
	var harness = newHarness(t)
	if harness == nil {
		return
	}
	defer harness.Close()
	var srv = harness.GetServer()
	var send = harness.GetSend()

	send.Quit()

	var quit = srv.AwaitPacket(constant.PACKET_CLIENT_QUIT, 10)
	if quit == nil {
		t.Errorf("no PACKET_CLIENT_QUIT within 10s")
		return
	}
	if quit.GetPacket().Remaining() != 0 {
		t.Errorf("quit: expected an empty payload")
	}
	openttd.Await(func() bool { return !send.IsRunning() }, 5, "the sender thread to stop")
	if send.IsRunning() {
		t.Errorf("the sender should have stopped after the quit packet")
	}
}

// hashPassword, "see network.cpp, method: GenerateCompanyPasswordHash". The
// oracle is the upstream rule: the password padded with spaces to the length of
// the server id, every byte but the last xored with the server id and the
// game seed shifted by i % 32, then MD5 as lower-case hex.
func TestHashPassword(t *testing.T) {
	var cases = []struct {
		password string
		serverId string
		gameSeed int64
		expected string
	}{
		{"game-password", "server-id-0123456789abcdefghij", int64(123456789), "71d29740f5a236c6f9c92cf63238d6ee"},
		{"x", "ab", int64(7), "800618943025315f869e4e1f09471012"},
		{"a-much-longer-password-than-the-server-id", "0123456789", int64(-9876543210), "78d44da8b95db3c7ae9491034c74b957"},
	}
	for _, one := range cases {
		var hashed = client.HashPassword(one.password, one.serverId, one.gameSeed)
		if hashed != one.expected {
			t.Errorf("hashPassword(%s, %s, %d): expected %s, got %s", one.password, one.serverId,
				one.gameSeed, one.expected, hashed)
		}
	}
}

// An empty password is returned unchanged, which is the early return upstream's
// StrEmpty check stands for.
func TestHashPasswordWithoutPassword(t *testing.T) {
	var hashed = client.HashPassword("", "server-id-0123456789abcdefghij", int64(123456789))
	if hashed != "" {
		t.Errorf("empty password: expected it back unchanged, got %s", hashed)
	}
}
