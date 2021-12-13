package gosecret

import (
	"github.com/godbus/dbus/v5"
)

// I'm still not 100% certain what Sessions are used for, aside from getting Secrets from Items.

/*
	NewSession returns a pointer to a new Session based on a Service and a dbus.ObjectPath.
	You will almost always want to use Service.GetSession or Service.OpenSession instead.
*/
func NewSession(service *Service, path dbus.ObjectPath) (session *Session, err error) {

	if _, err = validConnPath(service.Conn, path); err != nil {
		return
	}

	var ssn Session = Session{
		DbusObject: &DbusObject{
			Conn: service.Conn,
		},
		service: service,
	}
	ssn.Dbus = ssn.Conn.Object(DbusInterfaceSession, path)

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

// path is a *very* thin wrapper around Session.Dbus.Path().
func (s *Session) path() (dbusPath dbus.ObjectPath) {

	dbusPath = s.Dbus.Path()

	return
}
