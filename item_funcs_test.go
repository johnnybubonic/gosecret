package gosecret

import (
	`reflect`
	`testing`
)

// Some functions are covered in the Service tests and Collection tests.

/*
	TestItem tests all remaining Item funcs (see Service and Collection funcs for the other tests.

*/
func TestItem(t *testing.T) {

	var svc *Service
	var collection *Collection
	var item *Item
	var secret *Secret
	var newItemLabel string
	var testLabel string
	var attrs map[string]string
	var modAttrs map[string]string
	var newAttrs map[string]string
	var newAttrsGnome map[string]string
	var typeString string
	var err error

	// Setup.
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

	// Create an Item/Secret.
	secret = NewSecret(svc.Session, []byte{}, []byte(testSecretContent), "text/plain")

	if item, err = collection.CreateItem(testItemLabel, itemAttrs, secret, true); err != nil {
		t.Errorf("could not create item %v in collection '%v': %v", testItemLabel, collectionName.String(), err.Error())
		if err = collection.Delete(); err != nil {
			t.Errorf("could not delete collection '%v': %v", collectionName.String(), err.Error())
		}
		if err = svc.Close(); err != nil {
			t.Fatalf("could not close Service.Session: %v", err.Error())
		}
		return
	}

	// Fetch attributes
	if attrs, err = item.Attributes(); err != nil {
		t.Errorf("failed to fetch attributes for item %v in collection '%v': %v", testItemLabel, collectionName.String(), err.Error())
	} else {
		t.Logf(
			"Fetch result; original attributes: %#v, fetched attributes: %#v for item '%v' in '%v'",
			itemAttrs, attrs, testItemLabel, collectionName.String(),
		)
	}

	newAttrs = map[string]string{
		"foo": "bar",
		"bar": "baz",
		"baz": "quux",
	}

	// Replace attributes.
	if err = item.ReplaceAttributes(newAttrs); err != nil {
		t.Errorf("could not replace attributes for item '%v' in collection '%v': %v", testItemLabel, collectionName.String(), err.Error())
	} else {
		// Modify attributes.
		// "flat" modification.
		modAttrs = map[string]string{
			"foo": "quux",
		}
		if err = item.ModifyAttributes(modAttrs); err != nil {
			t.Errorf(
				"could not modify attributes for item '%v' in collection '%v' (%#v => %#v): %v",
				testItemLabel, collectionName.String(), newAttrs, modAttrs, err.Error(),
			)
		}
		// "delete" modification.
		newAttrs = map[string]string{
			"foo": "quux",
			"bar": "baz",
		}
		newAttrsGnome = make(map[string]string, 0)
		for k, v := range newAttrs {
			newAttrsGnome[k] = v
		}
		// Added via SecretService automatically? Seahorse? It appears sometimes and others it does not. Cause is unknown.
		newAttrsGnome["xdg:schema"] = DbusDefaultItemType
		modAttrs = map[string]string{
			"baz": ExplicitAttrEmptyValue,
		}

		if err = item.ModifyAttributes(modAttrs); err != nil {
			t.Errorf(
				"could not modify (with deletion) attributes for item '%v' in collection '%v' (%#v => %#v): %v",
				testItemLabel, collectionName.String(), newAttrs, modAttrs, err.Error(),
			)
		} else {
			if attrs, err = item.Attributes(); err != nil {
				t.Errorf("failed to fetch attributes for item %v in collection '%v': %v", testItemLabel, collectionName.String(), err.Error())
			}
			if !reflect.DeepEqual(attrs, newAttrs) && !reflect.DeepEqual(attrs, newAttrsGnome) {
				t.Errorf("newly-modified attributes (%#v) do not match expected attributes (%#v)", attrs, newAttrs)
			} else {
				t.Logf("modified attributes (%#v) match expected attributes (%#v)", attrs, newAttrs)
			}
		}
	}

	// Item.Relabel
	newItemLabel = testItemLabel + "_RELABELED"
	if err = item.Relabel(newItemLabel); err != nil {
		t.Errorf("failed to relabel item '%v' to '%v': %v", testItemLabel, newItemLabel, err.Error())
	}
	if testLabel, err = item.Label(); err != nil {
		t.Errorf("failed to fetch label for '%v': %v", string(item.Dbus.Path()), err.Error())
	}
	if newItemLabel != testLabel {
		t.Errorf("new item label post-relabeling ('%v') does not match explicitly set label ('%v')", testLabel, newItemLabel)
	}

	// And Item.Type.
	if typeString, err = item.Type(); err != nil {
		t.Errorf("failed to get Item.Type for '%v': %v", string(item.Dbus.Path()), err.Error())
	} else {
		if typeString != DbusDefaultItemType {
			t.Errorf("Item.Type mismatch for '%v': '%v' (should be '%v')", string(item.Dbus.Path()), typeString, DbusDefaultItemType)
		}
		t.Logf("item type for '%v': %v", string(item.Dbus.Path()), typeString)
	}

	// Teardown.
	if err = item.Delete(); err != nil {
		t.Errorf("failed to delete item '%v': %v", string(item.Dbus.Path()), err.Error())
	}
	if err = collection.Delete(); err != nil {
		t.Errorf("failed to delete collection '%v': %v", collectionName.String(), err.Error())
	}
	if err = svc.Close(); err != nil {
		t.Errorf("could not close Service.Session: %v", err.Error())
	}
}
