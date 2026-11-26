// GameProcessOptions is com.openttd.robot.launcher.GameProcessOptions: the
// command line and working directory GameProcess starts the dedicated server with.
//
// Java keeps the directory as a java.io.File; the port keeps the path as a string,
// which is what os.StartProcess wants, and is what the caller passed in the first
// place. There is no Java test for this class, so game_process_test.go is new
// coverage written against this source as the specification.

package launcher

// GameProcessOptions is the two fields of the Java class: the directory to start
// in and the command line to start with.
type GameProcessOptions struct {
	openttdPath string
	options     []string
}

// NewGameProcessOptions is GameProcessOptions(String openttdPath, String
// openttdCommand): the command plus the two switches the launcher always adds,
// -D for "daemonize" and -d 1 for the debug level that makes the server print
// the "dbg: " lines GameProcess routes to the debug log.
func NewGameProcessOptions(openttdPath string, openttdCommand string) *GameProcessOptions {
	return &GameProcessOptions{
		openttdPath: openttdPath,
		options:     []string{openttdCommand, "-D", "-d 1"}}
}

// AddOption is addOption(String option): one more switch on the command line,
// which is how the caller adds the scenario, the password and the console port.
func (o *GameProcessOptions) AddOption(option string) {
	o.options = append(o.options, option)
}

// GetOpenttdPath is getOpenttdPath(): the working directory of the process, as a
// path instead of Java's File.
func (o *GameProcessOptions) GetOpenttdPath() string { return o.openttdPath }

// GetOptions is getOptions(): the command line, argv[0] first, which is what
// ProcessBuilder and os.StartProcess both want.
func (o *GameProcessOptions) GetOptions() []string { return o.options }
