# Changelog — Go port (`src/go`)

The library version follows the OpenTTD release this protocol targets:
`GameVersion.OPENTTD` = `libraryVersion` in `build.gradle` = **15.3**.

## 15.3 — 2026-11-28 — branch `go-port`

The first full Go port of openttd-admin: module `com.openttd`, `go 1.27`, stdlib only,
`src/go` mirroring `src/java` package for package, with the Java tests as the spec.
`src/java` and `test/java` were never touched: the jar and the two apps that consume it
(`../strategyboard`, `../openttd-robot`) keep building.

### Added, stage by stage (oldest commit first)

| Commit | Stage |
|---|---|
| `8887f5c` | Stage 1: `network/constant` (TcpAdmin/NetworkType enums as typed consts with `*Of()` lookups reproducing Java's `valueOf` fallbacks), `Packet`/`Socket` codec with `[len16][type8]` framing, the slf4j shim in `util/log`, `PacketWriter` harness. `TcpAdminWireFormatTest` and `PacketCodecTest` green. |
| `3f1f497` | Stage 2: the offline harness — `FakeAdminServer` (accept loop, framing, automatic PING→PONG, handshake), `UpstreamAdminProtocol` (`GAME_ADMIN_VERSION`, the 15.3 frequency mask table, PROTOCOL/WELCOME packets), `UpstreamCalendar` (`CalendarConvertDateToYMD`), `test_utils.go` (`Wait`/`Await`/`AwaitSize`). |
| `8d6192c` | `network/admin` data model (`NetworkException`, `Client`, `GameInfo`, `Company`, the events, `NetworkModel`) and `util.Convert`. Java's `BigInteger` money becomes `uint64`; `retreiveCompanies`/`retreiveClients` sort by id because Go maps iterate randomly; enum name tables filled in. |
| `463c8c3` | `gamescript`: `GSExecutor` JSON RPC correlated by request id; hand-rolled JSON because gson keeps insertion order (Go maps sort keys) and keeps original number text, so `50000000000` does not arrive as `5e10`; `GSRequest` as an interface; `GSJsonRequest`/`GSNewsPaper`; `game_script.go` and `ottd.go` constants. |
| `356977d` | Sender layer: `core.Configuration`/`NetworkSender`, a hand-rolled `PacketQueue` (this toolchain's `sync` has no condition variable), `NetworkAdminSender` with every `ADMIN_PACKET_ADMIN_*` command byte for byte, the frequency fallbacks and the goal commands. All 28 `NetworkAdminSenderTest` cases ported. |
| `7b5e03b` | `network/admin.NetworkAdmin`: listener thread, 5 s reconnect loop, decoder whose `handlePacket` returns a `*NetworkException` instead of throwing; the seven session-ending packets reproduce Java's log strings byte for byte. 15 decode cases; `FakeAdminServer.SendRaw`. |
| `2589c69` | `admin` facade: `admin/event` (ten event classes with `Kind()`/`Notify()` standing in for Java's Class-keyed listener map), `model.Game` packet→event mapping, `OpenttdAdmin` wiring, `server.AdminHarness`; sender tests move to packet-count deltas. |
| `13f71f7` | Client layer: `NetworkClient` and its sender, `tcp_game.go` with the 15.3 ids (the Java enum is stale; `PACKET_SERVER_ERROR` and `ADMIN_PACKET_ADMIN_POLL` share id 3, so protocols compare by `SubProtocol`), UDP bound-socket API (`NewUdpListener`/`ReceiveFrom`/`SendTo`), fixed `hashPassword`. 15 new offline cases. |
| `77c370d` | Robot layer: `robot/model`, `ExternalServices`, the ten rules, the dedicated-server launcher. `AbstractRule` becomes a composed base. New tests for rules, objectives and the process (no Java tests existed). |
| `4822dd2` | `go fmt` over the ported tree (formatting only). |
| `2d7001f` | Demo layer: `CLIUtil`, `TestUtil`, `AdminPing`, `HelloWorldDemo`, `OpenttdAdminConsole`, `AdministrationDemo`, plus `DayOfWeek`. `demo_test.go`: 10 new tests (Java has none). |
| `16194f1` | Integration ladder: the tree moves from `go/` to `src/go/`; `server/dedicated_server.go` (`CheckAdminPort`, `CanConnectAdminPort`, `AwaitAdminPort`, `Destroy`, `ReadConsole`); `integration_test.go` + `protocol_integration_test.go` (3 facade + 22 protocol tests), which **skip when nothing listens on 3977** — `go test` has no `@BeforeAll`/`@AfterAll` pair that could own the server process. |
| `8bdb70e` | `go fmt` sweep for the integration stage's new files. |
| `5c3eddc` | `README` → `README.md`: server prerequisites, both Java ladders, the Go tasks. |
| (this commit) | Packaging: `goVet` + `goPackage` in `build.gradle` (see below). |

### Bugs the port had to fix while mirroring Java

- `splitProperty` consumed the `=`/`:` separator instead of keeping it, so
  `admin_password = test-password` yielded `= test-password` and every join failed with
  `NETWORK_ERROR_WRONG_PASSWORD` (`16194f1`).
- The transcription of `Game.updateAdmin`'s fall-through groups swallowed `SHUTDOWN` into
  `NEWGAME` (Go cases do not fall through, and an empty `case A:` dispatches nothing);
  caught by the event-mapping table test (`2589c69`).
- `util/convert.go` had a day-offset fudge and a March-based civil month, leaving
  `AddYears`/`AddMonths` off by one day and `WeekOfYear`/`Trimester` unusable (`77c370d`).
- `hashPassword` is a deliberate deviation: Java's version cannot return at all (an
  unflushed `PrintWriter` makes the next write `write(chars, -1, 1)`, which killed the
  listener thread). The port implements the upstream intent (`13f71f7`).

### Deliberate deviations (dialect-forced, pinned by tests or comments)

- No `main()`s: every Java demo entry point is a `Main` function of package `demo`.
  `CLIUtil.parseArguments` reports usability instead of `System.exit`; `TestUtil` logs
  instead of printing and flags `endGame` instead of exiting; `AdministrationDemo.Main`
  keeps Java's copy-paste slip (it builds a `HelloWorldDemo`).
- `AdminPing` gains `Shutdown()`; its `Ping()` logs instead of NPE-ing when no sender exists;
  sender/listener paths that Java would fail on with a `NullPointerException` log and do nothing.
- Volatile fields become lock-guarded; `Thread.sleep(5000)` is polled against `IsRunning()`
  so `Shutdown()` never interrupts; the dispatcher's missing right-kind listener logs
  instead of throwing `ClassCastException`; `System.err` goes through the logger.

### Packaging tasks (in `build.gradle`)

- `gradle goVet` — `go vet` over the module (also compiles it).
- `gradle goPackage` — depends on `goVet`, packs `src/go` into
  `build/libs/openttd-admin-go-15.3.tgz`. A Go library ships as source, so the whole
  module (go.mod, packages, tests, this changelog) is included.

### Verification

- `go vet com.openttd/...` clean; `go test com.openttd/...` green — offline packages and,
  with a dedicated 15.3 server on 3977, the integration ladder too (skips otherwise).
- `gradle test` (108) and `gradle integrationTest` (25) still pass throughout the branch;
  `gradle jar` still produces `build/libs/openttd-admin.jar` unchanged.
