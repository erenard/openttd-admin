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
- Gradle compiles with an auto-provisioned Temurin JDK 11 (`java.toolchain` in `build.gradle`);
  first run downloads it into `~/.gradle/jdks`.

## Tests

`gradle test` needs a running OpenTTD server plus the gamescript. Two prerequisites, both verified:

1. `mkdir -p openttd/game && gradle buildGameScript` → installs `openttd/game/AdminCmd-v1.tar`.
   Without it `testGameScript` fails (`expected: <2> but was: <0>`); OpenTTD logs
   "The GameScript by the name 'AdminCmd' was no longer found". The script's `cp -Ra lang`
   error is harmless (no `lang` dir).
2. Start a dedicated server from the repo root *before* the tests:
   `openttd -D -d 9,console=1,sprite=0,grf=0,sl=0 -c openttd/openttd.cfg`
   (the repo root is OpenTTD's search path, so `openttd/` must exist).
   `OpenttdServerWrapper` only spawns its own server when port 3977 is free and never waits for
   the admin listener, so the first `OpenttdAdmin.startup()` hits "Connection refused" and
   `getSend()` is null → `assertNotNull(send)` fails. A pre-started server makes the suite green
   and also satisfies `OpenttdClientTest`, which does not extend the wrapper and needs port 3979
   (otherwise `shutdown()` NPEs on a sender that was never created).

Single test: `gradle test --tests 'com.openttd.admin.OpenttdAdminTest.testPingPong'`.

Tests read `./openttd/secrets.cfg` (`admin_password`) relative to the repo root via `TestUtils`.

## Layout

- `network/constant/*` mirror upstream enums (`TcpAdmin`, `TcpGame`, `PacketType`); when packets
  change, cross-check `/home/eric/www/OpenTTD/src/network/network_admin.cpp` / `network_client.cpp`
  — the code comments point there ("See network_admin.cpp for updates"). That checkout is upstream
  master (16.0-dev), not the binary in use, so read the matching tag: `git show 15.3:src/network/…`.
- `network/admin/NetworkAdmin` + `NetworkAdminSender` are listener/sender threads that reconnect
  every 5 s. `getSend()` returns null until a connection succeeds, and `shutdown()` assumes it is
  non-null.
- `admin/OpenttdAdmin` is the facade: `EventDispatcher` + `admin/model/Game` translate protocol
  packets into functional events (`CompanyEvent`, `ClientEvent`, …). `gamescript/GSExecutor`
  does JSON RPC into the AdminCmd gamescript, correlating replies by request id.
- `robot/rule/*` and `robot/ExternalServices` are what the embedding apps use. `ExternalServices`
  is a singleton whose `ExternalGameService`/`ExternalUserService` are injected by strategyboard;
  strategyboard calls `TimerObjective(OpenttdAdmin, ExternalUsers, long, int)` and
  `ExposeGameInfo(OpenttdAdmin, long)` — keep those signatures.
- `src/squirrel/*.nut` is a copy of the AdminCmd gamescript; the canonical source is
  `../AdminCmd/src` (its `make_tar.py` installs into `~/.openttd/game`, this repo's into
  `./openttd/game`). Edit both, or edit upstream and re-copy.
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
