package gosecret

import (
	"testing"
	`time`

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
		Collection.Created
*/
func TestService_Collections(t *testing.T) {

	var err error
	var svc *Service
	var colls []*Collection
	var collLabel string
	var created time.Time
	var modified time.Time

	if svc, err = NewService(); err != nil {
		t.Fatalf("could not get new Service via NewService: %v", err.Error())
	}

	if colls, err = svc.Collections(); err != nil {
		t.Errorf("could not get Service.Collections: %v", err.Error())
	} else {
		t.Logf("found %v collections via Service.Collections", len(colls))

		for idx, c := range colls {
			if collLabel, err = c.Label(); err != nil {
				t.Errorf(
					"failed to get label for collection '%v': %v",
					string(c.Dbus.Path()), err.Error(),
				)
			}
			if created, err = c.Created(); err != nil {
				t.Errorf(
					"failed to get created time for collection '%v': %v",
					string(c.Dbus.Path()), err.Error(),
				)
			}
			if modified, _, err = c.Modified(); err != nil {
				t.Errorf(
					"failed to get modified time for collection '%v': %v",
					string(c.Dbus.Path()), err.Error(),
				)
			}
			t.Logf(
				"collection #%v (name '%v', label '%v'): created %v, last modified %v",
				idx, c.path(), collLabel, created, modified,
			)
		}
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
/* DISABLED. Currently (as of 0.20.4), *only* the alias "default" is allowed. TODO: revisit in future?
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
		(all calls in TestService_Collections)


			NewErrors
		Collection.Created
		Service.GetCollection
			NewCollection
			Collection.Modified
		Service.ReadAlias

	The default collection (login) is fetched instead of creating one as this collection should exist,
	and thus this function tests fetching existing collections instead of newly-created ones.
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
		t.Logf("got collection '%v' via reference '%v'", coll.LabelName, defaultCollection)
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
		Item.Delete

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
	var created time.Time
	var modified time.Time
	var newModified time.Time
	var wasModified bool
	var isModified bool

	if svc, err = NewService(); err != nil {
		t.Fatalf("could not get new Service via NewService: %v", err.Error())
	}

	if collection, err = svc.CreateCollection(collectionName.String()); err != nil {
		t.Errorf("could not create collection '%v': %v", collectionName.String(), err.Error())
		if err = svc.Close(); err != nil {
			t.Fatalf("could not close Service.Session: %v", err.Error())
		}
		return
	} else {
		t.Logf("created collection '%v' at path '%v' successfully", collectionName.String(), string(collection.Dbus.Path()))
	}

	if created, err = collection.Created(); err != nil {
		t.Errorf(
			"failed to get created time for '%v': %v",
			collectionName.String(), err.Error(),
		)
	}
	if modified, wasModified, err = collection.Modified(); err != nil {
		t.Errorf(
			"failed to get modified time for '%v': %v",
			collectionName.String(), err.Error(),
		)
	}
	t.Logf(
		"%v: collection '%v': created at %v, last modified at %v; was changed: %v",
		time.Now(), collectionName.String(), created, modified, wasModified,
	)

	// Create a secret
	testSecret = NewSecret(svc.Session, nil, []byte(testSecretContent), "text/plain")
	if testItem, err = collection.CreateItem(testItemLabel, itemAttrs, testSecret, true); err != nil {
		t.Errorf("could not create Item in collection '%v': %v", collectionName.String(), err.Error())
		if err = svc.Close(); err != nil {
			t.Fatalf("could not close Service.Session: %v", err.Error())
		}
		return
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
			t.Errorf(
				"number of unlocked items in collection '%v' (%v) is not equal to 1; items dump pending...",
				collectionName.String(), len(itemResultsUnlocked),
			)
			for idx, i := range itemResultsUnlocked {
				t.Logf("ITEM #%v IN COLLECTION %v: %v ('%v')", idx, collectionName.String(), i.LabelName, string(i.Dbus.Path()))
			}
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

	// Confirm the modification information changed.
	_ = newModified
	_ = isModified
	/* TODO: Disabled for now; it *seems* the collection modification time doesn't auto-update? See collection.setModify if not.
	if newModified, isModified, err = collection.Modified(); err != nil {
		t.Errorf(
			"failed to get modified time for '%v': %v",
			collectionName.String(), err.Error(),
		)
	}
	t.Logf(
		"%v: (post-change) collection '%v': last modified at %v; was changed: %v",
		time.Now(), collectionName.String(), newModified, isModified,
	)
	if !isModified {
		t.Errorf(
			"modification tracking for collection '%v' failed; expected true but got false",
			collectionName.String(),
		)
	} else {
		t.Logf("(modification check passed)")
	}
	if !newModified.After(modified) {
		t.Errorf(
			"modification timestamp update for '%v' failed: old %v, new %v",
			collectionName.String(), modified, newModified,
		)
	}
	*/

	/*
		Delete the item and collection to clean up.
		*Technically* the item is deleted if the collection is, but its path is held in memory if not
		manually removed, leading to a "ghost" item.
	*/
	if err = testItem.Delete(); err != nil {
		t.Errorf("could not delete test item '%v' in collection '%v': %v",
			string(testItem.Dbus.Path()), collectionName.String(), err.Error(),
		)
	}
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
		t.Errorf("could not create collection '%v': %v", collectionName.String(), err.Error())
		if err = svc.Close(); err != nil {
			t.Fatalf("could not close Service.Session: %v", err.Error())
		}
		return
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
		if err = collection.Unlock(); err != nil {
			t.Errorf("could not unlock collection '%v': %v", collectionName.String(), err.Error())
		}
		if stateChangeLock, err = collection.Locked(); err != nil {
			t.Errorf("received error when checking collection '%v' lock status: %v", collectionName.String(), err.Error())
		}
		if err = collection.Lock(); err != nil {
			t.Errorf("could not lock collection '%v': %v", collectionName.String(), err.Error())
		}
	} else {
		if err = collection.Lock(); err != nil {
			t.Errorf("could not lock collection '%v': %v", collectionName.String(), err.Error())
		}
		if stateChangeLock, err = collection.Locked(); err != nil {
			t.Errorf("received error when checking collection '%v' lock status: %v", collectionName.String(), err.Error())
		}
		if err = collection.Unlock(); err != nil {
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
