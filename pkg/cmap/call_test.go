package cmap

import (
	"reflect"
	"testing"
)

func TestSearchCall(t *testing.T) {
	testNodes := []Node{
		{
			ID:   ID(0),
			Name: "node0",
			Addr: "localhost:1000",
			Type: MDS,
			Stat: Alive,
		},
		{
			ID:   ID(1),
			Name: "node1",
			Addr: "localhost:2000",
			Type: DS,
			Stat: Alive,
		},
		{
			ID:   ID(2),
			Name: "node2",
			Addr: "localhost:3000",
			Type: DS,
			Stat: Faulty,
		},
		{
			ID:   ID(3),
			Name: "node3",
			Addr: "localhost:4000",
			Type: GW,
			Stat: Alive,
		},
	}

	testMap := CMap{
		Version: 1,
		Nodes:   testNodes,
	}

	ct, err := NewController("")
	if err != nil {
		t.Fatal(err)
	}
	ct.Update(&testMap)

	for _, n := range testMap.Nodes {
		// 1. Search with the all conditions are matched.
		c := ct.SearchCall()
		c.ID(n.ID).Name(n.Name).Type(n.Type).Status(n.Stat)

		if find, err := c.Do(); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(find, n) {
			t.Errorf("expected %+v, got %+v", n, find)
		}

		// 2. Search with only condition for id.
		c = ct.SearchCall()
		c.ID(n.ID)

		if find, err := c.Do(); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(find, n) {
			t.Errorf("expected %+v, got %+v", n, find)
		}

		// 3. Search with wrong contition.
		c = ct.SearchCall()
		c.Name(n.Name + "wrong name")

		if _, err := c.Do(); err == nil {
			t.Error("expected error, got nil")
		}

		// 4. Search with type and status.
		c = ct.SearchCall()
		c.Type(n.Type).Status(n.Stat)

		if find, err := c.Do(); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(find, n) {
			t.Errorf("expected %+v, got %+v", n, find)
		}
	}
}
