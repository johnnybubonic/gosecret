package gosecret

import (
	`github.com/godbus/dbus`
)

// NewService returns a pointer to a new Service.
func NewService() (service *Service, err error) {

	service = &Service{
		Conn: nil,
		Dbus: nil,
	}

	if service.Conn, err = dbus.SessionBus(); err != nil {
		return
	}
	service.Dbus = service.Conn.Object(DbusServiceName, dbus.ObjectPath(DbusPath))

	return
}

// Path returns the path of the underlying Dbus connection.
func (s Service) Path() (path dbus.ObjectPath) {

	// Remove this method in V1. It's bloat since we now have an exported Dbus.
	path = s.Dbus.Path()

	return
}

// Open returns a pointer to a Session from the Service.
func (s *Service) Open() (session *Session, err error) {

	var output dbus.Variant
	var path dbus.ObjectPath

	if err = s.Dbus.Call(
		"org.freedesktop.Secret.Service.OpenSession", 0, "plain", dbus.MakeVariant(""),
	).Store(&output, &path); err != nil {
		return
	}

	session = NewSession(s.Conn, path)

	return
}

// Collections returns a slice of Collection keyrings accessible to this Service.
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

// CreateCollection creates a new Collection (keyring) via a Service with the name specified by label.
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
