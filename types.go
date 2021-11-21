package libsecret

import (
	`github.com/godbus/dbus`
)

/*
	Collection is an accessor for libsecret collections, which contain multiple Secret items.
	Reference:
	https://developer-old.gnome.org/libsecret/0.18/SecretCollection.html
	https://specifications.freedesktop.org/secret-service/latest/ch03.html
*/
type Collection struct {
	Conn *dbus.Conn
	Dbus dbus.BusObject
}
