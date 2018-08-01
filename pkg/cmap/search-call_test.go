package cmap

import (
	"os"
	"testing"
)

func TestSearchCallNode(t *testing.T) {
	testNodes := []Node{
		{
			ID:   ID(0),
			Name: NodeName("node0"),
			Addr: NodeAddress("localhost:1000"),
			Type: MDS,
			Stat: NodeAlive,
		},
		{
			ID:   ID(1),
			Name: NodeName("node1"),
			Addr: NodeAddress("localhost:2000"),
			Type: DS,
			Stat: NodeAlive,
		},
		{
			ID:   ID(2),
			Name: NodeName("node2"),
			Addr: NodeAddress("localhost:3000"),
			Type: DS,
			Stat: NodeFaulty,
		},
		{
			ID:   ID(3),
			Name: NodeName("node3"),
			Addr: NodeAddress("localhost:4000"),
			Type: GW,
			Stat: NodeAlive,
		},
	}

	testMap := CMap{
		Version: 1,
		Nodes:   testNodes,
	}

	ct := newManager()
	if err := ct.Update(&testMap); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll("./cmap")

	for _, n := range testMap.Nodes {
		// 1. Search with the all conditions are matched.
		c := ct.SearchCall()
		if find, err := c.Node().ID(n.ID).Name(n.Name).Type(n.Type).Status(n.Stat).Do(); err != nil {
			t.Error(err)
		} else if find.ID != n.ID {
			t.Errorf("expected %+v, got %+v", n, find)
		}

		// 2. Search with only condition for id.
		if find, err := c.Node().ID(n.ID).Do(); err != nil {
			t.Error(err)
			t.Errorf("%+v", c)
		} else if find.ID != n.ID {
			t.Errorf("expected %+v, got %+v", n, find)
		}

		// 3. Search with wrong contition.
		if _, err := c.Node().Name(n.Name + "wrong name").Do(); err == nil {
			t.Error("expected error, got nil")
		}

		// 4. Search with type and status.
		if find, err := c.Node().Type(n.Type).Status(n.Stat).Do(); err != nil {
			t.Error(err)
		} else if find.ID != n.ID {
			t.Errorf("expected %+v, got %+v", n, find)
		}
	}
}
