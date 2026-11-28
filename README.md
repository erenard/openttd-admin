Build: `gradle jar` → `build/libs/openttd-admin.jar` (no dependencies inside; see
`../openttd-robot/lib` for the runtime jars). Source `. /home/eric/www/tools/env.sh`
first (Temurin JDK 17 + Gradle 7.5).

Usage, against a running dedicated server (admin port 3977):

```
java -cp ".:build/libs/openttd-admin.jar:/home/eric/www/openttd-robot/lib/gson-2.2.2.jar:/home/eric/www/openttd-robot/lib/slf4j-api-1.6.6.jar:/home/eric/www/openttd-robot/lib/slf4j-log4j12-1.6.6.jar:/home/eric/www/openttd-robot/lib/log4j-1.2.17.jar" \
  com.openttd.demo.OpenttdAdminConsole <admin_password>
```

The leading `.` is required: log4j 1.2 reads `log4j.xml` from the classpath, which is the repo root.

## Running against a real OpenTTD server

The demos and the integration tests need a dedicated OpenTTD server (locally `/usr/bin/openttd`,
currently 15.3) listening on the admin port 3977:

1. Gamescript prerequisite (for the gamescript RPC tests):
   `mkdir -p openttd/game && gradle buildGameScript` installs `openttd/game/AdminCmd-v1.tar`.
   Re-run it after editing `src/squirrel/*.nut` and restart the server: the tar is read at game
   start, not per admin connection.
2. Start the server **from the repo root** (the root is OpenTTD's search path, so `openttd/` must
   resolve), in a background shell:
   ```
   openttd -D -d 9,console=1,sprite=0,grf=0,sl=0 -c openttd/openttd.cfg
   ```
   Allow a few seconds for map generation before connecting.
3. Verify the listener: `ss -ltn | grep 3977`.
4. Config requirements: `openttd/openttd.cfg` must keep `allow_insecure_admin_login = true` and
   `server_admin_chat = true`; the password lives in `openttd/secrets.cfg` (`admin_password`),
   which the tests read relative to the repo root.
5. Stop it with `kill $(pgrep -x openttd)`. Never `pkill -f 'openttd -D'` — that pattern also
   matches the shell that started the server.

## Java tests

- `gradle test` — offline protocol tests, no OpenTTD process.
- `gradle integrationTest` — against a real dedicated server (prerequisites above).
  `OpenttdServerWrapper` spawns its own server only when 3977 is free; pre-starting stays reliable.

## Go port (`src/go`)

Module `com.openttd`, `go 1.27`, stdlib only — nothing to fetch or vendor. Run tasks from that
directory:

```
cd src/go
go build com.openttd/...              # compile every package
go vet com.openttd/...                # static checks
go test com.openttd/...               # both ladders in one run
go clean -cache                       # drop the build cache (go clean -testcache for test results)
```

`go test com.openttd/...` runs the offline packages (mirror of `gradle test`) plus the two
integration files in `com.openttd/admin` (mirror of `gradle integrationTest`). The Go ladder
never spawns a server: integration tests skip when nothing listens on 3977, so start one first
(see above) and check the output for skips rather than passes. With a server up the full ladder
passes; `com.openttd/admin` takes ~54s, mostly integration time.

- Single test: `go test com.openttd/admin -run TestJoinAndWelcome`
- Re-run against the server without cached results: `go test -count=1 com.openttd/admin`

The Go demos (`src/go/demo`) are the same samples as the Java `demo/*` classes but without
`main()`s; they are exercised offline by `demo_test.go`, not runnable against a server directly.

Package layout, protocol facts and version-bump rules: see `AGENTS.md`.
