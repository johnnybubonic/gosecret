package gosecret

import (
	`github.com/godbus/dbus`
)

// NewSession returns a pointer to a new Session based on a Dbus connection and a Dbus path.
func NewSession(conn *dbus.Conn, path dbus.ObjectPath) (session *Session) {

	session = &Session{
		Conn: conn,
		Dbus: conn.Object(DBusServiceName, path),
	}

	return
}

// Path returns the path of the underlying Dbus connection.
func (s Session) Path() (path dbus.ObjectPath) {

	// Remove this method in V1. It's bloat since we now have an exported Dbus.
	path = s.Dbus.Path()

	return
}
