// types.go holds the names this package needs from other packages but cannot
// import here. The OpenttdAdmin facade is package admin (go/admin) and the
// admin protocol's listener and sender are package admin too
// (go/network/admin): this dialect has no import aliasing, so one file can only
// ever spell one of them as `admin`. admin_harness.go needs the facade, so the
// sender type it returns is named through the alias below instead.

package server

import "com.openttd/network/admin"

// NetworkAdminSender is network/admin's NetworkAdminSender: the thread sending
// messages to the server, which AdminHarness.getSend() hands out.
type NetworkAdminSender = admin.NetworkAdminSender
