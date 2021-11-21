package libsecret

import (
	"github.com/godbus/dbus"
)

// NewCollection returns a pointer to a new Collection based on a dbus connection and a dbus path for a specific Session.
func NewCollection(conn *dbus.Conn, path dbus.ObjectPath) (c *Collection, err error) {

	var dbusNames []string

	if conn == nil {
		err = ErrNoDbusConn
		return
	} else if path == "" {
		err = ErrBadDbusPath
		return
	}

	// dbus.Conn.Names() will (should) ALWAYS return a slice at LEAST one element in length.
	if dbusNames = conn.Names(); dbusNames == nil || len(dbusNames) < 1 {
		err = ErrNoDbusConn
		return
	}

	c = &Collection{
		Conn: conn,
		Dbus: conn.Object(DBusServiceName, path),
	}

	return
}

// Items returns a slice of Item items in the Collection.
func (collection *Collection) Items() (items []Item, err error) {

	var variant dbus.Variant
	var paths []dbus.ObjectPath

	if variant, err = collection.Dbus.GetProperty("org.freedesktop.Secret.Collection.Items"); err != nil {
		return
	}

	paths = variant.Value().([]dbus.ObjectPath)

	items = make([]Item, len(paths))

	for idx, path := range paths {
		items[idx] = *NewItem(collection.Conn, path)
	}

	return
}

// Delete removes a Collection from a Session.
func (collection *Collection) Delete() (err error) {

	var promptPath dbus.ObjectPath
	var prompt *Prompt

	if err = collection.Dbus.Call("org.freedesktop.Secret.Collection.Delete", 0).Store(&prompt); err != nil {
		return
	}

	if isPrompt(promptPath) {
		prompt = NewPrompt(collection.Conn, promptPath)

		if _, err := prompt.Prompt(); err != nil {
			return
		}
	}

	return
}

// SearchItems (IN Dict<String,String> attributes, OUT Array<ObjectPath> results);
func (collection *Collection) SearchItems(profile string) ([]Item, error) {
	attributes := make(map[string]string)
	attributes["profile"] = profile

	var paths []dbus.ObjectPath

	err := collection.Dbus.Call("org.freedesktop.Secret.Collection.SearchItems", 0, attributes).Store(&paths)
	if err != nil {
		return []Item{}, err
	}

	items := []Item{}
	for _, path := range paths {
		items = append(items, *NewItem(collection.Conn, path))
	}

	return items, nil
}

// CreateItem (IN Dict<String,Variant> properties, IN Secret secret, IN Boolean replace, OUT ObjectPath item, OUT ObjectPath prompt);
func (collection *Collection) CreateItem(label string, secret *Secret, replace bool) (*Item, error) {
	properties := make(map[string]dbus.Variant)
	attributes := make(map[string]string)

	attributes["profile"] = label
	properties["org.freedesktop.Secret.Item.Label"] = dbus.MakeVariant(label)
	properties["org.freedesktop.Secret.Item.Attributes"] = dbus.MakeVariant(attributes)

	var path dbus.ObjectPath
	var prompt dbus.ObjectPath

	err := collection.Dbus.Call("org.freedesktop.Secret.Collection.CreateItem", 0, properties, secret, replace).Store(&path, &prompt)
	if err != nil {
		return &Item{}, err
	}

	if isPrompt(prompt) {
		prompt := NewPrompt(collection.Conn, prompt)

		result, err := prompt.Prompt()
		if err != nil {
			return &Item{}, err
		}

		path = result.Value().(dbus.ObjectPath)
	}

	return NewItem(collection.Conn, path), nil
}

// READ Boolean Locked;
func (collection *Collection) Locked() (bool, error) {
	val, err := collection.Dbus.GetProperty("org.freedesktop.Secret.Collection.Locked")
	if err != nil {
		return true, err
	}

	return val.Value().(bool), nil
}
