package gosecret

import (
	`github.com/google/uuid`
)

// Paths.
const (
	dbusCollectionPath        string = DbusPath + "/collection"
	dbusDefaultCollectionPath string = dbusCollectionPath + "/login"
)

// Strings.
const (
	defaultCollectionAlias string = "default" // SHOULD point to a collection named "login" (below); "default" is the alias.
	defaultCollection      string = "login"
	defaultCollectionLabel string = "Login" // a display name; the label is lowercased and normalized for the path (per above).
	testAlias              string = "GOSECRET_TESTING_ALIAS"
	testSecretContent      string = "This is a test secret for gosecret."
	testItemLabel          string = "Gosecret Test Item"
)

// Objects.
var (
	collectionName  uuid.UUID         = uuid.New()
	collectionAlias uuid.UUID         = uuid.New()
	itemAttrs       map[string]string = map[string]string{
		"GOSECRET": "yes",
		"foo":      "bar",
		"profile":  testItemLabel,
	}
)
