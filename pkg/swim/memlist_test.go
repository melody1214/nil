package swim

import (
	"testing"
)

func TestMemlist(t *testing.T) {
	meml := newMemList("1")

	testCases := []Member{
		Member{ID: "1"},
		Member{ID: "2"},
		Member{ID: "3"},
	}

	for i, c := range testCases {
		if v, ok := meml.get(c.ID); ok {
			t.Errorf("expected not found but: got %+v", v)
		}

		meml.set(c)
		if v, _ := meml.get(c.ID); v != c {
			t.Errorf("expected %+v but got %+v", c, v)
		}

		if v := meml.fetch(i + 1); len(v) != i+1 {
			t.Errorf("expected fetch %d number of members: got %d members", i+1, len(v))
		}
	}
}

func TestFetchOptions(t *testing.T) {
	meml := newMemList("me")

	mems := []Member{
		Member{ID: "me", Status: Alive},
		Member{ID: "steve", Status: Suspect},
		Member{ID: "john", Status: Suspect},
	}

	for _, m := range mems {
		meml.set(m)
	}

	testCases := []struct {
		option   func() fetchOption
		expected int
	}{
		{
			withNotAlive,
			2,
		},
		{
			withNotMyself,
			2,
		},
		{
			withNotSuspect,
			1,
		},
	}

	for _, c := range testCases {
		if v := len(meml.fetch(0, c.option())); v != c.expected {
			t.Errorf("expected fetch %d number of members: got %d members", c.expected, v)
		}
	}
}
