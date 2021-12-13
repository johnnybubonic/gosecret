package gosecret

import (
	"time"

	"github.com/godbus/dbus/v5"
)

/*
	NewCollection returns a pointer to a Collection based on a Service and a Dbus path.
	You will almost always want to use Service.GetCollection instead.
*/
func NewCollection(service *Service, path dbus.ObjectPath) (coll *Collection, err error) {

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
		// LastModified: time.Now(),
	}

	// Populate the struct fields...
	// TODO: use channel for errors; condense into a MultiError.
	go coll.Locked()
	go coll.Label()
	go coll.Created()
	go coll.Modified()

	return
}

/*
	CreateItem returns a pointer to an Item based on a label, some attributes, a Secret,
	whether any existing secret with the same label should be replaced or not, and the optional itemType.

	itemType is optional; if specified, it should be a Dbus interface (only the first element is used).
	If not specified, the default DbusDefaultItemType will be used. The most common itemType is DbusDefaultItemType
	and is the current recommendation.
	Other types used are:

		org.gnome.keyring.NetworkPassword
		org.gnome.keyring.Note

	These are libsecret schemas as defined at
	https://gitlab.gnome.org/GNOME/libsecret/-/blob/master/libsecret/secret-schemas.c (and bundled in with libsecret).
	Support for adding custom schemas MAY come in the future but is unsupported currently.
*/
func (c *Collection) CreateItem(label string, attrs map[string]string, secret *Secret, replace bool, itemType ...string) (item *Item, err error) {

	var prompt *Prompt
	var path dbus.ObjectPath
	var promptPath dbus.ObjectPath
	var variant *dbus.Variant
	var props map[string]dbus.Variant = make(map[string]dbus.Variant)
	var typeString string

	if itemType != nil && len(itemType) > 0 {
		typeString = itemType[0]
	} else {
		typeString = DbusDefaultItemType
	}

	props[DbusItemLabel] = dbus.MakeVariant(label)
	props[DbusItemType] = dbus.MakeVariant(typeString)
	props[DbusItemAttributes] = dbus.MakeVariant(attrs)

	if err = c.Dbus.Call(
		DbusCollectionCreateItem, 0, props, secret, replace,
	).Store(&path, &promptPath); err != nil {
		return
	}
	if err = c.setModify(); err != nil {
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
	if err = item.setCreate(); err != nil {
		return
	}

	return
}

/*
	Delete removes a Collection.
	While *technically* not necessary, it is recommended that you iterate through
	Collection.Items and do an Item.Delete for each item *before* calling Collection.Delete;
	the item paths are cached as "orphaned paths" in Dbus otherwise if not deleted before deleting
	their Collection. They should clear on a reboot or restart of Dbus (but rebooting Dbus on a system in use is... troublesome).
*/
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

// Label returns the Collection label (name).
func (c *Collection) Label() (label string, err error) {

	var variant dbus.Variant

	if variant, err = c.Dbus.GetProperty(DbusCollectionLabel); err != nil {
		return
	}

	label = variant.Value().(string)

	c.LabelName = label

	return
}

// Lock will lock an unlocked Collection. It will no-op if the Collection is currently locked.
func (c *Collection) Lock() (err error) {

	if _, err = c.Locked(); err != nil {
		return
	}
	if c.IsLocked {
		return
	}

	if err = c.service.Lock(c); err != nil {
		return
	}
	c.IsLocked = true

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
	c.IsLocked = isLocked

	return
}

// Relabel modifies the Collection's label in Dbus.
func (c *Collection) Relabel(newLabel string) (err error) {

	var variant dbus.Variant = dbus.MakeVariant(newLabel)

	if err = c.Dbus.SetProperty(DbusCollectionLabel, variant); err != nil {
		return
	}
	c.LabelName = newLabel

	return
}

/*
	SearchItems searches a Collection for a matching profile string.
	It's mostly a carry-over from go-libsecret, and is here for convenience. IT MAY BE REMOVED IN THE FUTURE.

	I promise it's not useful for any other implementation/storage of SecretService whatsoever.

	Deprecated: Use Service.SearchItems instead.
*/
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

// SetAlias is a thin wrapper/shorthand for Service.SetAlias (but specific to this Collection).
func (c *Collection) SetAlias(alias string) (err error) {

	var call *dbus.Call

	call = c.service.Dbus.Call(
		DbusServiceSetAlias, 0, alias, c.Dbus.Path(),
	)

	err = call.Err

	return
}

// Unlock will unlock a locked Collection. It will no-op if the Collection is currently unlocked.
func (c *Collection) Unlock() (err error) {

	if _, err = c.Locked(); err != nil {
		return
	}
	if !c.IsLocked {
		return
	}

	if err = c.service.Unlock(c); err != nil {
		return
	}
	c.IsLocked = false

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
	time (Collection.LastModified) will be set to the latest modification time of the Collection
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
		c.LastModified = modified
		c.lastModifiedSet = true
	}

	isChanged = modified.After(c.LastModified)
	c.LastModified = modified

	return
}

/*
	setCreate updates the Collection's creation time (as specified by Collection.Created).
	It seems that this does not generate automatically.
*/
func (c *Collection) setCreate() (err error) {

	var t time.Time = time.Now()

	if err = c.Dbus.SetProperty(DbusCollectionCreated, uint64(t.Unix())); err != nil {
		return
	}
	c.CreatedAt = t

	if err = c.setModify(); err != nil {
		return
	}

	return
}

/*
	setModify updates the Collection's modification time (as specified by Collection.Modified).
	It seems that this does not update automatically.
*/
func (c *Collection) setModify() (err error) {

	var t time.Time = time.Now()

	err = c.Dbus.SetProperty(DbusCollectionModified, uint64(t.Unix()))
	c.LastModified = t

	return
}

// path is a *very* thin wrapper around Collection.Dbus.Path(). It is needed for LockableObject interface membership.
func (c *Collection) path() (dbusPath dbus.ObjectPath) {

	dbusPath = c.Dbus.Path()

	return
}
