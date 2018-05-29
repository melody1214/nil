package cmap

import (
	"reflect"
	"testing"
)

func TestMergeCMap(t *testing.T) {
	srcMap := CMap{
		Version: Version(30),
		Nodes: []Node{
			Node{
				ID:   ID(0),
				Type: MDS,
				Addr: NodeAddress("localhost:1000"),
				Stat: NodeAlive,
				Incr: Incarnation(20),
			},
			Node{
				ID:   ID(1),
				Type: DS,
				Addr: NodeAddress("localhost:2000"),
				Stat: NodeAlive,
				Incr: Incarnation(20),
			},
			Node{
				ID:   ID(2),
				Type: DS,
				Addr: NodeAddress("localhost:3000"),
				Stat: NodeAlive,
				Incr: Incarnation(20),
			},
			Node{
				ID:   ID(3),
				Type: DS,
				Addr: NodeAddress("localhost:4000"),
				Stat: NodeAlive,
				Incr: Incarnation(20),
			},
			Node{
				ID:   ID(4),
				Type: DS,
				Addr: NodeAddress("localhost:5000"),
				Stat: NodeAlive,
				Incr: Incarnation(20),
			},
			Node{
				ID:   ID(5),
				Type: DS,
				Addr: NodeAddress("localhost:6000"),
				Stat: NodeAlive,
				Incr: Incarnation(20),
			},
		},
		Vols: []Volume{
			Volume{
				ID:   ID(0),
				Node: ID(1),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(1),
				Node: ID(1),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(2),
				Node: ID(1),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(3),
				Node: ID(2),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(4),
				Node: ID(2),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(5),
				Node: ID(3),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(6),
				Node: ID(4),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(7),
				Node: ID(4),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(8),
				Node: ID(5),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
		},
		EncGrps: []EncodingGroup{
			EncodingGroup{
				ID:   ID(0),
				Size: 10,
				Used: 4,
				Free: 6,
				Stat: EGAlive,
				Vols: []EGVol{EGVol{ID: 0}, EGVol{ID: 3}, EGVol{ID: 6}},
				Incr: Incarnation(30),
			},
			EncodingGroup{
				ID:   ID(1),
				Size: 10,
				Used: 2,
				Free: 8,
				Stat: EGAlive,
				Vols: []EGVol{EGVol{ID: 1}, EGVol{ID: 4}, EGVol{ID: 8}},
				Incr: Incarnation(30),
			},
			EncodingGroup{
				ID:   ID(2),
				Size: 10,
				Used: 4,
				Free: 6,
				Stat: EGAlive,
				Vols: []EGVol{EGVol{ID: 1}, EGVol{ID: 6}, EGVol{ID: 8}},
				Incr: Incarnation(30),
			},
		},
	}
	dstMap := CMap{
		Version: Version(31),
		Nodes: []Node{
			Node{
				ID:   ID(0),
				Type: MDS,
				Addr: NodeAddress("localhost:1000"),
				Stat: NodeAlive,
				Incr: Incarnation(20),
			},
			Node{
				ID:   ID(1),
				Type: DS,
				Addr: NodeAddress("localhost:2000"),
				Stat: NodeAlive,
				Incr: Incarnation(20),
			},
			Node{
				ID:   ID(2),
				Type: DS,
				Addr: NodeAddress("localhost:3000"),
				Stat: NodeAlive,
				Incr: Incarnation(20),
			},
			Node{
				ID:   ID(3),
				Type: DS,
				Addr: NodeAddress("localhost:4000"),
				Stat: NodeAlive,
				Incr: Incarnation(20),
			},
			Node{
				ID:   ID(4),
				Type: DS,
				Addr: NodeAddress("localhost:5000"),
				Stat: NodeAlive,
				Incr: Incarnation(20),
			},
			Node{
				ID:   ID(5),
				Type: DS,
				Addr: NodeAddress("localhost:6000"),
				Stat: NodeAlive,
				Incr: Incarnation(20),
			},
		},
		Vols: []Volume{
			Volume{
				ID:   ID(0),
				Node: ID(1),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(1),
				Node: ID(1),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(2),
				Node: ID(1),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(3),
				Node: ID(2),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(4),
				Node: ID(2),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(5),
				Node: ID(3),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(6),
				Node: ID(4),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(7),
				Node: ID(4),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
			Volume{
				ID:   ID(8),
				Node: ID(5),
				Size: 100,
				Stat: VolActive,
				Incr: Incarnation(30),
			},
		},
		EncGrps: []EncodingGroup{
			EncodingGroup{
				ID:   ID(0),
				Size: 10,
				Used: 4,
				Free: 6,
				Stat: EGAlive,
				Vols: []EGVol{EGVol{ID: 0}, EGVol{ID: 3}, EGVol{ID: 6}},
				Incr: Incarnation(30),
			},
			EncodingGroup{
				ID:   ID(1),
				Size: 10,
				Used: 2,
				Free: 8,
				Stat: EGAlive,
				Vols: []EGVol{EGVol{ID: 1}, EGVol{ID: 4}, EGVol{ID: 8}},
				Incr: Incarnation(30),
			},
			EncodingGroup{
				ID:   ID(2),
				Size: 10,
				Used: 4,
				Free: 6,
				Stat: EGAlive,
				Vols: []EGVol{EGVol{ID: 1}, EGVol{ID: 6}, EGVol{ID: 8}},
				Incr: Incarnation(30),
			},
		},
	}

	expectedVols := make([]Volume, 0)

	// Case 1.
	srcMap.Vols[2].Incr = Incarnation(dstMap.Vols[2].Incr.Uint32() - 1)
	srcMap.Vols[2].Stat = VolPrepared
	expectedVols = append(expectedVols, dstMap.Vols[2])

	mergeCMap(&srcMap, &dstMap)
	for _, v := range expectedVols {
		if reflect.DeepEqual(dstMap.Vols[v.ID.Int64()], v) == false {
			t.Errorf("expected %+v, got %+v", v, dstMap.Vols[v.ID.Int64()])
		}
	}

	expectedEncGrps := make([]EncodingGroup, 0)

	// Case 1.
	srcMap.EncGrps[0].Incr = Incarnation(dstMap.EncGrps[0].Incr.Uint32() - 1)
	srcMap.EncGrps[0].Vols = []EGVol{EGVol{ID: 0}, EGVol{ID: 4}, EGVol{ID: 6}}
	expectedEncGrps = append(expectedEncGrps, dstMap.EncGrps[0])

	// Case 2.
	srcMap.EncGrps[1].Incr = Incarnation(dstMap.EncGrps[1].Incr.Uint32() + 1)
	srcMap.EncGrps[1].Used = 3
	srcMap.EncGrps[1].Free = 7
	expectedEncGrps = append(expectedEncGrps, srcMap.EncGrps[1])

	// Case 3.
	srcMap.EncGrps[2].Incr = Incarnation(dstMap.EncGrps[2].Incr.Uint32() + 1)
	srcMap.EncGrps[2].Used = 5
	srcMap.EncGrps[2].Free = 5
	dstMap.EncGrps[2].Vols = []EGVol{EGVol{ID: 1}, EGVol{ID: 5}, EGVol{ID: 8}}
	expectedEncGrp := EncodingGroup{
		ID:   dstMap.EncGrps[2].ID,
		Size: srcMap.EncGrps[2].Size,
		Used: srcMap.EncGrps[2].Used,
		Free: srcMap.EncGrps[2].Free,
		Stat: srcMap.EncGrps[2].Stat,
		Vols: dstMap.EncGrps[2].Vols,
		Incr: srcMap.EncGrps[2].Incr,
	}
	expectedEncGrps = append(expectedEncGrps, expectedEncGrp)

	mergeCMap(&srcMap, &dstMap)
	for _, eg := range expectedEncGrps {
		if reflect.DeepEqual(dstMap.EncGrps[eg.ID.Int64()], eg) == false {
			t.Errorf("expected %+v, got %+v", eg, dstMap.EncGrps[eg.ID.Int64()])
		}
	}
}

func TestCopyCMap(t *testing.T) {
	origin := CMap{
		Version: Version(0),
		Time:    Now(),
		Nodes: []Node{
			Node{
				ID:   ID(0),
				Type: MDS,
				Addr: NodeAddress("localhost:1000"),
				Stat: NodeAlive,
				Incr: Incarnation(20),
				Vols: []ID{
					ID(0), ID(1), ID(2),
				},
			},
		},
		Vols: []Volume{
			Volume{
				ID:    ID(0),
				Speed: Low,
				Stat:  VolActive,
				EncGrps: []ID{
					ID(0), ID(1), ID(2),
				},
			},
		},
		EncGrps: []EncodingGroup{
			EncodingGroup{
				ID:   ID(0),
				Stat: EGAlive,
				Incr: Incarnation(10),
				Size: 100,
				Vols: []EGVol{EGVol{ID: 0}, EGVol{ID: 1}, EGVol{ID: 2}},
			},
		},
	}
	copied := copyCMap(&origin)

	// Check copied correctly.
	if reflect.DeepEqual(origin, *copied) == false {
		t.Fatalf("copied map copied differently\norigin: %v\ncopied: %v", origin, *copied)
	}

	// Change slice values for testing deep copy.
	origin.Nodes[0].Vols[0] = ID(99)
	origin.Vols[0].EncGrps[0] = ID(99)
	origin.EncGrps[0].Vols[0] = EGVol{ID: 99}
	if reflect.DeepEqual(origin, *copied) {
		t.Fatalf("slices are not deep copied.\norigin: %v\ncopied: %v", origin, *copied)
	}
}
