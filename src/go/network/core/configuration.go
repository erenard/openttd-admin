// Configuration mirrors com.openttd.network.core.Configuration: the five fields
// an embedding app sets before calling startup(). The defaults are OpenTTD's
// dedicated server's admin port (3977), game port (3979) and the empty admin
// password, i.e. what openttd.cfg has before an admin password is configured.
//
// The fields are public in Java, so an embedding app assigns them directly,
// which is what the tests do with the fake server's port; the Go port keeps them
// exported for the same reason.
package core

type Configuration struct {
	Host       string
	AdminPort  int
	ClientPort int
	Password   string
	Name       string
}

// NewConfiguration is the constructor: the defaults Java's field initializers
// give.
func NewConfiguration() *Configuration {
	return &Configuration{Host: "localhost", AdminPort: 3977, ClientPort: 3979, Password: "", Name: "openttd"}
}
