package gosecret

import (
	"github.com/godbus/dbus/v5"
)

// Constants for use with gosecret.
const (
	/*
		ExplicitAttrEmptyValue is the constant used in Item.ModifyAttributes to explicitly set a value as empty.
		Between the surrounding with %'s, the weird name that includes "gosecret", and the UUID4...
		I am fairly confident this is unique enough.
	*/
	ExplicitAttrEmptyValue string = "%EXPLICIT_GOSECRET_BLANK_VALUE_8A4E3D7D-F30E-4754-8C56-9C172D1400F6%"
)

// Libsecret/SecretService Dbus interfaces.
const (
	// DbusService is the Dbus service bus identifier.
	DbusService string = "org.freedesktop.secrets"
	// DbusServiceBase is the base identifier used by interfaces.
	DbusServiceBase string = "org.freedesktop.Secret"
	// DbusPrompterInterface is an interface for issuing a Prompt. Yes, it should be doubled up like that.
	DbusPrompterInterface string = DbusServiceBase + ".Prompt.Prompt"
	/*
		DbusDefaultItemType is the default type to use for Item.Type/Collection.CreateItem.
	*/
	DbusDefaultItemType string = DbusServiceBase + ".Generic"
)

// Libsecret/SecretService special values.
var (
	// DbusRemoveAliasPath is used to remove an alias from a Collection and/or Item.
	DbusRemoveAliasPath dbus.ObjectPath = dbus.ObjectPath("/")
)

// Service interface.
const (
	/*
		DbusInterfaceService is the Dbus interface for working with a Service.
		Found at /org/freedesktop/secrets/(DbusInterfaceService)
	*/
	DbusInterfaceService string = DbusServiceBase + ".Service"

	// Methods

	/*
		DbusServiceChangeLock has some references in the SecretService Dbus API but
		it seems to be obsolete - undocumented, at the least.
		So we don't implement it.
	*/
	// DbusServiceChangeLock string = DbusInterfaceService + ".ChangeLock"

	// DbusServiceCreateCollection is used to create a new Collection via Service.CreateCollection.
	DbusServiceCreateCollection string = DbusInterfaceService + ".CreateCollection"

	// DbusServiceGetSecrets is used to fetch multiple Secret values from multiple Item items in a given Collection (via Service.GetSecrets).
	DbusServiceGetSecrets string = DbusInterfaceService + ".GetSecrets"

	// DbusServiceLock is used by Service.Lock.
	DbusServiceLock string = DbusInterfaceService + ".Lock"

	// DbusServiceLockService is [FUNCTION UNKNOWN/UNDOCUMENTED; TODO? NOT IMPLEMENTED.]
	// DbusServiceLockService string = DbusInterfaceService + ".LockService"

	// DbusServiceOpenSession is used by Service.OpenSession.
	DbusServiceOpenSession string = DbusInterfaceService + ".OpenSession"

	// DbusServiceReadAlias is used by Service.ReadAlias to return a Collection based on its aliased name.
	DbusServiceReadAlias string = DbusInterfaceService + ".ReadAlias"

	// DbusServiceSearchItems is used by Service.SearchItems to get arrays of locked and unlocked Item objects.
	DbusServiceSearchItems string = DbusInterfaceService + ".SearchItems"

	// DbusServiceSetAlias is used by Service.SetAlias to set an alias for a Collection.
	DbusServiceSetAlias string = DbusInterfaceService + ".SetAlias"

	// DbusServiceUnlock is used by Service.Unlock.
	DbusServiceUnlock string = DbusInterfaceService + ".Unlock"

	// Properties

	// DbusServiceCollections is used to get a Dbus array of Collection items (Service.Collections).
	DbusServiceCollections string = DbusInterfaceService + ".Collections"
)

// Session interface.
const (
	/*
		DbusInterfaceSession is the Dbus interface for working with a Session.
		Found at /org/freedesktop/secrets/session/<session ID>/(DbusInterfaceSession)
	*/
	DbusInterfaceSession = DbusServiceBase + ".Session"

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

	// Properties

	// DbusItemLocked is a Dbus boolean for Item.Locked.
	DbusItemLocked string = DbusInterfaceItem + ".Locked"

	/*
		DbusItemAttributes contains attributes (metadata, schema, etc.) for
		Item.Attributes, Item.ReplaceAttributes, and Item.ModifyAttributes.
	*/
	DbusItemAttributes string = DbusInterfaceItem + ".Attributes"

	// DbusItemLabel is the name (label) for Item.Label.
	DbusItemLabel string = DbusInterfaceItem + ".Label"

	// DbusItemType is the type of Item (Item.ItemType).
	DbusItemType string = DbusInterfaceItem + ".Type"

	// DbusItemCreated is the time an Item was created (in a UNIX Epoch uint64) for Item.Created.
	DbusItemCreated string = DbusInterfaceItem + ".Created"

	// DbusItemModified is the time an Item was last modified (in a UNIX Epoch uint64) for Item.Modified.
	DbusItemModified string = DbusInterfaceItem + ".Modified"
)

// Dbus paths.
const (
	// DbusPath is the path for DbusService.
	DbusPath string = "/org/freedesktop/secrets"
	// DbusPromptPrefix is the path used for prompts comparison.
	DbusPromptPrefix string = DbusPath + "/prompt/"
	// DbusNewCollectionPath is used to create a new Collection.
	DbusNewCollectionPath string = DbusPath + "/collection/"
	// DbusNewSessionPath is used to create a new Session.
	DbusNewSessionPath string = DbusPath + "/session/"
)

// FLAGS
// These are not currently used, but may be in the future.

// SERVICE

// ServiceInitFlag is a flag for Service.OpenSession.
type ServiceInitFlag int

const (
	FlagServiceNone ServiceInitFlag = iota
	FlagServiceOpenSession
	FlagServiceLoadCollections
)

// ServiceSearchFlag is a flag for Service.SearchItems.
type ServiceSearchFlag int

const (
	FlagServiceSearchNone ServiceSearchFlag = iota
	FlagServiceSearchAll
	FlagServiceSearchUnlock
	FlagServiceSearchLoadSecrets
)

// COLLECTION

// CollectionInitFlag is a flag for Collection.SearchItems and Collection.Items.
type CollectionInitFlag int

const (
	FlagCollectionNone CollectionInitFlag = iota
	FlagCollectionLoadItems
)

// ITEM

// ItemInitFlag are flags for Collection.SearchItems and Collection.Items.
type ItemInitFlag int

const (
	FlagItemNone ItemInitFlag = iota
	FlagItemLoadSecret
)

// ItemSearchFlag are flags for Collection.CreateItem.
type ItemSearchFlag int

const (
	FlagItemCreateNone ItemSearchFlag = iota
	FlatItemCreateReplace
)

// ERRORS

/*
	SecretServiceErrEnum are just constants for the enum'd errors;
	see SecretServiceError type and ErrSecretService* vars for what
	actually gets returned.
	They're used for finding the appropriate matching error.
*/
type SecretServiceErrEnum int

const (
	EnumErrProtocol SecretServiceErrEnum = iota
	EnumErrIsLocked
	EnumErrNoSuchObject
	EnumErrAlreadyExists
	EnumErrInvalidFileFormat
)
