package cmap

import (
	"os"
	"reflect"
	"testing"
	"time"
)

func TestGetLatestMapFile(t *testing.T) {
	testCases := []CMap{
		{Version: 0},
		{Version: 1},
		{Version: 2},
	}

	defer os.RemoveAll(baseDir)
	latestVer := Version(0)
	for _, c := range testCases {
		if latestVer < c.Version {
			latestVer = c.Version
		}

		path := filePath(c.Version.Int64())

		if err := createFile(path); err != nil {
			t.Fatal(err)
		}
	}

	f, err := getLatestMapFile()
	if err != nil {
		t.Fatal(err)
	}

	if f != filePath(latestVer.Int64()) {
		t.Errorf("got %s, expected %s", f, filePath(latestVer.Int64()))
	}
}

func TestEncodeDecode(t *testing.T) {
	testMap := CMap{
		Version: 1,
		Time:    time.Now().UTC().String(),
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
		Vols: []Volume{
			{
				ID:   ID(1),
				Size: 100000,
			},
			{
				ID:   ID(2),
				Size: 100000,
			},
		},
		EncGrps: []EncodingGroup{
			{
				ID: ID(1),
			},
			{
				ID: ID(2),
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
