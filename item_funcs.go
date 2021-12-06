package gosecret

import (
	`strconv`
	`strings`
	`time`

	`github.com/godbus/dbus/v5`
)

// TODO: add method Relabel

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
	if _, err = item.Attributes(); err != nil {
		return
	}
	if _, err = item.Type(); err != nil {
		return
	}

	_, _, err = item.Modified()

	return
}

// Attributes updates the Item.Attrs from Dbus (and returns them).
func (i *Item) Attributes() (attrs map[string]string, err error) {

	var variant dbus.Variant

	if variant, err = i.Dbus.GetProperty(DbusItemAttributes); err != nil {
		return
	}

	i.Attrs = variant.Value().(map[string]string)
	attrs = i.Attrs

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
	ModifyAttributes modifies the Item.Attrs, both in the object and in Dbus.
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

	err = i.ReplaceAttributes(currentProps)

	return
}

// ReplaceAttributes replaces the Item.Attrs, both in the object and in Dbus.
func (i *Item) ReplaceAttributes(newAttrs map[string]string) (err error) {

	var label string
	var props map[string]dbus.Variant = make(map[string]dbus.Variant, 0)

	if label, err = i.Label(); err != nil {
		return
	}

	props[DbusItemLabel] = dbus.MakeVariant(label)
	props[DbusItemAttributes] = dbus.MakeVariant(newAttrs)

	if err = i.Dbus.SetProperty(DbusItemAttributes, props); err != nil {
		return
	}

	i.Attrs = newAttrs

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

	return
}

// Type updates the Item.ItemType from DBus (and returns it).
func (i *Item) Type() (itemType string, err error) {

	var variant dbus.Variant

	if variant, err = i.Dbus.GetProperty(DbusItemType); err != nil {
		return
	}

	i.ItemType = variant.Value().(string)
	itemType = i.ItemType

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

	return
}

/*
	Modified returns the time.Time of when an Item was last modified along with a boolean
	that indicates if the collection has changed since the last call of Item.Modified.

	Note that when calling NewItem, the internal library-tracked modification
	time (Item.lastModified) will be set to the latest modification time of the Item
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
		i.lastModified = modified
		i.lastModifiedSet = true
	}

	isChanged = modified.After(i.lastModified)
	i.lastModified = modified

	return
}
