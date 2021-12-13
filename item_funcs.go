package gosecret

import (
	`strconv`
	`strings`
	`time`

	`github.com/godbus/dbus/v5`
)

// NewItem returns a pointer to an Item based on Collection and a Dbus path.
func NewItem(collection *Collection, path dbus.ObjectPath) (item *Item, err error) {

	var splitPath []string

	if collection == nil {
		err = ErrNoDbusConn
	}

	if _, err = validConnPath(collection.Conn, path); err != nil {
		return
	}

	item = &Item{
		DbusObject: &DbusObject{
			Conn: collection.Conn,
			Dbus: collection.Conn.Object(DbusService, path),
		},
	}

	splitPath = strings.Split(string(item.Dbus.Path()), "/")

	item.idx, err = strconv.Atoi(splitPath[len(splitPath)-1])
	item.collection = collection

	// Populate the struct fields...
	// TODO: use channel for errors; condense into a MultiError.
	go item.GetSecret(collection.service.Session)
	go item.Locked()
	go item.Attributes()
	go item.Label()
	go item.Type()
	go item.Created()
	go item.Modified()

	return
}

// Attributes returns the Item's attributes from Dbus.
func (i *Item) Attributes() (attrs map[string]string, err error) {

	var variant dbus.Variant

	if variant, err = i.Dbus.GetProperty(DbusItemAttributes); err != nil {
		return
	}

	attrs = variant.Value().(map[string]string)
	i.Attrs = attrs

	return
}

/*
	ChangeItemType changes an Item.Type to newItemType.
	Note that this is probably a bad idea unless you're also doing Item.SetSecret.
	It must be a Dbus interface path (e.g. "foo.bar.Baz").
	If newItemType is an empty string, DbusDefaultItemType will be used.
*/
func (i *Item) ChangeItemType(newItemType string) (err error) {

	var variant dbus.Variant

	if strings.TrimSpace(newItemType) == "" {
		newItemType = DbusDefaultItemType
	}

	variant = dbus.MakeVariant(newItemType)

	if err = i.Dbus.SetProperty(DbusItemType, variant); err != nil {
		return
	}
	i.SecretType = newItemType

	if err = i.setModify(); err != nil {
		return
	}

	return
}

// Delete removes an Item from a Collection.
func (i *Item) Delete() (err error) {

	var promptPath dbus.ObjectPath
	var prompt *Prompt

	if err = i.Dbus.Call(DbusItemDelete, 0).Store(&promptPath); err != nil {
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

// GetSecret returns the Secret in an Item using a Session.
func (i *Item) GetSecret(session *Session) (secret *Secret, err error) {

	if session == nil {
		err = ErrNoDbusConn
	}

	if _, err = connIsValid(session.Conn); err != nil {
		return
	}

	if err = i.Dbus.Call(
		DbusItemGetSecret, 0, session.Dbus.Path(),
	).Store(&secret); err != nil {
		return
	}

	secret.session = session
	secret.item = i
	i.Secret = secret

	return
}

// Label returns the label ("name") of an Item.
func (i *Item) Label() (label string, err error) {

	var variant dbus.Variant

	if variant, err = i.Dbus.GetProperty(DbusItemLabel); err != nil {
		return
	}

	label = variant.Value().(string)

	return
}

/*
	ModifyAttributes modifies the Item's attributes in Dbus.
	This is similar to Item.ReplaceAttributes but will only modify the map's given keys so you do not need to provide
	the entire attribute map.
	If you wish to remove an attribute, use the value "" (empty string).
	If you wish to explicitly provide a blank value/empty string, use the constant gosecret.ExplicitAttrEmptyValue.

	This is more or less a convenience/wrapper function around Item.ReplaceAttributes.
*/
func (i *Item) ModifyAttributes(replaceAttrs map[string]string) (err error) {

	var ok bool
	var currentProps map[string]string = make(map[string]string, 0)
	var currentVal string

	if replaceAttrs == nil || len(replaceAttrs) == 0 {
		err = ErrMissingAttrs
		return
	}

	if currentProps, err = i.Attributes(); err != nil {
		return
	}

	for k, v := range replaceAttrs {
		if currentVal, ok = currentProps[k]; !ok { // If it isn't in the replacement map, do nothing (i.e. keep it).
			continue
		} else if v == currentVal { // If the value is the same, do nothing.
			continue
		} else if v == ExplicitAttrEmptyValue { // If it's the "magic empty value" constant, delete the key/value pair.
			delete(currentProps, k)
			continue
		} else { // Otherwise, replace the value.
			currentProps[k] = v
		}
	}

	if err = i.ReplaceAttributes(currentProps); err != nil {
		return
	}

	return
}

// Relabel modifies the Item's label in Dbus.
func (i *Item) Relabel(newLabel string) (err error) {

	var variant dbus.Variant = dbus.MakeVariant(newLabel)

	if err = i.Dbus.SetProperty(DbusItemLabel, variant); err != nil {
		return
	}
	i.LabelName = newLabel

	if err = i.setModify(); err != nil {
		return
	}

	return
}

// ReplaceAttributes replaces the Item's attributes in Dbus.
func (i *Item) ReplaceAttributes(newAttrs map[string]string) (err error) {

	var props dbus.Variant

	props = dbus.MakeVariant(newAttrs)

	if err = i.Dbus.SetProperty(DbusItemAttributes, props); err != nil {
		return
	}
	i.Attrs = newAttrs

	if err = i.setModify(); err != nil {
		return
	}

	return
}

// SetSecret sets the Secret for an Item.
func (i *Item) SetSecret(secret *Secret) (err error) {

	var c *dbus.Call

	c = i.Dbus.Call(
		DbusItemSetSecret, 0,
	)
	if c.Err != nil {
		err = c.Err
		return
	}
	i.Secret = secret

	if err = i.setModify(); err != nil {
		return
	}

	return
}

// Type updates the Item.ItemType from DBus (and returns it).
func (i *Item) Type() (itemType string, err error) {

	var variant dbus.Variant

	if variant, err = i.Dbus.GetProperty(DbusItemType); err != nil {
		return
	}

	itemType = variant.Value().(string)
	i.SecretType = itemType

	return
}

// Lock will lock an unlocked Item. It will no-op if the Item is currently locked.
func (i *Item) Lock() (err error) {

	if _, err = i.Locked(); err != nil {
		return
	}
	if i.IsLocked {
		return
	}

	if err = i.collection.service.Lock(i); err != nil {
		return
	}
	i.IsLocked = true

	return
}

// Locked indicates if an Item is locked (true) or unlocked (false).
func (i *Item) Locked() (isLocked bool, err error) {

	var variant dbus.Variant

	if variant, err = i.Dbus.GetProperty(DbusItemLocked); err != nil {
		isLocked = true
		return
	}

	isLocked = variant.Value().(bool)
	i.IsLocked = isLocked

	return
}

// Unlock will unlock a locked Item. It will no-op if the Item is currently unlocked.
func (i *Item) Unlock() (err error) {

	if _, err = i.Locked(); err != nil {
		return
	}
	if !i.IsLocked {
		return
	}

	if err = i.collection.service.Unlock(i); err != nil {
		return
	}
	i.IsLocked = false

	return
}

// Created returns the time.Time of when an Item was created.
func (i *Item) Created() (created time.Time, err error) {

	var variant dbus.Variant
	var timeInt uint64

	if variant, err = i.Dbus.GetProperty(DbusItemCreated); err != nil {
		return
	}

	timeInt = variant.Value().(uint64)

	created = time.Unix(int64(timeInt), 0)
	i.CreatedAt = created

	return
}

/*
	Modified returns the time.Time of when an Item was last modified along with a boolean
	that indicates if the collection has changed since the last call of Item.Modified.

	Note that when calling NewItem, the internal library-tracked modification
	time (Item.LastModified) will be set to the latest modification time of the Item
	itself as reported by Dbus rather than the time that NewItem was called.
*/
func (i *Item) Modified() (modified time.Time, isChanged bool, err error) {

	var variant dbus.Variant
	var timeInt uint64

	if variant, err = i.Dbus.GetProperty(DbusItemModified); err != nil {
		return
	}

	timeInt = variant.Value().(uint64)

	modified = time.Unix(int64(timeInt), 0)

	if !i.lastModifiedSet {
		// It's "nil", so set it to modified. We can't check for a zero-value in case Dbus has it as a zero-value.
		i.LastModified = modified
		i.lastModifiedSet = true
	}

	isChanged = modified.After(i.LastModified)
	i.LastModified = modified

	return
}

/*
	setCreate updates the Item's creation time (as specified by Item.Created).
	It seems that this does not generate automatically.
*/
func (i *Item) setCreate() (err error) {

	var t time.Time = time.Now()

	if err = i.Dbus.SetProperty(DbusItemCreated, uint64(t.Unix())); err != nil {
		return
	}
	i.CreatedAt = t

	if err = i.setModify(); err != nil {
		return
	}

	return
}

/*
	setModify updates the Item's modification time (as specified by Item.Modified).
	It seems that this does not update automatically.
*/
func (i *Item) setModify() (err error) {

	var t time.Time = time.Now()

	err = i.Dbus.SetProperty(DbusItemModified, uint64(t.Unix()))
	i.LastModified = t

	return
}

// path is a *very* thin wrapper around Item.Dbus.Path(). It is needed for LockableObject membership.
func (i *Item) path() (dbusPath dbus.ObjectPath) {

	dbusPath = i.Dbus.Path()

	return
}
