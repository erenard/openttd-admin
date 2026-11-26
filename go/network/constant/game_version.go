// GameVersion mirrors com.openttd.network.constant.GameVersion.
//
// Hardcoded to the OpenTTD binary in use. The game port rejects a client whose
// revision or newgrf version differs (NETWORK_ERROR_WRONG_REVISION), so a stale
// value makes NetworkClient joins silently fail; the admin port only logs the
// mismatch. Bump these together with openttd/openttd.cfg and openttd/secrets.cfg.

package constant

const MAJOR_VERSION = 15
const MINOR_VERSION = 3

// OPENTTD is the advertised version string (see .version).
const OPENTTD = "15.3"

// NEWGRF is the newgrf version checksum (see rev.cpp.in).
const NEWGRF = (MAJOR_VERSION+16)<<24 | MINOR_VERSION<<20 | 1<<19 | 28004
