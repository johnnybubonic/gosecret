package gosecret

import (
	`testing`

	`github.com/godbus/dbus/v5`
)

// Some functions are covered in the Service tests.

/*
	TestNewCollection tests the following internal functions/methods via nested calls:

		(all calls in TestNewService)
		NewService
		NewCollection

*/
func TestNewCollection(t *testing.T) {

	var svc *Service
	var collection *Collection
	var err error

	if svc, err = NewService(); err != nil {
		t.Fatalf("NewService failed: %v", err.Error())
	}

	if collection, err = NewCollection(svc, dbus.ObjectPath(dbusDefaultCollectionPath)); err != nil {
		t.Errorf(
			"TestNewCollection failed when fetching collection at '%v': %v",
			dbusDefaultCollectionPath, err.Error(),
		)
	}
	_ = collection

	if err = svc.Close(); err != nil {
		t.Errorf("could not close Service.Session: %v", err.Error())
	}
}

/*
	TestCollection_Items tests the following internal functions/methods via nested calls:

		(all calls in TestNewCollection)
		Service.GetCollection
		Collection.Items
		NewSecret
		Collection.CreateItem
		Collection.SearchItems
		Item.Delete

*/
func TestCollection_Items(t *testing.T) {

	var svc *Service
	var collection *Collection
	var items []*Item
	var item *Item
	var searchItemResults []*Item
	var secret *Secret
	var err error

	if svc, err = NewService(); err != nil {
		t.Fatalf("NewService failed: %v", err.Error())
	}

	if collection, err = svc.GetCollection(defaultCollection); err != nil {
		t.Errorf("failed when fetching collection '%v': %v",
			defaultCollection, err.Error(),
		)
		if err = svc.Close(); err != nil {
			t.Fatalf("could not close Service.Session: %v", err.Error())
		}
		return
	}

	if items, err = collection.Items(); err != nil {
		t.Errorf(
			"failed fetching items for '%v' at '%v': %v",
			defaultCollection, string(collection.Dbus.Path()), err.Error(),
		)
	} else {
		t.Logf("found %v items in collection '%v' at '%v'", len(items), defaultCollection, string(collection.Dbus.Path()))
	}

	/* This is almost always going to trigger the warning. See Item.idx for details why.
	var label string
	for idx, i := range items {
		if label, err = i.Label(); err != nil {
			t.Errorf("failed to get label of item '%v' in collection '%v': %v", string(i.Dbus.Path()), collectionName.String(), err.Error())
			continue
		}
		if i.idx != idx {
			t.Logf(
				"WARN: item '%v' ('%v') in collection '%v' internal IDX ('%v') does NOT match native slice IDX ('%v')",
				string(i.Dbus.Path()), label, collectionName.String(), i.idx, idx,
			)
		}
	}
	*/

	secret = NewSecret(svc.Session, []byte{}, []byte(testSecretContent), "text/plain")

	if item, err = collection.CreateItem(testItemLabel, itemAttrs, secret, false); err != nil {
		t.Errorf(
			"could not create item '%v' in collection '%v': %v",
			testItemLabel, defaultCollection, err.Error(),
		)
	} else {

		if searchItemResults, err = collection.SearchItems(testItemLabel); err != nil {
			t.Errorf("failed to find item '%v' via Collection.SearchItems: %v", string(item.Dbus.Path()), err.Error())
		} else if len(searchItemResults) == 0 {
			t.Errorf("failed to find item '%v' via Collection.SearchItems, returned 0 results (should be at least 1)", testItemLabel)
		} else {
			t.Logf("found %v results for Collection.SearchItems", len(searchItemResults))
		}

		if err = item.Delete(); err != nil {
			t.Errorf("failed to delete created item '%v': %v", string(item.Dbus.Path()), err.Error())
		}

	}

	if err = svc.Close(); err != nil {
		t.Errorf("could not close Service.Session: %v", err.Error())
	}
}

/*
	TestCollection_Label tests the following internal functions/methods via nested calls:

		(all calls in TestNewCollection)
		Service.GetCollection
		Collection.Label
		Collection.PathName

*/
func TestCollection_Label(t *testing.T) {

	var svc *Service
	var collection *Collection
	var collLabel string
	var err error

	if svc, err = NewService(); err != nil {
		t.Fatalf("NewService failed: %v", err.Error())
	}

	if collection, err = svc.GetCollection(defaultCollectionLabel); err != nil {
		t.Errorf(
			"failed when fetching collection '%v': %v",
			defaultCollectionLabel, err.Error(),
		)
		if err = svc.Close(); err != nil {
			t.Fatalf("could not close Service.Session: %v", err.Error())
		}
		return
	}

	if collLabel, err = collection.Label(); err != nil {
		t.Errorf("cannot fetch label for '%v': %v", string(collection.Dbus.Path()), err.Error())
		if err = svc.Close(); err != nil {
			t.Fatalf("could not close Service.Session: %v", err.Error())
		}
		return
	}

	if defaultCollectionLabel != collLabel {
		t.Errorf("fetched collection ('%v') does not match fetched collection label ('%v')", collLabel, defaultCollectionLabel)
	}

	if err = svc.Close(); err != nil {
		t.Errorf("could not close Service.Session: %v", err.Error())
	}

}

/*
	TestCollection_Locked tests the following internal functions/methods via nested calls:

		(all calls in TestNewCollection)
		Collection.Locked

*/
func TestCollection_Locked(t *testing.T) {

	var svc *Service
	var collection *Collection
	var isLocked bool
	var err error

	if svc, err = NewService(); err != nil {
		t.Fatalf("NewService failed: %v", err.Error())
	}

	if collection, err = svc.GetCollection(defaultCollection); err != nil {
		t.Errorf(
			"failed when fetching collection '%v': %v",
			defaultCollectionLabel, err.Error(),
		)
		if err = svc.Close(); err != nil {
			t.Fatalf("could not close Service.Session: %v", err.Error())
		}
		return
	}

	if isLocked, err = collection.Locked(); err != nil {
		t.Errorf("failed to get lock status for collection '%v': %v", collection.PathName(), err.Error())
	} else {
		t.Logf("collection '%v' lock status: %v", collection.PathName(), isLocked)
	}

	if err = svc.Close(); err != nil {
		t.Errorf("could not close Service.Session: %v", err.Error())
	}
}

/*
	TestCollection_Relabel tests the following internal functions/methods via nested calls:

		(all calls in TestNewCollection)
		Service.CreateCollection
		Collection.Relabel

*/
func TestCollection_Relabel(t *testing.T) {

	var svc *Service
	var collection *Collection
	var collLabel string
	var newCollLabel string = collectionAlias.String()
	var err error

	if svc, err = NewService(); err != nil {
		t.Fatalf("NewService failed: %v", err.Error())
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

	if collLabel, err = collection.Label(); err != nil {
		t.Errorf("could not fetch label for collection '%v': %v", string(collection.Dbus.Path()), err.Error())
		if err = svc.Close(); err != nil {
			t.Fatalf("could not close Service.Session: %v", err.Error())
		}
		return
	}

	if err = collection.Relabel(newCollLabel); err != nil {
		t.Errorf("failed to relabel collection '%v' to '%v': %v", collLabel, newCollLabel, err.Error())
	} else {
		t.Logf("relabeled collection '%v' to '%v'", collLabel, newCollLabel)
	}

	if collLabel, err = collection.Label(); err != nil {
		t.Errorf("could not fetch label for collection '%v': %v", string(collection.Dbus.Path()), err.Error())
		if err = collection.Delete(); err != nil {
			t.Errorf("failed to delete collection '%v': %v", string(collection.Dbus.Path()), err.Error())
		}
		if err = svc.Close(); err != nil {
			t.Fatalf("could not close Service.Session: %v", err.Error())
		}
		return
	} else {
		if collLabel != newCollLabel {
			t.Errorf("collection did not relabel; new label '%v', actual label '%v'", newCollLabel, collLabel)
		}
	}

	if err = collection.Delete(); err != nil {
		t.Errorf("failed to delete collection '%v': %v", string(collection.Dbus.Path()), err.Error())
	}

	if err = svc.Close(); err != nil {
		t.Errorf("could not close Service.Session: %v", err.Error())
	}
}
