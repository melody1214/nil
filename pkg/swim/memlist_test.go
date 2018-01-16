package swim

import (
	"testing"
)

func TestMemlist(t *testing.T) {
	meml := newMemList()

	testCases := []*Member{
		&Member{ID: "1"},
		&Member{ID: "2"},
		&Member{ID: "3"},
	}

	for i, c := range testCases {
		if v := meml.get(c.ID); v != nil {
			t.Errorf("expected value of %s to be nil: got %+v", c.ID, v)
		}

		meml.set(c)
		if v := meml.get(c.ID); v == c {
			t.Errorf("expected copied object of %s: got same object", c.ID)
		}

		if v := meml.get(c.ID); *v != *c {
			t.Errorf("expected value of %s to be %+v: got %+v", c.ID, c, v)
		}

		if v := meml.fetch(i + 1); len(v) != i+1 {
			t.Errorf("expected fetch %d number of members: got %d members", i+1, len(v))
		}
	}
}
