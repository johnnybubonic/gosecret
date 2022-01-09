package gosecret

import (
	"time"

	"github.com/godbus/dbus/v5"
)

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

type LockableObject interface {
	Locked() (bool, error)
	path() dbus.ObjectPath
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
	// IsLocked indicates if the Service is locked or not. Status updated by Service.Locked.
	IsLocked bool `json:"locked"`
	/*
		Legacy indicates that this SecretService implementation breaks current spec
		by implementing the legacy/obsolete draft spec rather than current libsecret spec
		for the Dbus API.

		If you're using SecretService with KeePassXC, for instance, or a much older version
		of Gnome-Keyring *before* libsecret integration(?),	or if you are getting strange errors
		when performing a Service.SearchItems, you probably need to enable this field on the
		Service returned by NewService. The coverage of this field may expand in the future, but
		currently it only prevents/suppresses the (non-existent, in legacy spec) Type property
		from being read or written on Items during NewItem and Collection.CreateItem respectively.
	*/
	Legacy bool `json:"is_legacy"`
}

/*
	Session is a session/instance/connection to SecretService.
	https://developer-old.gnome.org/libsecret/0.18/SecretService.html
	https://specifications.freedesktop.org/secret-service/latest/ch06.html
*/
type Session struct {
	*DbusObject
	// service tracks the Service this Session was created from.
	service *Service
}

/*
	Collection is an accessor for libsecret collections, which contain multiple Secret Item items.
	Do not change any of these values directly; use the associated methods instead.
	Reference:
	https://developer-old.gnome.org/libsecret/0.18/SecretCollection.html
	https://specifications.freedesktop.org/secret-service/latest/ch03.html
*/
type Collection struct {
	*DbusObject
	// IsLocked indicates if the Collection is locked or not. Status updated by Collection.Locked.
	IsLocked bool `json:"locked"`
	// LabelName is the Collection's label (as given by Collection.Label and modified by Collection.Relabel).
	LabelName string `json:"label"`
	// CreatedAt is when this Collection was created (used by Collection.Created).
	CreatedAt time.Time `json:"created"`
	// LastModified is when this Item was last changed; it's used by Collection.Modified.
	LastModified time.Time `json:"modified"`
	// Alias is the Collection's alias (as handled by Service.ReadAlias and Service.SetAlias).
	Alias string `json:"alias"`
	// lastModifiedSet is unexported; it's only used to determine if this is a first-initialization of the modification time or not.
	lastModifiedSet bool
	// service tracks the Service this Collection was created from.
	service *Service
}

/*
	Item is an entry in a Collection that contains a Secret. Do not change any of these values directly; use the associated methods instead.
	https://developer-old.gnome.org/libsecret/0.18/SecretItem.html
	https://specifications.freedesktop.org/secret-service/latest/re03.html
*/
type Item struct {
	*DbusObject
	// Secret is the corresponding Secret object.
	Secret *Secret `json:"secret"`
	// IsLocked indicates if the Item is locked or not. Status updated by Item.Locked.
	IsLocked bool
	// Attrs are the Item's attributes (as would be returned via Item.Attributes).
	Attrs map[string]string `json:"attributes"`
	// LabelName is the Item's label (as given by Item.Label and modified by Item.Relabel).
	LabelName string `json:"label"`
	// SecretType is the Item's secret type (as returned by Item.Type).
	SecretType string `json:"type"`
	// CreatedAt is when this Item was created (used by Item.Created).
	CreatedAt time.Time `json:"created"`
	// LastModified is when this Item was last changed; it's used by Item.Modified.
	LastModified time.Time `json:"modified"`
	// lastModifiedSet is unexported; it's only used to determine if this is a first-initialization of the modification time or not.
	lastModifiedSet bool
	/*
		idx is the index identifier of the Item.
		It is almost guaranteed to not match the index in Collection.Items (unless you have like, only one item)
		as those indices are static and do not determine the order that Dbus returns the list of item paths.
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

// SecretValue is a custom type that handles JSON encoding a little more easily.
type SecretValue []byte
