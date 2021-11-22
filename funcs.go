package libsecret

import (
	`strings`

	`github.com/godbus/dbus`
)

// isPrompt returns a boolean that is true if path is/requires a prompt(ed path) and false if it is/does not.
func isPrompt(path dbus.ObjectPath) (prompt bool) {

	prompt = strings.HasPrefix(string(path), PromptPrefix)

	return

}
