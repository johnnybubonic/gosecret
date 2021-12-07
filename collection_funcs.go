package gosecret

import (
	`strings`
	"time"

	`github.com/godbus/dbus/v5`
)

/*
	NewCollection returns a pointer to a Collection based on a Service and a Dbus path.
	You will almost always want to use Service.GetCollection instead.
*/
func NewCollection(service *Service, path dbus.ObjectPath) (coll *Collection, err error) {

	var splitPath []string

	if service == nil {
		err = ErrNoDbusConn
	}

	if _, err = validConnPath(service.Conn, path); err != nil {
		return
	}

	coll = &Collection{
		DbusObject: &DbusObject{
			Conn: service.Conn,
			Dbus: service.Conn.Object(DbusService, path),
		},
		service: service,
		// lastModified: time.Now(),
	}

	splitPath = strings.Split(string(coll.Dbus.Path()), "/")

	coll.name = splitPath[len(splitPath)-1]

	_, _, err = coll.Modified()

	return
}

// Items returns a slice of Item pointers in the Collection.
func (c *Collection) Items() (items []*Item, err error) {

	var paths []dbus.ObjectPath
	var item *Item
	var variant dbus.Variant
	var errs []error = make([]error, 0)

	if variant, err = c.Dbus.GetProperty(DbusCollectionItems); err != nil {
		return
	}

	paths = variant.Value().([]dbus.ObjectPath)

	items = make([]*Item, len(paths))

	for idx, path := range paths {
		if item, err = NewItem(c, path); err != nil {
			errs = append(errs, err)
			err = nil
			continue
		}
		items[idx] = item
	}
	err = NewErrors(err)

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
func (c *Collection) SearchItems(profile string) (items []*Item, err error) {

	var paths []dbus.ObjectPath
	var errs []error = make([]error, 0)
	var attrs map[string]string = make(map[string]string, 0)

	attrs["profile"] = profile

	if err = c.Dbus.Call(
		DbusCollectionSearchItems, 0, attrs,
	).Store(&paths); err != nil {
		return
	}

	items = make([]*Item, len(paths))

	for idx, path := range paths {
		if items[idx], err = NewItem(c, path); err != nil {
			errs = append(errs, err)
			err = nil
			continue
		}
	}
	err = NewErrors(err)

	return
}

// CreateItem returns a pointer to an Item based on a label, some attributes, a Secret, and whether any existing secret with the same label should be replaced or not.
func (c *Collection) CreateItem(label string, attrs map[string]string, secret *Secret, replace bool) (item *Item, err error) {

	var prompt *Prompt
	var path dbus.ObjectPath
	var promptPath dbus.ObjectPath
	var variant *dbus.Variant
	var props map[string]dbus.Variant = make(map[string]dbus.Variant)

	props[DbusItemLabel] = dbus.MakeVariant(label)
	props[DbusItemAttributes] = dbus.MakeVariant(attrs)

	if err = c.Dbus.Call(
		DbusCollectionCreateItem, 0, props, secret, replace,
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

	item, err = NewItem(c, path)

	return
}

// Locked indicates if a Collection is locked (true) or unlocked (false).
func (c *Collection) Locked() (isLocked bool, err error) {

	var variant dbus.Variant

	if variant, err = c.Dbus.GetProperty(DbusCollectionLocked); err != nil {
		isLocked = true
		return
	}

	isLocked = variant.Value().(bool)

	return
}

// Label returns the Collection label (name).
func (c *Collection) Label() (label string, err error) {

	var variant dbus.Variant

	if variant, err = c.Dbus.GetProperty(DbusCollectionLabel); err != nil {
		return
	}

	label = variant.Value().(string)

	if label != c.name {
		c.name = label
	}

	return
}

// Relabel modifies the Collection's label in Dbus.
func (c *Collection) Relabel(newLabel string) (err error) {

	var variant dbus.Variant = dbus.MakeVariant(newLabel)

	if err = c.Dbus.SetProperty(DbusItemLabel, variant); err != nil {
		return
	}

	return
}

// Created returns the time.Time of when a Collection was created.
func (c *Collection) Created() (created time.Time, err error) {

	var variant dbus.Variant
	var timeInt uint64

	if variant, err = c.Dbus.GetProperty(DbusCollectionCreated); err != nil {
		return
	}

	timeInt = variant.Value().(uint64)

	created = time.Unix(int64(timeInt), 0)

	return
}

/*
	Modified returns the time.Time of when a Collection was last modified along with a boolean
	that indicates if the collection has changed since the last call of Collection.Modified.

	Note that when calling NewCollection, the internal library-tracked modification
	time (Collection.lastModified) will be set to the latest modification time of the Collection
	itself as reported by Dbus rather than the time that NewCollection was called.
*/
func (c *Collection) Modified() (modified time.Time, isChanged bool, err error) {

	var variant dbus.Variant
	var timeInt uint64

	if variant, err = c.Dbus.GetProperty(DbusCollectionModified); err != nil {
		return
	}

	timeInt = variant.Value().(uint64)

	modified = time.Unix(int64(timeInt), 0)

	if !c.lastModifiedSet {
		// It's "nil", so set it to modified. We can't check for a zero-value in case Dbus has it as a zero-value.
		c.lastModified = modified
		c.lastModifiedSet = true
	}

	isChanged = modified.After(c.lastModified)
	c.lastModified = modified

	return
}
