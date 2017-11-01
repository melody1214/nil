package db

import (
	"testing"
)

func TestDB(t *testing.T) {
	db := New()

	testCases := []struct {
		key   string
		value interface{}
	}{
		{"hello", "world"},
		{"today", 20171101},
	}

	for _, c := range testCases {
		if v := db.Get(c.key); v != nil {
			t.Errorf("expected key value %s to be nil: got %+v", c.key, v)
		}

		db.Put(c.key, c.value)
		if v := db.Get(c.key); v != c.value {
			t.Errorf("expected key value %s to be %+v: got %+v", c.key, c.value, v)
		}

		db.Delete(c.key)
		if v := db.Get(c.key); v != nil {
			t.Errorf("expected key value %s to be nil: got %+v", c.key, v)
		}
	}
}
