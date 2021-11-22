package gosecret

import (
	`github.com/godbus/dbus`
)

// DBusObject is any type that has a Path method that returns a dbus.ObjectPath.
type DBusObject interface {
	Path() dbus.ObjectPath
}

/*
	Collection is an accessor for libsecret collections, which contain multiple Secret Item items.
	Reference:
	https://developer-old.gnome.org/libsecret/0.18/SecretCollection.html
	https://specifications.freedesktop.org/secret-service/latest/ch03.html
*/
type Collection struct {
	// Conn is an active connection to the Dbus.
	Conn *dbus.Conn
	// Dbus is the Dbus bus object.
	Dbus dbus.BusObject
}

/*
	Item is an entry in a Collection that contains a Secret.
	https://developer-old.gnome.org/libsecret/0.18/SecretItem.html
	https://specifications.freedesktop.org/secret-service/latest/re03.html
*/
type Item struct {
	// Conn is an active connection to the Dbus.
	Conn *dbus.Conn
	// Dbus is the Dbus bus object.
	Dbus dbus.BusObject
}

/*
	Prompt is an interface to handling unlocking prompts.
	https://developer-old.gnome.org/libsecret/0.18/SecretPrompt.html
	https://specifications.freedesktop.org/secret-service/latest/ch09.html
*/
type Prompt struct {
	// Conn is an active connection to the Dbus.
	Conn *dbus.Conn
	// Dbus is the Dbus bus object.
	Dbus dbus.BusObject
}

/*
	Secret is the "Good Stuff" - the actual secret content.
	https://developer-old.gnome.org/libsecret/0.18/SecretValue.html
	https://specifications.freedesktop.org/secret-service/latest/re03.html
	https://specifications.freedesktop.org/secret-service/latest/ch14.html#type-Secret
*/
type Secret struct {
	// Session is a Dbus object path for the associated Session.
	Session dbus.ObjectPath
	// Parameters are "algorithm dependent parameters for secret value encoding" - likely this will just be an empty byteslice.
	Parameters []byte
	// Value is the secret's content in []byte format.
	Value []byte
	// ContentType is the MIME type of Value.
	ContentType string
}

/*
	Service is a general SecretService interface, sort of handler for Dbus - it's used for fetching a Session, Collections, etc.
	https://developer-old.gnome.org/libsecret/0.18/SecretService.html
	https://specifications.freedesktop.org/secret-service/latest/re01.html
*/
type Service struct {
	// Conn is an active connection to the Dbus.
	Conn *dbus.Conn
	// Dbus is the Dbus bus object.
	Dbus dbus.BusObject
}

/*
	Session is a session/instance/connection to SecretService.
	https://developer-old.gnome.org/libsecret/0.18/SecretService.html
	https://specifications.freedesktop.org/secret-service/latest/ch06.html
*/
type Session struct {
	// Conn is an active connection to the Dbus.
	Conn *dbus.Conn
	// Dbus is the Dbus bus object.
	Dbus dbus.BusObject
}
