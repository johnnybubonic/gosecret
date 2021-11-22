package gosecret

import (
	`github.com/godbus/dbus`
)

// NewCollection returns a pointer to a new Collection based on a Dbus connection and a Dbus path.
func NewCollection(conn *dbus.Conn, path dbus.ObjectPath) (coll *Collection, err error) {

	// dbus.Conn.Names() will ALWAYS return a []0string with at least ONE element.
	if conn == nil || (conn.Names() == nil || len(conn.Names()) < 1) {
		err = ErrNoDbusConn
		return
	}

	if path == "" {
		err = ErrBadDbusPath
		return
	}

	coll = &Collection{
		Conn: conn,
		Dbus: conn.Object(DBusServiceName, path),
	}

	return
}

// Items returns a slice of Item pointers in the Collection.
func (c *Collection) Items() (items []*Item, err error) {

	var variant dbus.Variant
	var paths []dbus.ObjectPath

	if variant, err = c.Dbus.GetProperty(DbusItemsID); err != nil {
		return
	}

	paths = variant.Value().([]dbus.ObjectPath)

	items = make([]*Item, len(paths))

	for idx, path := range paths {
		items[idx] = NewItem(c.Conn, path)
	}

	return
}

// Delete removes a Collection.
func (c *Collection) Delete() (err error) {

	var promptPath dbus.ObjectPath
	var prompt *Prompt

	if err = c.Dbus.Call(DbusCollectionDelete, 0).Store(&promptPath); err != nil {
		return
	}

	if isPrompt(promptPath) {

		prompt = NewPrompt(c.Conn, promptPath)
		if _, err = prompt.Prompt(); err != nil {
			return
		}
	}

	return
}

// SearchItems searches a Collection for a matching profile string.
func (c *Collection) SearchItems(profile string) (items []Item, err error) {

	var paths []dbus.ObjectPath
	var attrs map[string]string = make(map[string]string, 0)

	attrs["profile"] = profile

	if err = c.Dbus.Call("org.freedesktop.Secret.Collection.SearchItems", 0, attrs).Store(&paths); err != nil {
		return
	}

	items = make([]Item, len(paths))

	for idx, path := range paths {
		items[idx] = *NewItem(c.Conn, path)
	}

	return
}

// CreateItem returns a pointer to an Item based on a label, a Secret, and whether any existing secret should be replaced or not.
func (c *Collection) CreateItem(label string, secret *Secret, replace bool) (item *Item, err error) {

	var prompt *Prompt
	var path dbus.ObjectPath
	var promptPath dbus.ObjectPath
	var variant *dbus.Variant
	var props map[string]dbus.Variant = make(map[string]dbus.Variant)
	var attrs map[string]string = make(map[string]string)

	attrs["profile"] = label
	props["org.freedesktop.Secret.Item.Label"] = dbus.MakeVariant(label)
	props["org.freedesktop.Secret.Item.Attributes"] = dbus.MakeVariant(attrs)

	if err = c.Dbus.Call(
		"org.freedesktop.Secret.Collection.CreateItem", 0, props, secret, replace,
	).Store(&path, &promptPath); err != nil {
		return
	}

	if isPrompt(promptPath) {
		prompt = NewPrompt(c.Conn, promptPath)

		if variant, err = prompt.Prompt(); err != nil {
			return
		}

		path = variant.Value().(dbus.ObjectPath)
	}

	item = NewItem(c.Conn, path)

	return
}

// Locked indicates that a Collection is locked (true) or unlocked (false).
func (c *Collection) Locked() (isLocked bool, err error) {

	var variant dbus.Variant

	if variant, err = c.Dbus.GetProperty("org.freedesktop.Secret.Collection.Locked"); err != nil {
		isLocked = true
		return
	}

	isLocked = variant.Value().(bool)

	return
}
