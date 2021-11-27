package gosecret

// Libsecret/SecretService Dbus interfaces.
const (
	// DbusService is the Dbus service bus identifier.
	DbusService string = "org.freedesktop.secrets"
	// DbusServiceBase is the base identifier used by interfaces.
	DbusServiceBase string = "org.freedesktop.Secret"
)

// Service interface.
const (
	/*
		DbusInterfaceService is the Dbus interface for working with a Service.
		Found at /org/freedesktop/secrets/(DbusInterfaceService)
	*/
	DbusInterfaceService string = DbusServiceBase + ".Service"

	// Methods

	// DbusServiceChangeLock is [FUNCTION UNKNOWN; TODO.]
	DbusServiceChangeLock string = DbusInterfaceService + ".ChangeLock"

	// DbusServiceCreateCollection is used to create a new Collection if it doesn't exist in Dbus.
	DbusServiceCreateCollection string = DbusInterfaceService + ".CreateCollection"

	/*
		DbusServiceGetSecrets is used to fetch all Secret / Item items in a given Collection
		(via Service.GetSecrets).
	*/
	DbusServiceGetSecrets string = DbusInterfaceService + ".GetSecrets"

	// DbusServiceLock is used by Service.Lock.
	DbusServiceLock string = DbusInterfaceService + ".Lock"

	// DbusServiceLockService is [FUNCTION UNKNOWN; TODO.]
	DbusServiceLockService string = DbusInterfaceService + ".LockService"

	// DbusServiceOpenSession is used by Service.Open.
	DbusServiceOpenSession string = DbusInterfaceService + ".OpenSession"

	// DbusServiceReadAlias is used by Service.GetAlias to return a Collection based on its aliased name.
	DbusServiceReadAlias string = DbusInterfaceService + ".ReadAlias"

	// DbusServiceSearchItems is used by Service.SearchItems to get arrays of locked and unlocked Item objects.
	DbusServiceSearchItems string = DbusInterfaceService + ".SearchItems"

	// DbusServiceSetAlias is used by Service.SetAlias to set an alias for a Collection.
	DbusServiceSetAlias string = DbusInterfaceService + ".SetAlias"

	// DbusServiceUnlock is used to unlock a Service.
	DbusServiceUnlock string = DbusInterfaceService + ".Unlock"

	// Properties

	// DbusServiceCollections is used to get a Dbus array of Collection items.
	DbusServiceCollections string = DbusInterfaceService + ".Collections"
)

// Session interface.
const (
	/*
		DbusInterfaceSession is the Dbus interface for working with a Session.
		Found at /org/freedesktop/secrets/session/<session ID>/(DbusInterfaceSession)
	*/
	DbusInterfaceSession = DbusServiceBase +  ".Session"

	// Methods

	// DbusSessionClose is used for Session.Close.
	DbusSessionClose string = DbusInterfaceSession + ".Close"
)

// Collection interface.
const (
	/*
		DbusInterfaceCollection is the Dbus interface for working with a Collection.
		Found at /org/freedesktop/secrets/collection/<collection name>/(DbusInterfaceCollection)
	*/
	DbusInterfaceCollection string = DbusServiceBase + ".Collection"

	// Methods

	// DbusCollectionCreateItem is used for Collection.CreateItem.
	DbusCollectionCreateItem string = DbusInterfaceCollection + ".CreateItem"

	// DbusCollectionDelete is used for Collection.Delete.
	DbusCollectionDelete string = DbusInterfaceCollection + ".Delete"

	// DbusCollectionSearchItems is used for Collection.SearchItems.
	DbusCollectionSearchItems string = DbusInterfaceCollection + ".SearchItems"

	// Properties

	// DbusCollectionItems is a Dbus array of Item.
	DbusCollectionItems string = DbusInterfaceCollection + ".Items"

	// DbusCollectionLocked is a Dbus boolean for Collection.Locked.
	DbusCollectionLocked string = DbusInterfaceCollection + ".Locked"

	// DbusCollectionLabel is the name (label) for Collection.Label.
	DbusCollectionLabel string = DbusInterfaceCollection + ".Label"

	// DbusCollectionCreated is the time a Collection was created (in a UNIX Epoch uint64) for Collection.Created.
	DbusCollectionCreated string = DbusInterfaceCollection + ".Created"

	// DbusCollectionModified is the time a Collection was last modified (in a UNIX Epoch uint64) for Collection.Modified.
	DbusCollectionModified string = DbusInterfaceCollection + ".Modified"

	// TODO: Signals?
)

// Item interface.
const (
	/*
		DbusInterfaceItem is the Dbus interface for working with Item items.
		Found at /org/freedesktop/secrets/collection/<collection name>/<item index>/(DbusInterfaceItem)
	*/
	DbusInterfaceItem string = DbusServiceBase + ".Item"

	// Methods

	// DbusItemDelete is used by Item.Delete.
	DbusItemDelete string = DbusInterfaceItem + ".Delete"

	// DbusItemGetSecret is used by Item.GetSecret.
	DbusItemGetSecret string = DbusInterfaceItem + ".GetSecret"

	// DbusItemSetSecret is used by Item.SetSecret.
	DbusItemSetSecret string = DbusInterfaceItem + ".SetSecret"
)

// Dbus paths.
const (
	// DbusPath is the path for DbusService.
	DbusPath string = "/org/freedesktop/secrets"
	// DbusPromptPrefix is the path used for prompts comparison.
	DbusPromptPrefix string = DbusPath + "/prompt/"
)
