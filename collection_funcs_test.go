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
	TestCollection_Label tests the following internal functions/methods via nested calls:
		(all calls in TestNewCollection)
		Service.GetCollection
		Collection.Label

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
		err = nil
		if err = svc.Close(); err != nil {
			t.Errorf("could not close Service.Session: %v", err.Error())
		}
	}

	if collLabel, err = collection.Label(); err != nil {
		t.Errorf("cannot fetch label for '%v': %v", string(collection.Dbus.Path()), err.Error())
		if err = svc.Close(); err != nil {
			t.Fatalf("could not close Service.Session: %v", err.Error())
		}
	}

	if defaultCollection != collLabel {
		t.Errorf("fetched collection ('%v') does not match fetched collection label ('%v')", collLabel, defaultCollection)
	}

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
		if err = svc.Close(); err != nil {
			t.Errorf("could not close Service.Session: %v", err.Error())
		}
		t.Fatalf("failed when fetching collection '%v': %v",
			defaultCollection, err.Error(),
		)
	}

	if items, err = collection.Items(); err != nil {
		t.Errorf(
			"failed fetching items for '%v' at '%v': %v",
			defaultCollection, string(collection.Dbus.Path()), err.Error(),
		)
	} else {
		t.Logf("found %v items in collection '%v' at '%v'", len(items), defaultCollection, string(collection.Dbus.Path()))
	}

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
