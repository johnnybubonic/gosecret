package gosecret

import (
	`github.com/google/uuid`
)

// Paths.
const (
	DbusDefaultCollectionPath string = DbusPath + "/collections/login"
)

// Strings.
const (
	defaultCollection string = "default" // SHOULD point to a collection named "login"; "default" is the alias.
	testAlias         string = "GOSECRET_TESTING_ALIAS"
	testSecretContent string = "This is a test secret for gosecret."
	testItemLabel     string = "gosecret_test_item"
)

// Objects.
var (
	collectionName  uuid.UUID = uuid.New()
	collectionAlias uuid.UUID = uuid.New()
)
