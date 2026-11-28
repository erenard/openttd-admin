// CLIUtil is com.openttd.demo.CLIUtil: the "dummy class that makes the samples
// easier to read" every demo of this library uses to turn its two command line
// arguments into a Configuration - the server to connect to and the admin
// password from openttd.cfg.
//
// Java writes the usage to System.err and then calls System.exit(0); the port
// writes the same two lines through the logger - its stand-in for System.err -
// and returns false, so the caller stops instead of the process dying inside a
// helper. There is no Java test for this class, so demo_test.go is new coverage
// written against this source.

package demo

import "strings"
import "com.openttd/network/core"
import "com.openttd/robot/rule"
import "com.openttd/util"

var cliUtilLog = util.GetLogger("CLIUtil")

// ReportUsage is the two System.err.println() calls of the catch block. The port
// logs them at error, which is its stand-in for System.err throughout: same
// stream by default, and a test can point it at a file.
func ReportUsage() {
	cliUtilLog.Error("Usage: java -jar openttd-robot.jar localhost:3977 admin_password")
	cliUtilLog.Error("See openttd.cfg to set your server's admin_password first !")
}

// ParseArguments is parseArguments(String[] args, Configuration configuration):
// host and admin port come from the first argument, split on the colon, and the
// password from the second.
//
// DEVIATION FROM JAVA: args[0] and url.split(":")[1] are read unguarded, so a
// command line that is too short or has no colon ends in
// ArrayIndexOutOfBoundsException, and only a bad number is caught. The port
// reports the usage for all three, which is what the catch block does for the
// one Java handles.
func ParseArguments(args []string, configuration *core.Configuration) bool {
	if len(args) < 2 {
		ReportUsage()
		return false
	}
	var parts = strings.Split(args[0], ":")
	if len(parts) < 2 {
		ReportUsage()
		return false
	}
	// rule.ParseInt is the port's Integer.parseInt: false where Java throws
	// NumberFormatException, including for a number outside the int range.
	var port, parsed = rule.ParseInt(parts[1])
	if !parsed {
		ReportUsage()
		return false
	}
	configuration.Host = parts[0]
	configuration.AdminPort = port
	configuration.Password = args[1]
	return true
}
