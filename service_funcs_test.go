package gosecret

import (
	"testing"

	"github.com/google/uuid"
)

var (
	collectionName  uuid.UUID = uuid.New()
	collectionAlias uuid.UUID = uuid.New()
)

const (
	defaultCollection string = "Login"
	testAlias         string = "GOSECRET_TESTING_ALIAS"
)

/*
	TestNewService tests the following internal functions/methods via nested calls:

		NewService
			Service.GetSession
				Service.OpenSession
					NewSession
						validConnPath
							connIsValid
							pathIsValid
		Service.Close
			Session.Close

*/
func TestNewService(t *testing.T) {

	var err error
	var svc *Service

	if svc, err = NewService(); err != nil {
		t.Fatalf("could not get new Service via NewService: %v", err.Error())
	}

	if err = svc.Close(); err != nil {
		t.Errorf("could not close Service.Session: %v", err.Error())
	}

}

/*
	TestService_Collections tests the following internal functions/methods via nested calls:

		(all calls in TestNewService)
		Service.Collections
			NewCollection
				Collection.Modified
			NewErrors
*/
func TestService_Collections(t *testing.T) {

	var err error
	var svc *Service
	var colls []*Collection

	if svc, err = NewService(); err != nil {
		t.Fatalf("could not get new Service via NewService: %v", err.Error())
	}

	if _, err = svc.Collections(); err != nil {
		t.Errorf("could not get Service.Collections: %v", err.Error())
	} else {
		t.Logf("found %v collections via Service.Collections", len(colls))
	}

	if err = svc.Close(); err != nil {
		t.Errorf("could not close Service.Session: %v", err.Error())
	}

}

/*
	TestService_CreateAliasedCollection tests the following internal functions/methods via nested calls:

		(all calls in TestNewService)
		Service.CreateAliasedCollection
			NewCollection
				Collection.Modified
		Collection.Delete
		Service.SetAlias

	(By extension, Service.CreateCollection is also tested as it's a very thin wrapper
	around Service.CreateAliasedCollection).
*/
func TestService_CreateAliasedCollection(t *testing.T) {

	var err error
	var svc *Service
	var collection *Collection

	if svc, err = NewService(); err != nil {
		t.Fatalf("could not get new Service via NewService: %v", err.Error())
	}

	if collection, err = svc.CreateAliasedCollection(collectionName.String(), collectionAlias.String()); err != nil {
		t.Errorf(
			"error when creating aliased collection '%v' with alias '%v': %v",
			collectionName.String(), collectionAlias.String(), err.Error(),
		)
	} else {
		if err = collection.Delete(); err != nil {
			t.Errorf(
				"error when deleting aliased collection '%v' with alias '%v': %v",
				collectionName.String(), collectionAlias.String(), err.Error(),
			)
		}
	}

	if err = svc.SetAlias(testAlias, collection.Dbus.Path()); err != nil {
		t.Errorf(
			"error when setting an alias '%v' for aliased collection '%v' (original alias '%v')",
			testAlias, collectionName.String(), collectionAlias.String(),
		)
	}

	if err = svc.Close(); err != nil {
		t.Errorf("could not close Service.Session: %v", err.Error())
	}
}

/*
	TestService_GetCollection tests the following internal functions/methods via nested calls:

		(all calls in TestService_CreateAliasedCollection)
		Service.GetCollection
			(all calls in TestService_Collections)
			Service.ReadAlias

*/
func TestService_GetCollection(t *testing.T) {

	var err error
	var svc *Service
	var coll *Collection

	if svc, err = NewService(); err != nil {
		t.Fatalf("could not get new Service via NewService: %v", err.Error())
	}

	if coll, err = svc.GetCollection(defaultCollection); err != nil {
		t.Errorf("failed to get collection '%v' via Service.GetCollection: %v", defaultCollection, err.Error())
	} else {
		t.Logf("got collection '%v' via reference '%v'", coll.name, defaultCollection)
	}

	if err = svc.Close(); err != nil {
		t.Errorf("could not close Service.Session: %v", err.Error())
	}
}

/*
	TestService_Secrets tests the following internal functions/methods via nested calls:

		(all calls in TestNewService)
		(all calls in TestService_CreateAliasedCollection)
		Service.CreateCollection
		Service.SearchItems
		Service.GetSecrets

*/
/* TODO: left off on this.
func TestService_Secrets(t *testing.T) {

	var err error
	var svc *Service
	var collection *Collection
	var itemPaths []dbus.ObjectPath

	if svc, err = NewService(); err != nil {
		t.Fatalf("could not get new Service via NewService: %v", err.Error())
	}

	if collection, err = svc.CreateCollection(collectionName.String()); err != nil {
		t.Errorf("could not create collection '%v': %v", collectionName.String(), err.Error())
	}

	if err = svc.Close(); err != nil {
		t.Errorf("could not close Service.Session: %v", err.Error())
	}
}
*/

// service.Lock & service.Unlock
