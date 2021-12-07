package gosecret

import (
	"time"

	"github.com/godbus/dbus/v5"
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
	// Session is a default Session initiated automatically.
	Session *Session `json:"-"`
}

/*
	Session is a session/instance/connection to SecretService.
	https://developer-old.gnome.org/libsecret/0.18/SecretService.html
	https://specifications.freedesktop.org/secret-service/latest/ch06.html
*/
type Session struct {
	*DbusObject
	// collection tracks the Service this Session was created from.
	service *Service
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
	// lastModifiedSet is unexported; it's only used to determine if this is a first-initialization of the modification time or not.
	lastModifiedSet bool
	// name is used for the Collection's name/label so the Dbus path doesn't need to be parsed all the time.
	name string
	// service tracks the Service this Collection was created from.
	service *Service
}

/*
	Item is an entry in a Collection that contains a Secret.
	https://developer-old.gnome.org/libsecret/0.18/SecretItem.html
	https://specifications.freedesktop.org/secret-service/latest/re03.html
*/
type Item struct {
	*DbusObject
	// Secret is the corresponding Secret object.
	Secret *Secret `json:"secret"`
	/*
		ItemType is the type of this Item as a Dbus interface name.
		e.g. org.gnome.keyring.NetworkPassword, org.freedesktop.Secret.Generic, org.remmina.Password, etc.
	*/
	ItemType string `json:"dbus_type"`
	// lastModified is unexported because it's important that API users don't change it; it's used by Collection.Modified.
	lastModified time.Time
	// lastModifiedSet is unexported; it's only used to determine if this is a first-initialization of the modification time or not.
	lastModifiedSet bool
	/*
		idx is the index identifier of the Item.
		It SHOULD correlate to indices in Collection.Items, but don't rely on this.
	*/
	idx int
	// collection tracks the Collection this Item is in.
	collection *Collection
}

/*
	Secret is the "Good Stuff" - the actual secret content.
	https://developer-old.gnome.org/libsecret/0.18/SecretValue.html
	https://specifications.freedesktop.org/secret-service/latest/re03.html
	https://specifications.freedesktop.org/secret-service/latest/ch14.html#type-Secret
*/
type Secret struct {
	// Session is a Dbus object path for the associated Session (the actual Session is stored in an unexported field).
	Session dbus.ObjectPath `json:"session_path"`
	/*
		Parameters are "algorithm dependent parameters for secret value encoding" - likely this will just be an empty byteslice.
		Refer to Session for more information.
	*/
	Parameters []byte `json:"params"`
	// Value is the secret's content in []byte format.
	Value SecretValue `json:"value"`
	// ContentType is the MIME type of Value.
	ContentType string `json:"content_type"`
	// item is the Item this Secret belongs to.
	item *Item
	// session is the Session used to decode/decrypt this Secret.
	session *Session
}

// SecretValue is a custom type that handles JSON encoding/decoding a little more easily.
type SecretValue []byte
