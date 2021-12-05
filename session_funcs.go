package gosecret

import (
	"github.com/godbus/dbus"
)

// I'm still not 100% certain what Sessions are used for?

/*
	NewSession returns a pointer to a new Session based on a Service and a dbus.ObjectPath.
	If path is empty (""), the default
*/
func NewSession(service *Service, path dbus.ObjectPath) (session *Session) {

	var ssn Session = Session{
		&DbusObject{
			Conn: service.Conn,
		},
	}
	session.Dbus = session.Conn.Object(DbusInterfaceSession, path)

	session = &ssn

	return
}

// Close cleanly closes a Session.
func (s *Session) Close() (err error) {

	var c *dbus.Call

	c = s.Dbus.Call(
		DbusSessionClose, 0,
	)

	_ = c

	return
}
