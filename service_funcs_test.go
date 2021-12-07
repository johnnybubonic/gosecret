package gosecret

import (
	`testing`
)

func TestNewService(t *testing.T) {

	var err error

	if _, err = NewService(); err != nil {
		t.Fatalf("could not get new Service via NewService: %v", err.Error())
	}

}
