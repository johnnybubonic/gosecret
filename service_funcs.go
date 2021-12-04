package gosecret

import (
	"github.com/godbus/dbus"
)

// NewService returns a pointer to a new Service connection.
func NewService() (service *Service, err error) {

	var svc Service = Service{}

	if svc.Conn, err = dbus.SessionBus(); err != nil {
		return
	}
	svc.Dbus = service.Conn.Object(DbusService, dbus.ObjectPath(DbusPath))

	if svc.Session, _, err = svc.Open(); err != nil {
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
func (s *Service) Collections() (collections []Collection, err error) {

	var paths []dbus.ObjectPath
	var variant dbus.Variant
	var coll *Collection
	var errs []error = make([]error, 0)

	if variant, err = s.Dbus.GetProperty(DbusServiceCollections); err != nil {
		return
	}

	paths = variant.Value().([]dbus.ObjectPath)

	collections = make([]Collection, len(paths))

	for idx, path := range paths {
		if coll, err = NewCollection(s.Conn, path); err != nil {
			// return
			errs = append(errs, err)
			err = nil
		}
		collections[idx] = *coll
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

	collection, err = NewCollection(s.Conn, path)

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
	GetAlias allows one to fetch a Collection dbus.ObjectPath based on an alias name.
	If the alias does not exist, objectPath will be dbus.ObjectPath("/").
	TODO: return a Collection instead of a dbus.ObjectPath.
*/
func (s *Service) GetAlias(alias string) (objectPath dbus.ObjectPath, err error) {

	err = s.Dbus.Call(
		DbusServiceReadAlias, 0, alias,
	).Store(&objectPath)

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
	/*
		// TODO: make any errs in here a MultiError instead.
		for _, secretPath := range itemPaths {
			if err = s.Dbus.Call(
				DbusServiceGetSecrets, 0, secretPath,
			).Store(&result); err != nil {
				return
			}
		}
	*/
	if err = s.Dbus.Call(
		DbusServiceGetSecrets, 0, itemPaths,
	).Store(&secrets); err != nil {
		return
	}

	return
}

/*
	Lock locks an Unlocked Service, Collection, etc.
	You can usually get objectPath for the object(s) to unlock via <object>.Dbus.Path().
	If objectPaths is nil or empty, the Service's own path will be used.
*/
func (s *Service) Lock(objectPaths ...dbus.ObjectPath) (err error) {

	var locked []dbus.ObjectPath
	var prompt *Prompt
	var resultPath dbus.ObjectPath

	if objectPaths == nil || len(objectPaths) == 0 {
		objectPaths = []dbus.ObjectPath{s.Dbus.Path()}
	}

	// TODO: make any errs in here a MultiError instead.
	for _, p := range objectPaths {
		if err = s.Dbus.Call(
			DbusServiceLock, 0, p,
		).Store(&locked, &resultPath); err != nil {
			return
		}

		if isPrompt(resultPath) {

			prompt = NewPrompt(s.Conn, resultPath)

			if _, err = prompt.Prompt(); err != nil {
				return
			}
		}
	}

	return
}

/*
	Open returns a pointer to a Session from the Service.
	It's a convenience function around NewSession.
*/
func (s *Service) Open() (session *Session, output dbus.Variant, err error) {

	var path dbus.ObjectPath

	// In *theory*, SecretService supports multiple "algorithms" for encryption in-transit, but I don't think it's implemented (yet)?
	// TODO: confirm this.
	// Possible flags are dbus.Flags consts: https://pkg.go.dev/github.com/godbus/dbus#Flags
	// Oddly, there is no "None" flag. So it's explicitly specified as a null byte.
	if err = s.Dbus.Call(
		DbusServiceOpenSession, 0, "plain", dbus.MakeVariant(""),
	).Store(&output, &path); err != nil {
		return
	}

	session = NewSession(s, path)

	return
}

/*
	SearchItems searches all Collection objects and returns all matches based on the map of attributes.
	TODO: return arrays of Items instead of dbus.ObjectPaths.
	TODO: check attributes for empty/nil.
*/
func (s *Service) SearchItems(attributes map[string]string) (unlockedItems []dbus.ObjectPath, lockedItems []dbus.ObjectPath, err error) {

	err = s.Dbus.Call(
		DbusServiceSearchItems, 0, attributes,
	).Store(&unlockedItems, &lockedItems)

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

	_ = c

	return
}

/*
	Unlock unlocks a Locked Service, Collection, etc.
	You can usually get objectPath for the object(s) to unlock via <object>.Dbus.Path().
	If objectPaths is nil or empty, the Service's own path will be used.
*/
func (s *Service) Unlock(objectPaths ...dbus.ObjectPath) (err error) {

	var unlocked []dbus.ObjectPath
	var prompt *Prompt
	var resultPath dbus.ObjectPath

	if objectPaths == nil || len(objectPaths) == 0 {
		objectPaths = []dbus.ObjectPath{s.Dbus.Path()}
	}

	// TODO: make any errs in here a MultiError instead.
	for _, p := range objectPaths {
		if err = s.Dbus.Call(
			DbusServiceUnlock, 0, p,
		).Store(&unlocked, &resultPath); err != nil {
			return
		}

		if isPrompt(resultPath) {

			prompt = NewPrompt(s.Conn, resultPath)

			if _, err = prompt.Prompt(); err != nil {
				return
			}
		}
	}

	return
}
