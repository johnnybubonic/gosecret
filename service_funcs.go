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

	var colls []*Collection

	// First check for an alias.
	if c, err = s.ReadAlias(name); err != nil && err != ErrDoesNotExist {
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
		if i.name == name {
			c = i
			return
		}
	}

	// Couldn't find it by the given name.
	err = ErrDoesNotExist

	return
}

/*
	GetSecrets allows you to fetch values (Secret) from multiple Item object paths using this Service's Session.
	An ErrMissingPaths will be returned for err if itemPaths is nil or empty.
	The returned secrets is a map with itemPaths as the keys and their corresponding Secret as the value.
*/
func (s *Service) GetSecrets(itemPaths ...dbus.ObjectPath) (secrets map[dbus.ObjectPath]*Secret, err error) {

	if itemPaths == nil || len(itemPaths) == 0 {
		err = ErrMissingPaths
		return
	}

	secrets = make(map[dbus.ObjectPath]*Secret, len(itemPaths))

	// TODO: trigger a Service.Unlock for any locked items?
	if err = s.Dbus.Call(
		DbusServiceGetSecrets, 0, itemPaths,
	).Store(&secrets); err != nil {
		return
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

/*
	Lock locks an Unlocked Service, Collection, etc.
	You can usually get objectPath for the object(s) to unlock via <object>.Dbus.Path().
	If objectPaths is nil or empty, the Service's own path will be used.
*/
func (s *Service) Lock(objectPaths ...dbus.ObjectPath) (err error) {

	var errs []error = make([]error, 0)
	// We only use these as destinations.
	var locked []dbus.ObjectPath
	var prompt *Prompt
	var resultPath dbus.ObjectPath

	if objectPaths == nil || len(objectPaths) == 0 {
		objectPaths = []dbus.ObjectPath{s.Dbus.Path()}
	}

	for _, p := range objectPaths {
		if err = s.Dbus.Call(
			DbusServiceLock, 0, p,
		).Store(&locked, &resultPath); err != nil {
			errs = append(errs, err)
			err = nil
			continue
		}

		if isPrompt(resultPath) {

			prompt = NewPrompt(s.Conn, resultPath)

			if _, err = prompt.Prompt(); err != nil {
				errs = append(errs, err)
				err = nil
				continue
			}
		}
	}

	if errs != nil && len(errs) > 0 {
		err = NewErrors(errs...)
	}

	return
}

/*
	OpenSession returns a pointer to a Session from the Service.
	It's a convenience function around NewSession.
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

/*
	Unlock unlocks a Locked Service, Collection, etc.
	You can usually get objectPath for the object(s) to unlock via <object>.Dbus.Path().
	If objectPaths is nil or empty, the Service's own path will be used.
*/
func (s *Service) Unlock(objectPaths ...dbus.ObjectPath) (err error) {

	var errs []error = make([]error, 0)
	var unlocked []dbus.ObjectPath
	var prompt *Prompt
	var resultPath dbus.ObjectPath

	if objectPaths == nil || len(objectPaths) == 0 {
		objectPaths = []dbus.ObjectPath{s.Dbus.Path()}
	}

	for _, p := range objectPaths {
		if err = s.Dbus.Call(
			DbusServiceUnlock, 0, p,
		).Store(&unlocked, &resultPath); err != nil {
			errs = append(errs, err)
			err = nil
			continue
		}

		if isPrompt(resultPath) {

			prompt = NewPrompt(s.Conn, resultPath)

			if _, err = prompt.Prompt(); err != nil {
				errs = append(errs, err)
				err = nil
				continue
			}
		}
	}

	if errs != nil && len(errs) > 0 {
		err = NewErrors(errs...)
	}

	return
}
