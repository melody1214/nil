package swim

import (
	"testing"
)

func TestMemlist(t *testing.T) {
	meml := newMemList()

	testCases := []*Member{
		&Member{UUID: "1"},
		&Member{UUID: "2"},
		&Member{UUID: "3"},
	}

	for i, c := range testCases {
		if v := meml.get(c.UUID); v != nil {
			t.Errorf("expected value of %s to be nil: got %+v", c.UUID, v)
		}

		meml.set(c)
		if v := meml.get(c.UUID); v == c {
			t.Errorf("expected copied object of %s: got same object", c.UUID)
		}

		if v := meml.get(c.UUID); *v != *c {
			t.Errorf("expected value of %s to be %+v: got %+v", c.UUID, c, v)
		}

		if v := meml.fetch(i + 1); len(v) != i+1 {
			t.Errorf("expected fetch %d number of members: got %d members", i+1, len(v))
		}
	}
}
