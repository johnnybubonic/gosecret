package gosecret

import (
	`strings`
	`time`

	`github.com/godbus/dbus`
)

/*
	CreateCollection creates a new Collection named `name` using the dbus.BusObject `secretServiceConn`.
	`secretServiceConn` should be the same as used for Collection.Dbus (and/or NewCollection).
	It will be called by NewCollection if the Collection does not exist in Dbus.

	Generally speaking, you should probably not use this function directly and instead use NewCollection.
*/
func CreateCollection(secretServiceConn *dbus.BusObject, name string) (c *Collection, err error) {

	var path dbus.ObjectPath

	if secretServiceConn == nil {
		err = ErrNoDbusConn
		return
	}

	path = dbus.ObjectPath(strings.Join([]string{DbusPath, ""}, "/"))

	// TODO.

	return
}

// NewCollection returns a pointer to a new Collection based on a Dbus connection and a Dbus path.
func NewCollection(conn *dbus.Conn, path dbus.ObjectPath) (coll *Collection, err error) {

	if _, err = validConnPath(conn, path); err != nil {
		return
	}

	coll = &Collection{
		Conn: conn,
		Dbus: conn.Object(DbusService, path),
		// lastModified: time.Now(),
	}

	_, _, err = coll.Modified()

	return
}

// Items returns a slice of Item pointers in the Collection.
func (c *Collection) Items() (items []*Item, err error) {

	var paths []dbus.ObjectPath

	if paths, err = pathsFromPath(c.Dbus, DbusCollectionItems); err != nil {
		return
	}

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

// Locked indicates if a Collection is locked (true) or unlocked (false).
func (c *Collection) Locked() (isLocked bool, err error) {

	var variant dbus.Variant

	if variant, err = c.Dbus.GetProperty("org.freedesktop.Secret.Collection.Locked"); err != nil {
		isLocked = true
		return
	}

	isLocked = variant.Value().(bool)

	return
}

// Label returns the Collection label (name).
func (c *Collection) Label() (label string, err error) {

	// TODO.

	return
}

// Created returns the time.Time of when a Collection was created.
func (c *Collection) Created() (created time.Time, err error) {

	// TODO.

	return
}

/*
	Modified returns the time.Time of when a Collection was last modified along with a boolean
	that indicates if the collection has changed since the last call of Collection.Modified.

	Note that when calling NewCollection, the internal library-tracked modification
	time (Collection.lastModified) will be set to the modification time of the Collection
	itself as reported by Dbus.
*/
func (c *Collection) Modified() (modified time.Time, isChanged bool, err error) {

	// TODO.

	if c.lastModified == time.Date(1, time.January, 1, 0, 0, 0, 0, time.UTC) {
		// It's "nil", so set it to modified.
		c.lastModified = modified
	}

	return
}
