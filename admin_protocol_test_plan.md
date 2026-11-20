# Admin protocol test plan

Cross-checked against the protocol actually spoken by the binary in use (OpenTTD 15.3):
`git show 15.3:src/network/network_admin.cpp`, `git show 15.3:src/network/core/tcp_admin.h`,
`git show 15.3:src/network/core/config.h`, `git show 15.3:src/network/network_type.h`.
`/home/eric/www/OpenTTD` is upstream master (16.0-dev), so always read the `15.3` tag.

## Status: Phases 0-3 done, committed and published

| Phase | State | Evidence |
|---|---|---|
| 0 protocol fixes | done | commit `517cc2d` "Fix admin protocol handling for OpenTTD 15.3" (14 files) |
| 1 offline tests | done | commit `9eb8d52`, `gradle test` = 108 tests, 0 failures, ~1m22s, no OpenTTD process |
| 2 integration tests | done | commit `9eb8d52`, `gradle integrationTest` = 25 tests, 0 failures, ~53s against a real dedicated server (see AGENTS.md for the two prerequisites) |
| 3 build and test wiring | done | `build.gradle` `test`/`integrationTest` filters, `AGENTS.md` rewritten, both jars published |

`gradle jar` = 153 KB, copied to `../strategyboard/lib/openttd-admin.jar` and
`../openttd-admin/openttd-admin.jar` (md5 `f36ab3dd1335d7fd3a0890913a07e3b8`). strategyboard is
green against it (`play check` and `play precompile` exit 0). The gamescript fix is committed
separately in `../AdminCmd` as `c12f136`.

`admin_protocol_test_plan.md` itself is intentionally untracked - it is the working record, not
part of the library.

Fixes beyond the list in section 4, found by the integration runs:
- `GameInfo.version` was always 0: the PROTOCOL version is now stored on the model and carried
  into WELCOME, which replaces the GameInfo (`NetworkAdmin.handlePacket`, asserted offline too).
- `src/squirrel/main.nut` / `../AdminCmd/src/main.nut`: `removeAllGoal` used `data | GSGoal.Remove(...)`
  which Squirrel rejects ("bitwise op between 'bool' and 'bool'"), so it is `||` in both copies.
- `Convert` matches upstream for game years 1-2050 but maps game year 0 to Java year 1; documented
  in `ConvertTest`, not fixed (upstream `CalendarConvertDateToYMD` has no year-0 game date).

Expectations pinned to upstream where the plan above guessed wrong:
- An admin's own chat is never echoed as `SERVER_CHAT` (`NetworkAdminChat` returns on `from_admin`),
  so the chat tests assert the console relay instead.
- `CLIENT_ID_SERVER = 1`, so a dedicated server has exactly client #1 and id 0 is
  `INVALID_CLIENT_ID` (chat to it prints nothing).
- Console relay origin is the debug category, so gamescript lines arrive as `Console (script)`, and
  Squirrel prints arguments as `(array : 0x...)`.
- Illegal frequency / illegal chat colour are asserted as "the connection drops and reconnects",
  not as an `SERVER_ERROR` event: `SendError` closes the socket before Java can read anything else.

## 1. Protocol inventory (admin protocol version 3)

**Admin -> server** (`tcp_admin.h`, ids 0-10): `JOIN 0`, `QUIT 1`, `UPDATE_FREQUENCY 2`, `POLL 3`,
`CHAT 4`, `RCON 5`, `GAMESCRIPT 6`, `PING 7`, `EXTERNAL_CHAT 8`, `JOIN_SECURE 9`, `AUTH_RESPONSE 10`.
Java declares all 11; 9-10 are constants only (`ADMIN_PACKET_ADMIN_JOIN_SECURE`,
`ADMIN_PACKET_ADMIN_AUTH_RESPONSE`) - the secure handshake is not implemented.

**Server -> admin** (ids 100-129): `FULL 100`, `BANNED 101`, `ERROR 102`, `PROTOCOL 103`,
`WELCOME 104`, `NEWGAME 105`, `SHUTDOWN 106`, `DATE 107`, `CLIENT_JOIN 108`, `CLIENT_INFO 109`,
`CLIENT_UPDATE 110`, `CLIENT_QUIT 111`, `CLIENT_ERROR 112`, `COMPANY_NEW 113`, `COMPANY_INFO 114`,
`COMPANY_UPDATE 115`, `COMPANY_REMOVE 116`, `COMPANY_ECONOMY 117`, `COMPANY_STATS 118`, `CHAT 119`,
`RCON 120`, `CONSOLE 121`, `CMD_NAMES 122`, `CMD_LOGGING_OLD 123`, `GAMESCRIPT 124`, `RCON_END 125`,
`PONG 126`, `CMD_LOGGING 127`, `AUTH_REQUEST 128`, `ENABLE_ENCRYPTION 129`.

**Frequencies per update type** (`network_admin.cpp:48-59`, `_admin_update_type_frequencies`):
`DATE` = Poll|Daily|Weekly|Monthly|Quarterly|Annually, `CLIENT_INFO` = Poll|Automatic,
`COMPANY_INFO` = Poll|Automatic, `COMPANY_ECONOMY` = Poll|Weekly|Monthly|Quarterly|Annually,
`COMPANY_STATS` = Poll|Weekly|Monthly|Quarterly|Annually, `CHAT` = Automatic, `CONSOLE` = Automatic,
`CMD_NAMES` = Poll, `CMD_LOGGING` = Automatic, `GAMESCRIPT` = Automatic.

**Field limits** (`config.h`, bytes including `\0`): revision 33, password 33, client/admin name 25,
rcon command 500, gamescript JSON 9000, chat 900. `COMPAT_MTU` 1460, `TCP_MTU` 32767.

## 2. Coverage before this work (3 admin tests)

| Packet | before | gap |
|---|---|---|
| `ADMIN_PING` / `SERVER_PONG` | `testPingPong` - count only | ping id echo unasserted |
| `ADMIN_RCON` / `SERVER_RCON` / `SERVER_RCON_END` | `testRCon` - count only | colour, output text, command echo unasserted |
| `ADMIN_GAMESCRIPT` / `SERVER_GAMESCRIPT` | `testGameScript` - `createNews` only | other 5 RPCs unasserted |
| `ADMIN_JOIN` / `QUIT` / `UPDATE_FREQUENCY` / `POLL` | implicit via `Game.updateAdmin` WELCOME handler | never asserted; wrong password, illegal frequency, `pollCompanyStats`, `pollCmdNames` never exercised |
| `ADMIN_CHAT`, `ADMIN_EXTERNAL_CHAT` | - | untested |
| `SERVER_PROTOCOL`, `WELCOME`, `DATE`, `CLIENT_*`, `COMPANY_*`, `CHAT`, `CONSOLE`, `CMD_NAMES`, `CMD_LOGGING`, `ERROR`, `FULL`, `BANNED`, `NEWGAME`, `SHUTDOWN` | - | decoded in `NetworkAdmin.handlePacket`, no assertions |

Every row is covered by Phase 1 (offline decode + sender bytes) and Phase 2 (real server), except the
secure-auth packets and the game-port packets.

## 3. Protocol mismatches found — all eight fixed in `517cc2d`

1. `PacketServerType.ADMIN_PACKET_SERVER_CMD_LOGGING` is wrong twice: wire id 123 (15.3 sends 127,
   123 is `CMD_LOGGING_OLD`) and the decode layout - `NetworkAdmin.java` reads
   `p1/p2/tile/text`, while 15.3 sends `uint32 client, uint8 company, uint16 cmd, <variable buffer>,
   uint32 frame`. `INVALID_ADMIN_PACKET` currently sits at ordinal 27 -> wire 127, so real logging
   traffic decodes as invalid, throws `NetworkException` and drops the connection into the 5 s
   reconnect loop.
2. `Game.updateAdmin` falls through: `SERVER_RCON_END` has no `break`, so every rcon also dispatches a
   `ConsoleEvent` with `origin == null` (`Game.java:152-158`).
3. `SERVER_COMPANY_STATS` is decoded but never dispatched as a `CompanyEvent`.
4. `Packet.readUint64` multiplies by `4294967295` instead of `2^32` (`Packet.java:130`), so loan,
   company value and "give money" amounts above 2^32 are wrong.
5. `Packet.readString` decodes bytes as ISO-8859-1 while `writeString` encodes UTF-8.
6. `NetworkAdminSender.updateFrequency` falls back to `ADMIN_FREQUENCY_POLL`; `CHAT`, `CONSOLE`,
   `CMD_LOGGING` and `GAMESCRIPT` have no `Poll` bit, so the server replies
   `NETWORK_ERROR_ILLEGAL_PACKET` and closes the connection.
7. `NetworkType.NetworkErrorCode` misses `NETWORK_ERROR_NOT_ON_ALLOW_LIST` and
   `NETWORK_ERROR_NO_AUTHENTICATION_METHOD_AVAILABLE` (15.3 appends them before `END`).
8. Goal/town gamescript helpers (`addCompanyGoal`, `removeGoal`, `removeAllGoal`, `setTownCargoGoal`)
   omit `GameScript.ID`, so AdminCmd answers `Malformed json.`; `GSExecutor` then NPEs on a reply with
   no id. `src/squirrel/main.nut`'s `removeGoal` also references an undefined `goalId`.

Out of scope (reported, not fixed): secure admin auth handshake (`JOIN_SECURE`/`AUTH_RESPONSE`,
locally `allow_insecure_admin_login = true` - the constants and wire ids are declared and asserted,
the handshake is not implemented), the stale `TcpGame.PacketGameType` enum (this makes
`OpenttdClientIntegrationTest` crash the dedicated server - `NOT_REACHED` in
`Receive_CLIENT_AUTH_RESPONSE`), `GSNewsPaper.NewsType` ordinals being off by one from 15.3's
`::NewsType` (missing `AccidentOther`), and `Convert`'s game-year-0 mapping (see section "Status").

`../openttd-robot` does not compile against the current admin API: it expects
`Company.getPerformance()/getValue()/getInauguration()`, `int` money where the model returns `long`,
and a multi-arg `AbstractRule` super (`openttd-admin`'s is `AbstractRule(OpenttdAdmin)`). That drift
predates this work - no public method was removed by `517cc2d`, and the root jar `../openttd-admin/
openttd-admin.jar` did not exist before this project, so robot could not compile at all. Its tree is
clean at `a569117`; porting `robot/rule/*` is a separate task.

## 4. Phase 0 - protocol fixes (done)

1. `TcpAdmin.PacketServerType`: rename ordinal 23 to `ADMIN_PACKET_SERVER_CMD_LOGGING_OLD`, declare
   `CMD_LOGGING` (127), `AUTH_REQUEST` (128), `ENABLE_ENCRYPTION` (129), move `INVALID_ADMIN_PACKET`
   to the end. Add `ADMIN_PACKET_ADMIN_JOIN_SECURE` (9) and `ADMIN_PACKET_ADMIN_AUTH_RESPONSE` (10)
   to `PacketAdminType` (constants only, no handshake).
2. `NetworkAdmin.handlePacket`: correct `CMD_LOGGING` fields, decode `CMD_LOGGING_OLD`, add
   `AUTH_REQUEST` / `ENABLE_ENCRYPTION` cases, keep the decoded company-remove reason and client error
   code on `NetworkAdminEvent` instead of discarding them.
3. `Packet`: `readUint64` uses `2^32`, `readString` decodes UTF-8, add `remaining()`.
4. `Game.updateAdmin`: add the missing `break` after `RCON_END`, dispatch `CompanyEvent UPDATE` for
   `COMPANY_STATS`.
5. `NetworkAdminSender.updateFrequency`: fall back to the first frequency the server advertises for
   that type (`Protocol.firstSupported`), and skip the packet when the type advertises none.
6. `NetworkType.NetworkErrorCode`: append the two missing codes.
7. `GSExecutor`: ignore gamescript replies without a correlatable id (no NPE). Goal/town helpers send
   their JSON through `GSExecutor`, so `i` is present.
8. `src/squirrel/main.nut` + `../AdminCmd/src/main.nut`: `removeGoal` uses `args[0]`.

## 5. Phase 1 - offline tests (no OpenTTD needed) — done, 108 tests

Counts per suite: `NetworkAdminServerPacketsTest` 33, `NetworkAdminSenderTest` 28,
`GSExecutorTest` 10, `TcpAdminWireFormatTest` 11, `PacketCodecTest` 11, `ProtocolTest` 7,
`GameEventMappingTest` 4, `ConvertTest` 4.

- `test/java/com/openttd/network/constant/TcpAdminWireFormatTest.java` - wire-id parity table copied
  from `15.3:tcp_admin.h` / `config.h` / `network_type.h` (tag cited in comments): both packet enums,
  `AdminUpdateType`, `AdminUpdateFrequency` masks, `AdminCompanyRemoveReason`, `NetworkAction`,
  `DestType`, `NetworkVehicleType`, `NetworkErrorCode`, unknown-id sentinels.
- `test/java/com/openttd/network/core/PacketCodecTest.java` - `[len16][type8]` framing, little-endian
  reads, negative int64, uint64 above 2^32, UTF-8 round-trip, `bool8`, partial-packet reassembly
  through `Socket.receive`, upstream field limits.
- `test/java/com/openttd/network/admin/ProtocolTest.java` - synthetic `SERVER_PROTOCOL` built like
  `SendProtocol()`: version 3, all 10 frequency masks, `hasProtocol` matrix, `toString`.
- `test/java/com/openttd/server/FakeAdminServer.java` - in-JVM admin server on an ephemeral port, its
  own little-endian encoder mirroring upstream `Packet::Send_*`, consumes `JOIN`, replies `PROTOCOL` +
  `WELCOME`, then scripted packets; records everything the client sends.
- `test/java/com/openttd/network/admin/NetworkAdminServerPacketsTest.java` - one test per server packet
  (all 27 ids): decoded `GameInfo`, `Client`, `Company` (share owners, bankruptcy quarters, money/loan/
  income/values, vehicle and station arrays), dispatched event and payload, plus `ERROR`, `FULL`,
  `BANNED`, `SHUTDOWN`, `NEWGAME`, multi-packet `CMD_NAMES`, `CMD_LOGGING` vs `CMD_LOGGING_OLD`.
- `test/java/com/openttd/network/admin/NetworkAdminSenderTest.java` - drive the real sender against the
  fake server and assert exact bytes for `JOIN`, `QUIT`, `UPDATE_FREQUENCY`, `POLL` (all six helpers,
  `UINT32_MAX` for "all"), `CHAT` (broadcast/team/client), `EXTERNAL_CHAT`, `RCON`, `GAMESCRIPT`,
  `PING`, plus protocol-guard paths.
- `test/java/com/openttd/admin/model/GameEventMappingTest.java` - table-driven `Game.updateAdmin`
  mapping; pins "RCON_END emits only RConEndEvent" and "COMPANY_STATS emits CompanyEvent UPDATE".
- `test/java/com/openttd/gamescript/GSExecutorTest.java` - id correlation, `data` vs `exception`
  replies, replies without an id, JSON shape `{"c","a","i"}` for all six `GSCommand`s.
- Also added: `PacketWriter` (upstream framing), `AdminHarness` (fake server + `OpenttdAdmin`),
  `EventRecorder` (all nine listeners + `clear()`), `UpstreamCalendar` (`CalendarConvertDateToYMD`)
  and `test/java/com/openttd/util/ConvertTest.java` (date mapping over 1900-2050).

## 6. Phase 2 - integration tests against the real 15.3 server — done, 22 tests

`OpenttdAdminProtocolIntegrationTest extends OpenttdServerWrapper` - 22 tests, using
`TestUtils.await(predicate, timeoutSeconds)` / `awaitSize(...)` instead of fixed sleeps:

- join: `PROTOCOL` version 3 lands in `GameInfo.version`, `WELCOME` fills server name, revision vs
  `GameVersion.OPENTTD`, dedicated flag, square map, seed, start year 1950, and the current date once
  the `DATE` poll answers. A captured `GameInfo` reference goes stale (`Game.updateAdmin` swaps the
  network model per packet), so it is re-read inside the awaited condition.
- wrong password: every reconnect is refused, so no WELCOME and no PONG ever arrive; asserting the
  absence is the only thing the socket allows (`SendError` closes the connection).
- `update_frequency`: every advertised type/frequency pair from `UpstreamAdminProtocol.FREQUENCIES`;
  an unadvertised one (`COMPANY_INFO` + `ANUALLY`) falls back and the connection survives.
- polls: `pollDate`, `pollClientInfo(UINT32_MAX)` and `pollClientInfo(1)` -> client #1, spectator
  company 255; `pollCompanyInfo`/`pollCompanyEconomy`/`pollCompanyStats` are accepted and answer
  nothing (a dedicated server has no companies); `pollCmdNames` -> over 10 command names.
- chat: broadcast and `chatClient(1, ...)` are printed on the server console and never come back as
  `ChatEvent`; `chatClient(0, ...)` (invalid id) prints nothing; team chat to empty company 0 prints
  nothing; UTF-8 text survives `StrMakeValid` unchanged.
- `externalChat`: `CC_INFO` is printed, `TC_INVALID` (255) is refused by `IsValidConsoleColour` and
  drops the connection, then the 5 s reconnect makes the admin usable again.
- rcon: `getdate` / `getseed` / `clients` / `list_cmds` colours and text, `companies` producing only
  `RCON_END`, unknown command in `CC_ERROR`, and rcon output staying out of `ConsoleEvent`.
- console relay: AdminCmd's `GSLog.Info` arrives as `Console (script)`.
- gamescript: `countIndustry`, `createNews` broadcast and per company, `addGoal`/`removeGoal`/
  `removeAllGoal`, `setTownCargoGoal` with an unknown town, the unknown-command `exception` reply,
  and a reply whose id has no pending request.
- lifecycle: `shutdown()` before `startup()`, `startup()` twice, `isConnected()` false afterwards.

`assertConnectionUsable()` pings inside the awaited condition, because the sender is recreated on
every reconnect and a ping sent while the connection is down is dropped. Goal state persists across
admin connections (the game runs until the server restarts), so the goal test adds and removes its
own goal rather than assuming an empty slate.

## 7. Phase 3 - build and test wiring — done

Renamed the server-backed tests to `*IntegrationTest` (`OpenttdAdminTest` ->
`OpenttdAdminIntegrationTest`, `OpenttdClientTest` -> `OpenttdClientIntegrationTest`, class names
too). `build.gradle`: `test` filters `excludeTestsMatching('*IntegrationTest')`, and the
`integrationTest` task (`group = 'verification'`, `useJUnitPlatform()`, `dependsOn 'testClasses'`,
`testClassesDirs`/`classpath` from `sourceSets.test`) includes only those and excludes
`com.openttd.client.OpenttdClientIntegrationTest` (stale game-port enum crashes the shared server).
Do not reassign `sourceSets` on the task - Gradle 7.5 rejects it with "There's an extension
registered with name 'sourceSets'". `AGENTS.md` carries the ladder, the prerequisites and the
protocol facts.

## 8. Verification (all green)

1. `. /home/eric/www/tools/env.sh && gradle test` - 108 tests, 0 failures, ~1m22s, no server.
2. `mkdir -p openttd/game && gradle buildGameScript`, then
   `openttd -D -d 9,console=1,sprite=0,grf=0,sl=0 -c openttd/openttd.cfg` from the repo root, then
   `gradle integrationTest` - 25 tests, 0 failures, ~53s. Restart the server after editing
   `src/squirrel/*.nut`: the gamescript tar is read at game start, not per connection.
   Kill it with `pgrep -x openttd`, not `pkill -f 'openttd -D'` - the latter kills the shell that
   started it.
3. `gradle jar` - 153 KB, copied to `../strategyboard/lib/` and to `../openttd-admin/openttd-admin.jar`.
   No `robot/rule/*` or `ExternalServices` signature changed, and strategyboard compiles against the
   new jar: `. /home/eric/www/tools/env.sh && cd ../strategyboard && play check && play precompile`
   both exit 0 (never `play dependencies`, see `/home/eric/www/AGENTS.md`).
4. `../openttd-robot` does not compile - pre-existing API drift, see section 3.
