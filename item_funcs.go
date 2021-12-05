package gosecret

import (
	`github.com/godbus/dbus`
)

// NewItem returns a pointer to an Item based on Collection and a Dbus path.
func NewItem(collection *Collection, path dbus.ObjectPath) (item *Item, err error) {

	if collection == nil {
		err = ErrNoDbusConn
	}

	if _, err = validConnPath(collection.Conn, path); err != nil {
		return
	}

	item = &Item{
		&DbusObject{
			Conn: collection.Conn,
			Dbus: collection.Conn.Object(DbusService, path),
		},
	}

	return
}

// Label returns the label ("name") of an Item.
func (i *Item) Label() (label string, err error) {

	var variant dbus.Variant

	if variant, err = i.Dbus.GetProperty("org.freedesktop.Secret.Item.Label"); err != nil {
		return
	}

	label = variant.Value().(string)

	return
}

// Locked indicates that an Item is locked (true) or unlocked (false).
func (i *Item) Locked() (isLocked bool, err error) {

	var variant dbus.Variant

	if variant, err = i.Dbus.GetProperty("org.freedesktop.Secret.Item.Locked"); err != nil {
		isLocked = true
		return
	}

	isLocked = variant.Value().(bool)

	return
}

// GetSecret returns the Secret in an Item using a Session.
func (i *Item) GetSecret(session *Session) (secret *Secret, err error) {

	secret = new(Secret)

	if err = i.Dbus.Call(
		"org.freedesktop.Secret.Item.GetSecret", 0, session.Path(),
	).Store(&secret); err != nil {
		return
	}

	return
}

// Delete removes an Item from a Collection.
func (i *Item) Delete() (err error) {

	var prompt *Prompt
	var promptPath dbus.ObjectPath

	if err = i.Dbus.Call("org.freedesktop.Secret.Item.Delete", 0).Store(&promptPath); err != nil {
		return
	}

	if isPrompt(promptPath) {
		prompt = NewPrompt(i.Conn, promptPath)

		if _, err = prompt.Prompt(); err != nil {
			return
		}
	}

	return
}
