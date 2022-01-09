package gosecret

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/godbus/dbus/v5"
	`r00t2.io/goutils/multierr`
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

	if err = s.Session.Close(); err != nil {
		return
	}

	if err = s.Conn.Close(); err != nil {
		return
	}

	return
}

/*
	Collections returns a slice of Collection items accessible to this Service.

	err MAY be a *multierr.MultiError.
*/
func (s *Service) Collections() (collections []*Collection, err error) {

	var paths []dbus.ObjectPath
	var variant dbus.Variant
	var coll *Collection
	var errs *multierr.MultiError = multierr.NewMultiError()

	if variant, err = s.Dbus.GetProperty(DbusServiceCollections); err != nil {
		return
	}

	paths = variant.Value().([]dbus.ObjectPath)

	collections = make([]*Collection, 0)

	for _, path := range paths {
		coll = nil
		if coll, err = NewCollection(s, path); err != nil {
			errs.AddError(err)
			err = nil
			continue
		}
		collections = append(collections, coll)
	}

	if !errs.IsEmpty() {
		err = errs
	}

	return
}

/*
	CreateAliasedCollection creates a new Collection (keyring) via a Service with the name specified by label,
	aliased to the name specified by alias, and returns the new Collection.
*/
func (s *Service) CreateAliasedCollection(label, alias string) (collection *Collection, err error) {

	var call *dbus.Call
	var variant *dbus.Variant
	var path dbus.ObjectPath
	var promptPath dbus.ObjectPath
	var prompt *Prompt
	var props map[string]dbus.Variant = make(map[string]dbus.Variant)

	props[DbusCollectionLabel] = dbus.MakeVariant(label)
	props[DbusCollectionCreated] = dbus.MakeVariant(uint64(time.Now().Unix()))
	props[DbusCollectionModified] = dbus.MakeVariant(uint64(time.Now().Unix()))

	if call = s.Dbus.Call(
		DbusServiceCreateCollection, 0, props, alias,
	); call.Err != nil {
		err = call.Err
		return
	}
	if err = call.Store(&path, &promptPath); err != nil {
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

	err MAY be a *multierr.MultiError.
*/
func (s *Service) GetCollection(name string) (c *Collection, err error) {

	var errs *multierr.MultiError = multierr.NewMultiError()
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
			errs.AddError(err)
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
	if !errs.IsEmpty() {
		err = errs
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
	var call *dbus.Call

	if itemPaths == nil || len(itemPaths) == 0 {
		err = ErrMissingPaths
		return
	}

	secrets = make(map[dbus.ObjectPath]*Secret, len(itemPaths))
	results = make(map[dbus.ObjectPath][]interface{}, len(itemPaths))

	// TODO: trigger a Service.Unlock for any locked items?
	if call = s.Dbus.Call(
		DbusServiceGetSecrets, 0, itemPaths, s.Session.Dbus.Path(),
	); call.Err != nil {
		err = call.Err
		return
	}
	if err = call.Store(&results); err != nil {
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

// Scrapping this idea for now; it would require introspection on a known Item path.
/*
	IsLegacy indicates with a decent likelihood of accuracy if this Service is
	connected to a legacy spec Secret Service (true) or if the spec is current (false).

	It also returns a confidence indicator as a float, which indicates how accurate
	the guess (because it is a guess) may/is likely to be (as a percentage). For example,
	if confidence is expressed as 0.25, the result of legacyAPI has a 25% of being accurate.
*/
/*
func (s *Service) IsLegacy() (legacyAPI bool, confidence int) {

	var maxCon int

	// Test 1, property introspection on Item. We're looking for a Type property.
	DbusInterfaceItem

	return
}
*/

// Lock locks an Unlocked Collection or Item (LockableObject).
func (s *Service) Lock(objects ...LockableObject) (err error) {

	var call *dbus.Call
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

	if call = s.Dbus.Call(
		DbusServiceLock, 0, toLock,
	); call.Err != nil {
		err = call.Err
		return
	}
	if err = call.Store(&locked, &promptPath); err != nil {
		return
	}

	if isPrompt(promptPath) {

		prompt = NewPrompt(s.Conn, promptPath)

		if _, err = prompt.Prompt(); err != nil {
			return
		}
	}

	// TODO: use channels and goroutines here.
	for _, o := range objects {
		if _, err = o.Locked(); err != nil {
			return
		}
	}

	return
}

/*
	OpenSession returns a pointer to a Session from the Service.
	It's a convenience function around NewSession.
	However, NewService attaches a Session by default at Service.Session so this is likely unnecessary.
*/
func (s *Service) OpenSession(algo, input string) (session *Session, output dbus.Variant, err error) {

	var call *dbus.Call
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
	if call = s.Dbus.Call(
		DbusServiceOpenSession, 0, algo, inputVariant,
	); call.Err != nil {
		err = call.Err
		return
	}
	if err = call.Store(&output, &path); err != nil {
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

	var call *dbus.Call
	var objectPath dbus.ObjectPath

	if call = s.Dbus.Call(
		DbusServiceReadAlias, 0, alias,
	); call.Err != nil {
		err = call.Err
		return
	}
	/*
		TODO: Confirm that a nonexistent alias will NOT cause an error to return.
			  If it does, alter the below logic.
	*/
	if err = call.Store(&objectPath); err != nil {
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

	if err = s.SetAlias(alias, DbusRemoveAliasPath); err != nil {
		return
	}

	return
}

/*
	SearchItems searches all Collection objects and returns all matches based on the map of attributes.

	err MAY be a *multierr.MultiError.
*/
func (s *Service) SearchItems(attributes map[string]string) (unlockedItems []*Item, lockedItems []*Item, err error) {

	var call *dbus.Call
	var locked []dbus.ObjectPath
	var unlocked []dbus.ObjectPath
	var collectionObjs []*Collection
	var collections map[dbus.ObjectPath]*Collection = make(map[dbus.ObjectPath]*Collection, 0)
	var ok bool
	var c *Collection
	var cPath dbus.ObjectPath
	var item *Item
	var errs *multierr.MultiError = multierr.NewMultiError()

	if attributes == nil || len(attributes) == 0 {
		err = ErrMissingAttrs
		return
	}

	if call = s.Dbus.Call(
		DbusServiceSearchItems, 0, attributes,
	); call.Err != nil {
	}
	if err = call.Store(&unlocked, &locked); err != nil {
		return
	}

	lockedItems = make([]*Item, 0)
	unlockedItems = make([]*Item, 0)

	if collectionObjs, err = s.Collections(); err != nil {
		return
	}

	for _, c = range collectionObjs {
		if _, ok = collections[c.Dbus.Path()]; !ok {
			collections[c.Dbus.Path()] = c
		}
	}

	// Locked items
	for _, i := range locked {

		item = nil
		cPath = dbus.ObjectPath(filepath.Dir(string(i)))

		if c, ok = collections[cPath]; !ok {
			errs.AddError(errors.New(fmt.Sprintf(
				"could not find matching Collection for locked item %v", string(i),
			)))
			continue
		}

		if item, err = NewItem(c, i); err != nil {
			errs.AddError(errors.New(fmt.Sprintf(
				"could not create Item for locked item %v; error follows", string(i),
			)))
			errs.AddError(err)
			err = nil
			continue
		}
		lockedItems = append(lockedItems, item)
	}

	// Unlocked items
	for _, i := range unlocked {

		item = nil
		cPath = dbus.ObjectPath(filepath.Dir(string(i)))

		if c, ok = collections[cPath]; !ok {
			errs.AddError(errors.New(fmt.Sprintf(
				"could not find matching Collection for unlocked item %v", string(i),
			)))
			continue
		}

		if item, err = NewItem(c, i); err != nil {
			errs.AddError(errors.New(fmt.Sprintf(
				"could not create Item for unlocked item %v; error follows", string(i),
			)))
			errs.AddError(err)
			err = nil
			continue
		}
		unlockedItems = append(unlockedItems, item)
	}

	if !errs.IsEmpty() {
		err = errs
	}

	return
}

/*
	SetAlias sets an alias for an existing Collection.
	(You can get its path via <Collection>.Dbus.Path().)
	To remove an alias, set objectPath to dbus.ObjectPath("/").
*/
func (s *Service) SetAlias(alias string, objectPath dbus.ObjectPath) (err error) {

	var call *dbus.Call
	var collection *Collection

	if collection, err = s.GetCollection(alias); err != nil {
		return
	}

	if call = s.Dbus.Call(
		DbusServiceSetAlias, 0, alias, objectPath,
	); call.Err != nil {
		err = call.Err
		return
	}

	if objectPath == DbusRemoveAliasPath {
		collection.Alias = ""
	} else {
		collection.Alias = alias
	}

	return
}

// Unlock unlocks a locked Collection or Item (LockableObject).
func (s *Service) Unlock(objects ...LockableObject) (err error) {

	var call *dbus.Call
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

	if call = s.Dbus.Call(
		DbusServiceUnlock, 0, toUnlock,
	); call.Err != nil {
		err = call.Err
		return
	}
	if err = call.Store(&unlocked, &resultPath); err != nil {
		return
	}

	if isPrompt(resultPath) {

		prompt = NewPrompt(s.Conn, resultPath)

		if _, err = prompt.Prompt(); err != nil {
			return
		}
	}

	// TODO: use channels and goroutines here.
	for _, o := range objects {
		if _, err = o.Locked(); err != nil {
			return
		}
	}

	return
}

// path is a *very* thin wrapper around Service.Dbus.Path().
func (s *Service) path() (dbusPath dbus.ObjectPath) {

	dbusPath = s.Dbus.Path()

	return
}
