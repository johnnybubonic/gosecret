package gosecret

import (
	"testing"

	`github.com/godbus/dbus/v5`
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

	if colls, err = svc.Collections(); err != nil {
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
/* DISABLED. Currently, *only* the alias "default" is allowed. TODO: revisit in future?
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
		if err = svc.SetAlias(testAlias, collection.Dbus.Path()); err != nil {
			t.Errorf(
				"error when setting an alias '%v' for aliased collection '%v' (original alias '%v')",
				testAlias, collectionName.String(), collectionAlias.String(),
			)
		}
		if err = collection.Delete(); err != nil {
			t.Errorf(
				"error when deleting aliased collection '%v' with alias '%v': %v",
				collectionName.String(), collectionAlias.String(), err.Error(),
			)
		}
		t.Logf("created collection '%v' at path '%v' successfully", collectionName.String(), string(collection.Dbus.Path()))
	}

	if err = svc.Close(); err != nil {
		t.Errorf("could not close Service.Session: %v", err.Error())
	}
}
*/

/*
	TestService_GetCollection tests the following internal functions/methods via nested calls:

		(all calls in TestService_CreateAliasedCollection)
		Service.GetCollection
			(all calls in TestService_Collections)
			Service.ReadAlias

	The default collection (login) is fetched instead of creating one as this collection should exist,
	and tests fetching existing collections instead of newly-created ones.
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
			NewItem
			NewErrors
		Service.GetSecrets
		NewSecret
		Collection.CreateItem
		Item.Label

*/
func TestService_Secrets(t *testing.T) {

	var err error
	var svc *Service
	var collection *Collection
	var itemResultsUnlocked []*Item
	var itemResultsLocked []*Item
	var itemName string
	var resultItemName string
	var testItem *Item
	var testSecret *Secret
	var secretsResult map[dbus.ObjectPath]*Secret
	var itemPaths []dbus.ObjectPath
	var itemAttrs map[string]string = map[string]string{
		"GOSECRET": "yes",
	}

	if svc, err = NewService(); err != nil {
		t.Fatalf("could not get new Service via NewService: %v", err.Error())
	}

	if collection, err = svc.CreateCollection(collectionName.String()); err != nil {
		if err = svc.Close(); err != nil {
			t.Errorf("could not close Service.Session: %v", err.Error())
		}
		t.Fatalf("could not create collection '%v': %v", collectionName.String(), err.Error())
	} else {
		t.Logf("created collection '%v' at path '%v' successfully", collectionName.String(), string(collection.Dbus.Path()))
	}

	// Create a secret
	testSecret = NewSecret(svc.Session, nil, []byte(testSecretContent), "text/plain")
	if testItem, err = collection.CreateItem(testItemLabel, itemAttrs, testSecret, true); err != nil {
		if err = svc.Close(); err != nil {
			t.Errorf("could not close Service.Session: %v", err.Error())
		}
		t.Fatalf("could not create Item in collection '%v': %v", collectionName.String(), err.Error())
	}

	if itemName, err = testItem.Label(); err != nil {
		t.Errorf(
			"could not get label for newly-created item '%v' in collection '%v': %v",
			string(testItem.Dbus.Path()), collectionName.String(), err.Error(),
		)
		itemName = testItemLabel
	}

	// Search items
	if itemResultsUnlocked, itemResultsLocked, err = svc.SearchItems(itemAttrs); err != nil {
		t.Errorf("did not find Item '%v' in collection '%v' SearchItems: %v", itemName, collectionName.String(), err.Error())
	} else {
		if len(itemResultsLocked) != 0 && itemResultsUnlocked != nil {
			t.Errorf("at least one locked item in collection '%v'", collectionName.String())
		}
		if len(itemResultsUnlocked) != 1 {
			t.Errorf("number of unlocked items in collection '%v' is not equal to 1", collectionName.String())
		}
		if resultItemName, err = itemResultsUnlocked[0].Label(); err != nil {
			t.Errorf("cannot fetch test Item name from collection '%v' in SearchItems: %v", collectionName.String(), err.Error())
		} else {
			if resultItemName != itemName {
				t.Errorf("seem to have fetched an improper Item from collection '%v' in SearchItems", collectionName.String())
			}
		}
	}

	// Fetch secrets
	itemPaths = make([]dbus.ObjectPath, len(itemResultsUnlocked))
	if len(itemResultsUnlocked) >= 1 {
		itemPaths[0] = itemResultsUnlocked[0].Dbus.Path()
		if secretsResult, err = svc.GetSecrets(itemPaths...); err != nil {
			t.Errorf("failed to fetch Item path '%v' via Service.GetSecrets: %v", string(itemPaths[0]), err.Error())
		} else if len(secretsResult) != 1 {
			t.Errorf("received %v secrets from Service.GetSecrets instead of 1", len(secretsResult))
		}
	}

	// Delete the collection to clean up.
	if err = collection.Delete(); err != nil {
		t.Errorf(
			"error when deleting collection '%v' when testing Service: %v",
			collectionName.String(), err.Error(),
		)
	}

	if err = svc.Close(); err != nil {
		t.Errorf("could not close Service.Session: %v", err.Error())
	}
}

// service.Lock & service.Unlock

/*
	TestService_Locking tests the following internal functions/methods via nested calls:

		(all calls in TestNewService)
		(all calls in TestService_CreateAliasedCollection)
		Service.Lock
		Service.Unlock
		Collection.Locked

*/
func TestService_Locking(t *testing.T) {

	var err error
	var isLocked bool
	var stateChangeLock bool
	var svc *Service
	var collection *Collection

	if svc, err = NewService(); err != nil {
		t.Fatalf("could not get new Service via NewService: %v", err.Error())
	}

	if collection, err = svc.CreateCollection(collectionName.String()); err != nil {
		if err = svc.Close(); err != nil {
			t.Errorf("could not close Service.Session: %v", err.Error())
		}
		t.Errorf("could not create collection '%v': %v", collectionName.String(), err.Error())
	} else {
		t.Logf("created collection '%v' at path '%v' successfully", collectionName.String(), string(collection.Dbus.Path()))
	}

	if isLocked, err = collection.Locked(); err != nil {
		t.Errorf("received error when checking collection '%v' lock status: %v", collectionName.String(), err.Error())
		if err = collection.Delete(); err != nil {
			t.Errorf(
				"error when deleting collection '%v' when testing Service: %v",
				collectionName.String(), err.Error(),
			)
		}
		if err = svc.Close(); err != nil {
			t.Errorf("could not close Service.Session: %v", err.Error())
		}
		return
	} else {
		t.Logf("collection '%v' original lock status: %v", collectionName.String(), isLocked)
	}

	// Change the state.
	if isLocked {
		if err = svc.Unlock(collection.Dbus.Path()); err != nil {
			t.Errorf("could not unlock collection '%v': %v", collectionName.String(), err.Error())
		}
		if stateChangeLock, err = collection.Locked(); err != nil {
			t.Errorf("received error when checking collection '%v' lock status: %v", collectionName.String(), err.Error())
		}
		if err = svc.Lock(collection.Dbus.Path()); err != nil {
			t.Errorf("could not lock collection '%v': %v", collectionName.String(), err.Error())
		}
	} else {
		if err = svc.Lock(collection.Dbus.Path()); err != nil {
			t.Errorf("could not lock collection '%v': %v", collectionName.String(), err.Error())
		}
		if stateChangeLock, err = collection.Locked(); err != nil {
			t.Errorf("received error when checking collection '%v' lock status: %v", collectionName.String(), err.Error())
		}
		if err = svc.Unlock(collection.Dbus.Path()); err != nil {
			t.Errorf("could not unlock collection '%v': %v", collectionName.String(), err.Error())
		}
	}

	if stateChangeLock != !isLocked {
		t.Errorf(
			"flipped lock state for collection '%v' (locked: %v) is not opposite of original lock state (locked: %v)",
			collectionName.String(), stateChangeLock, isLocked,
		)
	}

	// Delete the collection to clean up.
	if err = collection.Delete(); err != nil {
		t.Errorf(
			"error when deleting collection '%v' when testing Service: %v",
			collectionName.String(), err.Error(),
		)
	}

	if err = svc.Close(); err != nil {
		t.Errorf("could not close Service.Session: %v", err.Error())
	}
}
