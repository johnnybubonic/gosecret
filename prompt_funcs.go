package gosecret

import (
	`github.com/godbus/dbus`
)

// NewPrompt returns a pointer to a new Prompt based on a Dbus connection and a Dbus path.
func NewPrompt(conn *dbus.Conn, path dbus.ObjectPath) (prompt *Prompt) {

	prompt = &Prompt{
		Conn: conn,
		Dbus: conn.Object(DbusService, path),
	}

	return
}

// Path returns the path of the underlying Dbus connection.
func (p Prompt) Path() (path dbus.ObjectPath) {

	// Remove this method in V1. It's bloat since we now have an exported Dbus.
	path = p.Dbus.Path()

	return
}

// Prompt issues/waits for a prompt for unlocking a Locked Collection or Secret / Item.
func (p *Prompt) Prompt() (promptValue *dbus.Variant, err error) {

	var c chan *dbus.Signal
	var result *dbus.Signal

	// Prompts are asynchronous; we connect to the signal and block with a channel until we get a response.
	c = make(chan *dbus.Signal, 10)
	defer close(c)

	p.Conn.Signal(c)
	defer p.Conn.RemoveSignal(c)

	if err = p.Dbus.Call("org.freedesktop.Secret.Prompt.Prompt", 0, "").Store(); err != nil {
		return
	}

	for {
		if result = <-c; result.Path == p.Path() {
			*promptValue = result.Body[1].(dbus.Variant)
			return
		}
	}
}
