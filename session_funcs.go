package gosecret

import (
	"fmt"

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

	var call *dbus.Call

	if call = s.Dbus.Call(
		DbusSessionClose, 0,
	); call.Err != nil {
		/*
			I... still haven't 100% figured out why this happens, but the session DOES seem to close...?
			PRs or input welcome.
			TODO: figure out why this error gets triggered.
		*/
		if call.Err.Error() != fmt.Sprintf("The name %v was not provided by any .service files", DbusInterfaceSession) {
			err = call.Err
			return
		}
	}

	return
}

// path is a *very* thin wrapper around Session.Dbus.Path().
func (s *Session) path() (dbusPath dbus.ObjectPath) {

	dbusPath = s.Dbus.Path()

	return
}
