package gosecret

import (
	"time"

	"github.com/godbus/dbus"
)

/*
	MultiError is a type of error.Error that can contain multiple error.Errors. Confused? Don't worry about it.
*/
type MultiError struct {
	// Errors is a slice of errors to combine/concatenate when .Error() is called.
	Errors []error `json:"errors"`
	// ErrorSep is a string to use to separate errors for .Error(). The default is "\n".
	ErrorSep string `json:"separator"`
}

/*
	SecretServiceError is a translated error from SecretService API.
	See https://developer-old.gnome.org/libsecret/unstable/libsecret-SecretError.html#SecretError and
	ErrSecretService* errors.
*/
type SecretServiceError struct {
	// ErrCode is the SecretService API's enum value.
	ErrCode SecretServiceErrEnum `json:"code"`
	// ErrName is the SecretService API's error name.
	ErrName string `json:"name"`
	/*
		ErrDesc is the actual error description/text.
		This is what should be displayed to users, and is returned by SecretServiceError.Error.
	*/
	ErrDesc string `json:"desc"`
}

// ConnPathCheckResult contains the result of validConnPath.
type ConnPathCheckResult struct {
	// ConnOK is true if the dbus.Conn is valid.
	ConnOK bool `json:"conn"`
	// PathOK is true if the Dbus path given is a valid type and value.
	PathOK bool `json:"path"`
}

// DbusObject is a base struct type to be anonymized by other types.
type DbusObject struct {
	// Conn is an active connection to the Dbus.
	Conn *dbus.Conn `json:"-"`
	// Dbus is the Dbus bus object.
	Dbus dbus.BusObject `json:"-"`
}

/*
	Prompt is an interface to handling unlocking prompts.
	https://developer-old.gnome.org/libsecret/0.18/SecretPrompt.html
	https://specifications.freedesktop.org/secret-service/latest/ch09.html
*/
type Prompt struct {
	*DbusObject
}

/*
	Service is a general SecretService interface, sort of handler for Dbus - it's used for fetching a Session, Collections, etc.
	https://developer-old.gnome.org/libsecret/0.18/SecretService.html
	https://specifications.freedesktop.org/secret-service/latest/re01.html
*/
type Service struct {
	*DbusObject
	Session *Session `json:"-"`
}

/*
	Session is a session/instance/connection to SecretService.
	https://developer-old.gnome.org/libsecret/0.18/SecretService.html
	https://specifications.freedesktop.org/secret-service/latest/ch06.html
*/
type Session struct {
	*DbusObject
}

/*
	Collection is an accessor for libsecret collections, which contain multiple Secret Item items.
	Reference:
	https://developer-old.gnome.org/libsecret/0.18/SecretCollection.html
	https://specifications.freedesktop.org/secret-service/latest/ch03.html
*/
type Collection struct {
	*DbusObject
	// lastModified is unexported because it's important that API users don't change it; it's used by Collection.Modified.
	lastModified time.Time
}

/*
	Item is an entry in a Collection that contains a Secret.
	https://developer-old.gnome.org/libsecret/0.18/SecretItem.html
	https://specifications.freedesktop.org/secret-service/latest/re03.html
*/
type Item struct {
	*DbusObject
}

/*
	Secret is the "Good Stuff" - the actual secret content.
	https://developer-old.gnome.org/libsecret/0.18/SecretValue.html
	https://specifications.freedesktop.org/secret-service/latest/re03.html
	https://specifications.freedesktop.org/secret-service/latest/ch14.html#type-Secret
*/
type Secret struct {
	// Session is a Dbus object path for the associated Session.
	Session dbus.ObjectPath `json:"-"`
	// Parameters are "algorithm dependent parameters for secret value encoding" - likely this will just be an empty byteslice.
	Parameters []byte `json:"params"`
	// Value is the secret's content in []byte format.
	Value []byte `json:"value"`
	// ContentType is the MIME type of Value.
	ContentType string `json:"content_type"`
}
