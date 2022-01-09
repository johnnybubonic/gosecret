package gosecret

import (
	`strings`

	`github.com/godbus/dbus/v5`
	`r00t2.io/goutils/multierr`
)

// isPrompt returns a boolean that is true if path is/requires a prompt(ed path) and false if it is/does not.
func isPrompt(path dbus.ObjectPath) (prompt bool) {

	prompt = strings.HasPrefix(string(path), DbusPromptPrefix)

	return

}

// connIsValid returns a boolean if the dbus.conn named conn is active.
func connIsValid(conn *dbus.Conn) (ok bool, err error) {

	// dbus.Conn.Names() will ALWAYS return a []string with at least ONE element.
	if conn == nil || (conn.Names() == nil || len(conn.Names()) < 1) {
		err = ErrNoDbusConn
		return
	}

	ok = true

	return
}

/*
	pathIsValid implements path checking for valid Dbus paths. Currently it only checks to make sure path is not a blank string.
	The path argument can be either a string or dbus.ObjectPath.
*/
func pathIsValid(path interface{}) (ok bool, err error) {

	var realPath string

	switch p := path.(type) {
	case dbus.ObjectPath:
		if !p.IsValid() {
			err = ErrBadDbusPath
			return
		}
		realPath = string(p)
	case string:
		realPath = p
	default:
		err = ErrBadDbusPath
		return
	}

	if strings.TrimSpace(realPath) == "" {
		err = ErrBadDbusPath
		return
	}

	ok = true

	return
}

/*
	validConnPath condenses the checks for connIsValid and pathIsValid into one func due to how frequently this check is done.

	If err is not nil, it IS a *multierr.MultiError.
*/
func validConnPath(conn *dbus.Conn, path interface{}) (cr *ConnPathCheckResult, err error) {

	var errs *multierr.MultiError = multierr.NewMultiError()

	cr = new(ConnPathCheckResult)

	if cr.ConnOK, err = connIsValid(conn); err != nil {
		errs.AddError(err)
		err = nil
	}
	if cr.PathOK, err = pathIsValid(path); err != nil {
		errs.AddError(err)
		err = nil
	}

	if !errs.IsEmpty() {
		err = errs
	}

	return
}

/*
	pathsFromProp returns a slice of dbus.ObjectPath (paths) from a dbus.Variant (prop).
	If prop cannot typeswitch to paths, an ErrInvalidProperty will be raised.
*/
func pathsFromProp(prop dbus.Variant) (paths []dbus.ObjectPath, err error) {

	switch v := prop.Value().(type) {
	case []dbus.ObjectPath:
		paths = v
	default:
		err = ErrInvalidProperty
		return
	}

	return
}

/*
	pathsFromPath returns a slice of dbus.ObjectPath based on an object given by path using the dbus.Conn specified by conn.
	Internally it uses pathsFromProp.
*/
func pathsFromPath(bus dbus.BusObject, path string) (paths []dbus.ObjectPath, err error) {

	var v dbus.Variant

	if v, err = bus.GetProperty(path); err != nil {
		return
	}

	if paths, err = pathsFromProp(v); err != nil {
		return
	}

	return
}

/*
	NameFromPath returns an actual name (as it appears in Dbus) from a dbus.ObjectPath.
	Note that you can get any object's dbus.ObjectPath via <object>.Dbus.Path().
	path is validated to ensure it is not an empty string.
*/
func NameFromPath(path dbus.ObjectPath) (name string, err error) {

	var strSplit []string
	var ok bool

	if ok, err = pathIsValid(path); err != nil {
		return
	} else if !ok {
		err = ErrBadDbusPath
		return
	}

	strSplit = strings.Split(string(path), "/")

	if len(strSplit) < 1 {
		err = ErrBadDbusPath
		return
	}

	name = strSplit[len(strSplit)-1]

	return
}
