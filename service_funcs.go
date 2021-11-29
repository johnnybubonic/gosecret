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

	if svc.Session, err = svc.Open(); err != nil {
		return
	}

	service = &svc

	return
}

// Open returns a pointer to a Session from the Service.
func (s *Service) Open() (session *Session, output dbus.Variant, err error) {

	var path dbus.ObjectPath

	// In *theory*, SecretService supports multiple "algorithms" for encryption in-transit, but I don't think it's implemented (yet)?
	// TODO: confirm this.
	// Possible flags are dbus.Flags consts: https://pkg.go.dev/github.com/godbus/dbus#Flags
	// Oddly, there is no "None" flag. So it's explicitly specified as a null byte.
	if err = s.Dbus.Call(
		DbusServiceOpenSession, 0x0, "plain", dbus.MakeVariant(""),
	).Store(&output, &path); err != nil {
		return
	}

	session = NewSession(s.Conn, path)

	return
}

// Collections returns a slice of Collection items accessible to this Service.
func (s *Service) Collections() (collections []Collection, err error) {

	var paths []dbus.ObjectPath
	var variant dbus.Variant

	if variant, err = s.Dbus.GetProperty("org.freedesktop.Secret.Service.Collections"); err != nil {
		return
	}

	paths = variant.Value().([]dbus.ObjectPath)

	collections = make([]Collection, len(paths))

	for idx, path := range paths {
		collections[idx] = *NewCollection(s.Conn, path)
	}

	return
}

// CreateCollection creates a new Collection (keyring) via a Service with the name specified by label and returns the new Collection.
func (s *Service) CreateCollection(label string) (collection *Collection, err error) {

	var variant *dbus.Variant
	var path dbus.ObjectPath
	var promptPath dbus.ObjectPath
	var prompt *Prompt
	var props map[string]dbus.Variant = make(map[string]dbus.Variant)

	props["org.freedesktop.Secret.Collection.Label"] = dbus.MakeVariant(label)

	if err = s.Dbus.Call("org.freedesktop.Secret.Service.CreateCollection", 0, props, "").Store(&path, &promptPath); err != nil {
		return
	}

	if isPrompt(promptPath) {

		prompt = NewPrompt(s.Conn, promptPath)

		if variant, err = prompt.Prompt(); err != nil {
			return
		}

		path = variant.Value().(dbus.ObjectPath)
	}

	collection = NewCollection(s.Conn, path)

	return
}

// Unlock unlocks a Locked Service.
func (s *Service) Unlock(object DBusObject) (err error) {

	var unlocked []dbus.ObjectPath
	var prompt *Prompt
	var promptPath dbus.ObjectPath
	var paths []dbus.ObjectPath = []dbus.ObjectPath{object.Path()}

	if err = s.Dbus.Call("org.freedesktop.Secret.Service.Unlock", 0, paths).Store(&unlocked, &promptPath); err != nil {
		return
	}

	if isPrompt(promptPath) {

		prompt = NewPrompt(s.Conn, promptPath)

		if _, err = prompt.Prompt(); err != nil {
			return
		}
	}

	return
}

// Lock locks an Unlocked Service.
func (s *Service) Lock(object DBusObject) (err error) {

	var locked []dbus.ObjectPath
	var prompt *Prompt
	var promptPath dbus.ObjectPath
	var paths []dbus.ObjectPath = []dbus.ObjectPath{object.Path()}

	if err = s.Dbus.Call("org.freedesktop.Secret.Service.Lock", 0, paths).Store(&locked, &promptPath); err != nil {
		return
	}

	if isPrompt(promptPath) {

		prompt = NewPrompt(s.Conn, promptPath)

		if _, err = prompt.Prompt(); err != nil {
			return
		}
	}

	return
}
