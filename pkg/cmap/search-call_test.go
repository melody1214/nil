package cmap

import (
	"os"
	"reflect"
	"testing"
)

func TestSearchCallNode(t *testing.T) {
	testNodes := []Node{
		{
			ID:   ID(0),
			Name: NodeName("node0"),
			Addr: NodeAddress("localhost:1000"),
			Type: MDS,
			Stat: Alive,
		},
		{
			ID:   ID(1),
			Name: NodeName("node1"),
			Addr: NodeAddress("localhost:2000"),
			Type: DS,
			Stat: Alive,
		},
		{
			ID:   ID(2),
			Name: NodeName("node2"),
			Addr: NodeAddress("localhost:3000"),
			Type: DS,
			Stat: Faulty,
		},
		{
			ID:   ID(3),
			Name: NodeName("node3"),
			Addr: NodeAddress("localhost:4000"),
			Type: GW,
			Stat: Alive,
		},
	}

	testMap := CMap{
		Version: 1,
		Nodes:   testNodes,
	}

	ct, err := newCMapManager("")
	if err != nil {
		t.Fatal(err)
	}
	if err = ct.Update(&testMap); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("./cmap")

	for _, n := range testMap.Nodes {
		// 1. Search with the all conditions are matched.
		c := ct.SearchCallNode()
		c.ID(n.ID).Name(n.Name).Type(n.Type).Status(n.Stat)

		if find, err := c.Do(); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(find, n) {
			t.Errorf("expected %+v, got %+v", n, find)
		}

		// 2. Search with only condition for id.
		c = ct.SearchCallNode()
		c.ID(n.ID)

		if find, err := c.Do(); err != nil {
			t.Error(err)
			t.Errorf("%+v", c)
		} else if !reflect.DeepEqual(find, n) {
			t.Errorf("expected %+v, got %+v", n, find)
		}

		// 3. Search with wrong contition.
		c = ct.SearchCallNode()
		c.Name(n.Name + "wrong name")

		if _, err := c.Do(); err == nil {
			t.Error("expected error, got nil")
		}

		// 4. Search with type and status.
		c = ct.SearchCallNode()
		c.Type(n.Type).Status(n.Stat)

		if find, err := c.Do(); err != nil {
			t.Error(err)
		} else if !reflect.DeepEqual(find, n) {
			t.Errorf("expected %+v, got %+v", n, find)
		}
	}
}
