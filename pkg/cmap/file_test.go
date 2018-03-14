package cmap

import (
	"os"
	"reflect"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	testMap := CMap{
		OutDated: true,
		Nodes: []Node{
			{
				Addr: "localhost:1000",
				Type: MDS,
				Stat: Alive,
			},
			{
				Addr: "localhost:2000",
				Type: DS,
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

	testFile := "test.xml"
	if _, err := os.Create(testFile); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(testFile)

	if err := encode(testMap, testFile); err != nil {
		t.Fatal(err)
	}

	m, err := decode(testFile)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(m, testMap) {
		t.Error("encoded value and decoded value are not equal")
	}
}
