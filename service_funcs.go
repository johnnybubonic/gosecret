package gosecret

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/godbus/dbus/v5"
)

// NewService returns a pointer to a new Service connection.
func NewService() (service *Service, err error) {

	var svc Service = Service{
		DbusObject: &DbusObject{
			Conn: nil,
			Dbus: nil,
		},
		Session: nil,
	}

	if svc.Conn, err = dbus.SessionBus(); err != nil {
		return
	}
	svc.Dbus = svc.Conn.Object(DbusService, dbus.ObjectPath(DbusPath))

	if svc.Session, err = svc.GetSession(); err != nil {
		return
	}

	service = &svc

	return
}

// Close cleanly closes a Service and all its underlying connections (e.g. Service.Session).
func (s *Service) Close() (err error) {

	err = s.Session.Close()

	return
}

// Collections returns a slice of Collection items accessible to this Service.
func (s *Service) Collections() (collections []*Collection, err error) {

	var paths []dbus.ObjectPath
	var variant dbus.Variant
	var coll *Collection
	var errs []error = make([]error, 0)

	if variant, err = s.Dbus.GetProperty(DbusServiceCollections); err != nil {
		return
	}

	paths = variant.Value().([]dbus.ObjectPath)

	collections = make([]*Collection, len(paths))

	for idx, path := range paths {
		if coll, err = NewCollection(s, path); err != nil {
			// return
			errs = append(errs, err)
			err = nil
			continue
		}
		collections[idx] = coll
	}
	err = NewErrors(err)

	return
}

/*
	CreateAliasedCollection creates a new Collection (keyring) via a Service with the name specified by label,
	aliased to the name specified by alias, and returns the new Collection.
*/
func (s *Service) CreateAliasedCollection(label, alias string) (collection *Collection, err error) {

	var variant *dbus.Variant
	var path dbus.ObjectPath
	var promptPath dbus.ObjectPath
	var prompt *Prompt
	var props map[string]dbus.Variant = make(map[string]dbus.Variant)

	props[DbusCollectionLabel] = dbus.MakeVariant(label)

	if err = s.Dbus.Call(
		DbusServiceCreateCollection, 0, props, alias,
	).Store(&path, &promptPath); err != nil {
		return
	}

	if isPrompt(promptPath) {

		prompt = NewPrompt(s.Conn, promptPath)
		if variant, err = prompt.Prompt(); err != nil {
			return
		}

		path = variant.Value().(dbus.ObjectPath)
	}

	collection, err = NewCollection(s, path)
	if err = collection.setCreate(); err != nil {
		return
	}

	return
}

/*
	CreateCollection creates a new Collection (keyring) via a Service with the name specified by label and returns the new Collection.
	It is a *very* thin wrapper around Service.CreateAliasedCollection, but with a blank alias.
*/
func (s *Service) CreateCollection(label string) (collection *Collection, err error) {

	collection, err = s.CreateAliasedCollection(label, "")

	return
}

/*
	GetCollection returns a single Collection based on the name (name can also be an alias).
	It's a helper function that avoids needing to make multiple calls in user code.
*/
func (s *Service) GetCollection(name string) (c *Collection, err error) {

	var errs []error = make([]error, 0)
	var colls []*Collection
	var pathName string

	// First check for an alias.
	if c, err = s.ReadAlias(name); err != nil && err != ErrDoesNotExist {
		c = nil
		return
	}
	if c != nil {
		return
	} else {
		c = nil
	}

	// We didn't get it by alias, so let's try by name...
	if colls, err = s.Collections(); err != nil {
		return
	}
	for _, i := range colls {
		if pathName, err = NameFromPath(i.Dbus.Path()); err != nil {
			errs = append(errs, err)
			err = nil
			continue
		}
		if pathName == name {
			c = i
			return
		}
	}

	// Still nothing? Try by label.
	for _, i := range colls {
		if i.LabelName == name {
			c = i
			return
		}
	}

	// Couldn't find it by the given name.
	if errs != nil || len(errs) > 0 {
		errs = append([]error{ErrDoesNotExist}, errs...)
		err = NewErrors(errs...)
	} else {
		err = ErrDoesNotExist
	}

	return
}

/*
	GetSecrets allows you to fetch values (Secret) from multiple Item object paths using this Service's Session.
	An ErrMissingPaths will be returned for err if itemPaths is nil or empty.
	The returned secrets is a map with itemPaths as the keys and their corresponding Secret as the value.
	If you know which Collection your desired Secret is in, it is recommended to iterate through Collection.Items instead
	(as Secrets returned here may have missing functionality).
*/
func (s *Service) GetSecrets(itemPaths ...dbus.ObjectPath) (secrets map[dbus.ObjectPath]*Secret, err error) {

	/*
		Results are in the form of a map with the value consisting of:
		[]interface {}{
			"/org/freedesktop/secrets/session/sNNN",  // 0, session path
			[]uint8{},                                // 1, "params"
			[]uint8{0x0},                             // 2, value
			"text/plain",                             // 3, content type
		}
	*/
	var results map[dbus.ObjectPath][]interface{}

	if itemPaths == nil || len(itemPaths) == 0 {
		err = ErrMissingPaths
		return
	}

	secrets = make(map[dbus.ObjectPath]*Secret, len(itemPaths))
	results = make(map[dbus.ObjectPath][]interface{}, len(itemPaths))

	// TODO: trigger a Service.Unlock for any locked items?
	if err = s.Dbus.Call(
		DbusServiceGetSecrets, 0, itemPaths, s.Session.Dbus.Path(),
	).Store(&results); err != nil {
		return
	}

	for p, r := range results {
		secrets[p] = NewSecret(
			s.Session, r[1].([]byte), r[2].([]byte), r[3].(string),
		)
	}

	return
}

/*
	GetSession returns a single Session.
	It's a helper function that wraps Service.OpenSession.
*/
func (s *Service) GetSession() (ssn *Session, err error) {

	ssn, _, err = s.OpenSession("", "")

	return
}

// Lock locks an Unlocked Collection or Item (LockableObject).
func (s *Service) Lock(objects ...LockableObject) (err error) {

	var variant dbus.Variant
	var toLock []dbus.ObjectPath
	// We only use these as destinations.
	var locked []dbus.ObjectPath
	var prompt *Prompt
	var promptPath dbus.ObjectPath

	if objects == nil || len(objects) == 0 {
		err = ErrMissingObj
		return
	}

	toLock = make([]dbus.ObjectPath, len(objects))

	for idx, o := range objects {
		toLock[idx] = o.path()
	}
	variant = dbus.MakeVariant(toLock)

	if err = s.Dbus.Call(
		DbusServiceLock, 0, variant,
	).Store(&locked, &promptPath); err != nil {
		return
	}

	if isPrompt(promptPath) {

		prompt = NewPrompt(s.Conn, promptPath)

		if _, err = prompt.Prompt(); err != nil {
			return
		}
	}

	for _, o := range objects {
		go o.Locked()
	}

	return
}

/*
	OpenSession returns a pointer to a Session from the Service.
	It's a convenience function around NewSession.
	However, NewService attaches a Session by default at Service.Session so this is likely unnecessary.
*/
func (s *Service) OpenSession(algo, input string) (session *Session, output dbus.Variant, err error) {

	var path dbus.ObjectPath
	var inputVariant dbus.Variant

	if strings.TrimSpace(algo) == "" {
		algo = "plain"
	}

	inputVariant = dbus.MakeVariant(input)

	// In *theory*, SecretService supports multiple "algorithms" for encryption in-transit, but I don't think it's implemented (yet)?
	// TODO: confirm this.
	// Possible flags are dbus.Flags consts: https://pkg.go.dev/github.com/godbus/dbus#Flags
	// Oddly, there is no "None" flag. So it's explicitly specified as a null byte.
	if err = s.Dbus.Call(
		DbusServiceOpenSession, 0, algo, inputVariant,
	).Store(&output, &path); err != nil {
		return
	}

	session, err = NewSession(s, path)

	return
}

/*
	ReadAlias allows one to fetch a Collection based on an alias name.
	An ErrDoesNotExist will be raised if the alias does not exist.
	You will almost assuredly want to use Service.GetCollection instead; it works for both alias names and real names.
*/
func (s *Service) ReadAlias(alias string) (collection *Collection, err error) {

	var objectPath dbus.ObjectPath

	err = s.Dbus.Call(
		DbusServiceReadAlias, 0, alias,
	).Store(&objectPath)

	/*
		TODO: Confirm that a nonexistent alias will NOT cause an error to return.
			  If it does, alter the below logic.
	*/
	if err != nil {
		return
	}

	// If the alias does not exist, objectPath will be dbus.ObjectPath("/").
	if objectPath == dbus.ObjectPath("/") {
		err = ErrDoesNotExist
		return
	}

	if collection, err = NewCollection(s, objectPath); err != nil {
		return
	}

	return
}

// RemoveAlias is a thin wrapper around Service.SetAlias using the removal method specified there.
func (s *Service) RemoveAlias(alias string) (err error) {

	if err = s.SetAlias(alias, dbus.ObjectPath("/")); err != nil {
		return
	}

	return
}

/*
	SearchItems searches all Collection objects and returns all matches based on the map of attributes.
*/
func (s *Service) SearchItems(attributes map[string]string) (unlockedItems []*Item, lockedItems []*Item, err error) {

	var locked []dbus.ObjectPath
	var unlocked []dbus.ObjectPath
	var collectionObjs []*Collection
	var collections map[dbus.ObjectPath]*Collection = make(map[dbus.ObjectPath]*Collection, 0)
	var ok bool
	var c *Collection
	var cPath dbus.ObjectPath
	var errs []error = make([]error, 0)

	if attributes == nil || len(attributes) == 0 {
		err = ErrMissingAttrs
		return
	}

	err = s.Dbus.Call(
		DbusServiceSearchItems, 0, attributes,
	).Store(&unlocked, &locked)

	lockedItems = make([]*Item, len(locked))
	unlockedItems = make([]*Item, len(unlocked))

	if collectionObjs, err = s.Collections(); err != nil {
		return
	}

	for _, c = range collectionObjs {
		if _, ok = collections[c.Dbus.Path()]; !ok {
			collections[c.Dbus.Path()] = c
		}
	}

	// Locked items
	for idx, i := range locked {

		cPath = dbus.ObjectPath(filepath.Dir(string(i)))

		if c, ok = collections[cPath]; !ok {
			errs = append(errs, errors.New(fmt.Sprintf(
				"could not find matching Collection for locked item %v", string(i),
			)))
			continue
		}

		if lockedItems[idx], err = NewItem(c, i); err != nil {
			errs = append(errs, errors.New(fmt.Sprintf(
				"could not create Item for locked item %v", string(i),
			)))
			err = nil
			continue
		}
	}

	// Unlocked items
	for idx, i := range unlocked {

		cPath = dbus.ObjectPath(filepath.Dir(string(i)))

		if c, ok = collections[cPath]; !ok {
			errs = append(errs, errors.New(fmt.Sprintf(
				"could not find matching Collection for unlocked item %v", string(i),
			)))
			continue
		}

		if unlockedItems[idx], err = NewItem(c, i); err != nil {
			errs = append(errs, errors.New(fmt.Sprintf(
				"could not create Item for unlocked item %v", string(i),
			)))
			err = nil
			continue
		}
	}

	if errs != nil && len(errs) > 0 {
		err = NewErrors(errs...)
	}

	return
}

/*
	SetAlias sets an alias for an existing Collection.
	(You can get its path via <Collection>.Dbus.Path().)
	To remove an alias, set objectPath to dbus.ObjectPath("/").
*/
func (s *Service) SetAlias(alias string, objectPath dbus.ObjectPath) (err error) {

	var c *dbus.Call

	c = s.Dbus.Call(
		DbusServiceSetAlias, 0, alias, objectPath,
	)

	err = c.Err

	return
}

// Unlock unlocks a locked Collection or Item (LockableObject).
func (s *Service) Unlock(objects ...LockableObject) (err error) {

	var variant dbus.Variant
	var toUnlock []dbus.ObjectPath
	// We only use these as destinations.
	var unlocked []dbus.ObjectPath
	var prompt *Prompt
	var resultPath dbus.ObjectPath

	if objects == nil || len(objects) == 0 {
		err = ErrMissingObj
		return
	}

	toUnlock = make([]dbus.ObjectPath, len(objects))

	for idx, o := range objects {
		toUnlock[idx] = o.path()
	}
	variant = dbus.MakeVariant(toUnlock)

	if err = s.Dbus.Call(
		DbusServiceUnlock, 0, variant,
	).Store(&unlocked, &resultPath); err != nil {
		return
	}

	if isPrompt(resultPath) {

		prompt = NewPrompt(s.Conn, resultPath)

		if _, err = prompt.Prompt(); err != nil {
			return
		}
	}

	for _, o := range objects {
		go o.Locked()
	}

	return
}

// path is a *very* thin wrapper around Service.Dbus.Path().
func (s *Service) path() (dbusPath dbus.ObjectPath) {

	dbusPath = s.Dbus.Path()

	return
}
