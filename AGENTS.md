# openttd-admin

Java library that speaks OpenTTD's TCP protocols: the admin port (3977) and the game
port (3979). It is not an app — it is consumed as a jar by `../openttd-robot` and
`../strategyboard`.

## Build

- Source the toolchain first: `. /home/eric/www/tools/env.sh` (Temurin JDK 17 + Gradle 7.5).
  Unsourced, Gradle runs on Fedora's OpenJDK 25 and warns about native-access.
- `gradle jar` → `build/libs/openttd-admin.jar`. The jar bundles no dependencies;
  consumers supply gson/slf4j/log4j at runtime (`../openttd-robot/lib`, `../strategyboard/lib`).
- Gradle is the only build. The Ant `build.xml` was removed: its `dist` never worked (`lib/` does
  not exist — the dep jars live in `../openttd-robot/lib`) and its `clean` deleted the same
  `build/` directory Gradle uses, after which Gradle reported tasks `UP-TO-DATE` with no outputs
  and `--tests` failed with "No tests found" (`gradle clean <task>` is the recovery). Don't
  reintroduce Ant here.
- Publishing: copy `build/libs/openttd-admin.jar` to `../strategyboard/lib/`
  (see `/home/eric/www/PLAN.md` step 2). Never add `openttd-robot.jar` there — it shadows
  `com/openttd/robot/rule/*` constructor signatures strategyboard calls.
  `../openttd-robot/build.xml` compiles against `../openttd-admin/openttd-admin.jar`, a root-level
  path Gradle does not produce, so copy the jar there (`openttd-admin.jar` is gitignored) when
  building that project.
  Both copies are current as of `9eb8d52` (md5 `f36ab3dd…`, 153 KB): strategyboard compiles against
  them (`play check` and `play precompile` exit 0). `../openttd-robot` does *not* compile — its rules
  call `Company.getPerformance()/getValue()/getInauguration()`, treat money as `int` and pass extra
  arguments to `super` (`openttd-admin`'s is `AbstractRule(OpenttdAdmin)`). That is API drift predating
  the protocol work: no public method was removed, and robot could not compile before either because
  the root jar did not exist. Do not "fix" it by restoring old signatures; porting `robot/rule/*` is a
  separate task.
- Gradle compiles with an auto-provisioned Temurin JDK 11 (`java.toolchain` in `build.gradle`);
  first run downloads it into `~/.gradle/jdks`.

## Tests

Two ladders. Neither runs the other: `test` excludes `*IntegrationTest`,
`integrationTest` includes only those (see `build.gradle`).

- `gradle test` — 108 offline protocol tests, no OpenTTD process.
  `test/java/com/openttd/server/FakeAdminServer.java` frames packets the way
  `src/network/core/packet.cpp` does and `AdminHarness` wires it to an `OpenttdAdmin`;
  `PacketWriter`/`UpstreamAdminProtocol` build server->admin packets from the 15.3
  `Packet::Send_*` signatures, `UpstreamCalendar` mirrors `CalendarConvertDateToYMD`.
  Covers every `PacketServerType` id, the sender's byte-for-byte commands, the event
  mapping in `Game`, `Protocol`, `Convert`, `GSExecutor`. `ConvertTest` pins the date mapping
  for game years 1-2050 against `CalendarConvertDateToYMD`; game year 0 maps to Java year 1
  (documented, not fixed — upstream has no year-0 game date). Runs ~1m22s.
- `gradle integrationTest` — 25 tests against a real dedicated server: 22 in
  `OpenttdAdminProtocolIntegrationTest` plus the three pre-existing ones
  (`OpenttdAdminIntegrationTest`). Two prerequisites, both verified:
  1. `mkdir -p openttd/game && gradle buildGameScript` → installs `openttd/game/AdminCmd-v1.tar`.
     Without it the gamescript RPC tests fail; OpenTTD logs
     "The GameScript by the name 'AdminCmd' was no longer found". The script's `cp -Ra lang`
     error is harmless (no `lang` dir). Re-run it after editing `src/squirrel/*.nut`, and restart
     the server: the tar is read at game start, not per admin connection.
  2. `openttd -D -d 9,console=1,sprite=0,grf=0,sl=0 -c openttd/openttd.cfg` from the repo root
     *before* the tests (the repo root is OpenTTD's search path, so `openttd/` must exist).
     `OpenttdServerWrapper` spawns its own server only when 3977 is free, and
     `awaitAdminPort()` then waits for the listener; the server still needs a few seconds to
     generate its map, so pre-starting stays the reliable option. Kill it with `pgrep -x openttd`
     and `kill`, not `pkill -f 'openttd -D'` — that pattern matches the shell that started it.

- The Go port has the same two ladders in `src/go` (module `com.openttd`, `go 1.27`, stdlib only).
  Run it from that directory: `. /home/eric/www/tools/env.sh && cd src/go && go vet com.openttd/... &&
  go test com.openttd/...`. The offline packages mirror `gradle test`; `src/go/admin/integration_test.go`
  and `src/go/admin/protocol_integration_test.go` mirror `gradle integrationTest` and skip when nothing
  listens on 3977 (`server.CheckAdminPort()`), so the two prerequisites above apply unchanged. The Go
  ladder never spawns a server: `go test` has no @BeforeAll/@AfterAll pair that could own the process
  for the whole run. `openttd.Await` panics on a timeout, which would abort the binary, so integration
  tests wait through `awaitTrue` (`t.Errorf` + `t.FailNow()`) and callers `return` on a false result.

`com.openttd.client.OpenttdClientIntegrationTest` is excluded from `integrationTest`: its
`TcpGame.PacketGameType` is stale for 15.3, so the join is refused with
`NETWORK_ERROR_WRONG_REVISION` and `Receive_CLIENT_AUTH_RESPONSE` hits `NOT_REACHED` — the
dedicated server crashes (`openttd/crash*.json.log`). Fix the enum before putting it back.

Single test: `gradle test --tests 'com.openttd.network.admin.NetworkAdminServerPacketsTest'`,
`gradle integrationTest --tests 'com.openttd.admin.OpenttdAdminProtocolIntegrationTest'`.

Tests read `./openttd/secrets.cfg` (`admin_password`) relative to the repo root via `TestUtils`.
`openttd/openttd.cfg` must keep `allow_insecure_admin_login = true` and
`server_admin_chat = true`, or the join/chat tests fail.

Protocol facts the tests pin (read the 15.3 tag, not the master checkout):
`CLIENT_ID_SERVER = 1` (`network_type.h`), so a dedicated server has exactly client #1 and
id 0 is `INVALID_CLIENT_ID`; an admin's own chat is never echoed back (`NetworkAdminChat`
returns on `from_admin`), so chat assertions look at the console relay instead; only
DATE/CLIENT_INFO/COMPANY_INFO/COMPANY_ECONOMY/COMPANY_STATS/CMD_NAMES are pollable, and
`NetworkAdminSender.updateFrequency` falls back to an advertised frequency instead of letting
the server close the connection. Console lines carry the debug category as their origin, so
gamescript output arrives as `Console (script)`, and Squirrel prints an array argument as
`(array : 0x…)`, not `[]`. `Game.updateAdmin` swaps the network model per packet, so a captured
`GameInfo`/`Client` reference goes stale — re-read it inside the awaited condition. Goal state
persists across admin connections (the game runs until the server restarts).

## Layout

- `src/java` and `src/go` are the same library twice: `src/go` is the port, package for package
  (`src/go/network/…`, `src/go/admin/…`, `src/go/robot/…`, `src/go/server/…`), with the Java tests as
  the spec. Java stays buildable — `../strategyboard` and `../openttd-robot` consume the jar. The nine
  Java `*Listener` interfaces have no Go counterpart; they are folded into `src/go/admin/event/*_event.go`.
- `network/constant/*` mirror upstream enums (`TcpAdmin`, `TcpGame`, `PacketType`); when packets
  change, cross-check `/home/eric/www/OpenTTD/src/network/network_admin.cpp` / `network_client.cpp`
  — the code comments point there ("See network_admin.cpp for updates"). That checkout is upstream
  master (16.0-dev), not the binary in use, so read the matching tag: `git show 15.3:src/network/…`.
- `network/admin/NetworkAdmin` + `NetworkAdminSender` are listener/sender threads that reconnect
  every 5 s. `getSend()` returns null until a connection succeeds; `shutdown()` tolerates a null
  sender, closes the socket and interrupts the reconnect sleep, so it returns in milliseconds
  instead of waiting for `socket.receive()` / `Thread.sleep(5000)` to unblock. PROTOCOL arrives
  before WELCOME and WELCOME replaces the model's `GameInfo`, so the admin protocol version is
  carried over into it (`GameInfo.getVersion()`).
- `admin/OpenttdAdmin` is the facade: `EventDispatcher` + `admin/model/Game` translate protocol
  packets into functional events (`CompanyEvent`, `ClientEvent`, …). `gamescript/GSExecutor`
  does JSON RPC into the AdminCmd gamescript, correlating replies by request id.
- `robot/rule/*` and `robot/ExternalServices` are what the embedding apps use. `ExternalServices`
  is a singleton whose `ExternalGameService`/`ExternalUserService` are injected by strategyboard;
  strategyboard calls `TimerObjective(OpenttdAdmin, ExternalUsers, long, int)` and
  `ExposeGameInfo(OpenttdAdmin, long)` — keep those signatures.
- `src/squirrel/*.nut` is a copy of the AdminCmd gamescript; the canonical source is
  `../AdminCmd/src` (its `make_tar.py` installs into `~/.openttd/game`, this repo's into
  `./openttd/game`). Edit both, or edit upstream and re-copy. Squirrel has no bitwise operator
  between two bools, so `removeAllGoal` accumulates with `||`; an admin RPC without `"i"` is
  answered `Malformed json.`, which is why the goal and town helpers send through `GSExecutor`
  (`GSJsonRequest`).
- `demo/*` hold the runnable `main()`s (`OpenttdAdminConsole`, `AdminPing`,
  `AdministrationDemo`, `HelloWorldDemo`). Logging: log4j 1.2 needs `log4j.xml` on the classpath —
  tests get `test/resources/log4j.xml` automatically, and `src/java/log4j.xml` never lands on it
  (Gradle resources come from `src/squirrel`, not `src/java`). For demos, put the repo root first
  on `-cp`; the verified invocation is in `README`.

## OpenTTD version bumps

`network/constant/GameVersion` hardcodes `MAJOR_VERSION`/`MINOR_VERSION` (now 15.3, matching
`/usr/bin/openttd`) and derives the advertised version string plus the newgrf checksum
(`(major + 16) << 24 | minor << 20 | 1 << 19 | 28004`, i.e. `0x1F386D64`). The game port rejects a
client whose revision or newgrf version differs (`NETWORK_ERROR_WRONG_REVISION`), so a stale value
makes `NetworkClient` joins silently fail; the admin port only logs the mismatch. Bump `GameVersion`
together with `openttd/openttd.cfg` and `openttd/secrets.cfg` — OpenTTD rewrites their `[version]`
sections on start, and repo history is one commit per OpenTTD release.

## Local state

`openttd/` is OpenTTD's config/personal dir. Only `openttd.cfg` and `secrets.cfg` are tracked
(`.gitignore` has `openttd/*`, so everything else the server creates — `baseset/`, `save/`,
`game/`, `content_download/` — is untracked runtime state). Do not commit anything else in it.

Workspace-wide setup (distrobox container, MariaDB, Play/JDK provisioning) lives in
`/home/eric/www/AGENTS.md`.
