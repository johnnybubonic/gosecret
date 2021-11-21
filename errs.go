package libsecret

import (
	`errors`
)

var (
	ErrNoDbusConn error = errors.New("no valid dbus connection")
	ErrBadDbusPath error = errors.New("invalid dbus path")
)
