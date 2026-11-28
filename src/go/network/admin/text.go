// Shared formatting for the network/admin model classes.

package admin

// boolText is Java's string concatenation of a boolean: ToString prints
// "true" or "false".
func boolText(value bool) string {
	if value {
		return "true"
	}
	return "false"
}
