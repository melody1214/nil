package swim

import (
	"testing"

	"github.com/chanyoung/nil/pkg/swim/swimpb"
)

func TestMemlist(t *testing.T) {
	meml := newMemList()

	testCases := []*swimpb.Member{
		&swimpb.Member{Uuid: "1"},
		&swimpb.Member{Uuid: "2"},
		&swimpb.Member{Uuid: "3"},
	}

	for i, c := range testCases {
		if v := meml.get(c.Uuid); v != nil {
			t.Errorf("expected value of %s to be nil: got %+v", c.Uuid, v)
		}

		meml.set(c)
		if v := meml.get(c.Uuid); v == c {
			t.Errorf("expected copied object of %s: got same object", c.Uuid)
		}

		if v := meml.get(c.Uuid); *v != *c {
			t.Errorf("expected value of %s to be %+v: got %+v", c.Uuid, c, v)
		}

		if v := meml.fetch(i + 1); len(v) != i+1 {
			t.Errorf("expected fetch %d number of members: got %d members", i+1, len(v))
		}
	}
}
