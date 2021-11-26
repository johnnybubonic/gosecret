package gosecret

// Libsecret/SecretService/Dbus identifiers.
const (
	// DbusServiceName is the "root Dbus path" in identifier format.
	DbusServiceName string = "org.freedesktop.secrets"
	// DbusItemsID is the Dbus identifier for Item.
	DbusItemsID string = "org.freedesktop.Secret.Collection.Items"
	// DbusCollectionDelete is the Dbus identifier for Collection.Delete.
	DbusCollectionDelete string = "org.freedesktop.Secret.Collection.Delete"
)

// Dbus constants and paths.
const (
	// DbusPath is the path version of DbusServiceName.
	DbusPath     string = "/org/freedesktop/secrets"
	// PromptPrefix is the path used for prompts comparison.
	PromptPrefix string = DbusPath + "/prompt/"
)
