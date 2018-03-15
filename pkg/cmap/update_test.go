package cmap

import (
	"os"
	"testing"
)

func TestGetMdsAddr(t *testing.T) {
	mdsAddr := "localhost:1000"

	testMap := CMap{
		Version: 999999,
		Nodes: []Node{
			{
				Addr: mdsAddr,
				Type: MDS,
				Stat: Alive,
			},
			{
				Addr: "localhost:3000",
				Type: DS,
				Stat: Faulty,
			},
			{
				Addr: "localhost:4000",
				Type: GW,
				Stat: Alive,
			},
		},
	}

	if err := store(&testMap); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(baseDir)

	mds, err := getMdsAddr()
	if err != nil {
		t.Fatal(err)
	}
	if mds != mdsAddr {
		t.Errorf("expected address %s, got %s", mdsAddr, mds)
	}
}
