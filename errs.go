package gosecret

import (
	`errors`
)

// Errors.
var (
	// ErrBadDbusPath indicates an invalid path - either nothing exists at that path or the path is malformed.
	ErrBadDbusPath error = errors.New("invalid dbus path")
	// ErrInvalidProperty indicates a dbus.Variant is not the "real" type expected.
	ErrInvalidProperty error = errors.New("invalid variant type; cannot convert")
	// ErrNoDbusConn gets triggered if a connection to Dbus can't be detected.
	ErrNoDbusConn  error = errors.New("no valid dbus connection")
)
