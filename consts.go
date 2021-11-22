package gosecret

// Libsecret/SecretService identifiers.
const (
	DbusItemsID string = "org.freedesktop.Secret.Collection.Items"
	DbusCollectionDelete string = "org.freedesktop.Secret.Collection.Delete"
)

// Dbus constants
const (
	DBusServiceName string = "org.freedesktop.secrets"
	DBusPath        string = "/org/freedesktop/secrets"
	PromptPrefix    string = DBusPath + "/prompt/"
)
