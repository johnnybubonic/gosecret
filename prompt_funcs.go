package gosecret

import (
	`github.com/godbus/dbus/v5`
)

// NewPrompt returns a pointer to a new Prompt based on a Dbus connection and a Dbus path.
func NewPrompt(conn *dbus.Conn, path dbus.ObjectPath) (prompt *Prompt) {

	prompt = &Prompt{
		DbusObject: &DbusObject{
			Conn: conn,
			Dbus: conn.Object(DbusService, path),
		},
	}

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

	if err = p.Dbus.Call(
		DbusPrompterInterface, 0, "", // TODO: This last argument, the string, is for "window ID". I'm unclear what for.
	).Store(); err != nil {
		return
	}

	for {
		if result = <-c; result.Path == p.Dbus.Path() {
			*promptValue = result.Body[1].(dbus.Variant)
			return
		}
	}
}
