package partstore

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/chanyoung/nil/app/ds/repository"
)

func TestServiceAPIs(t *testing.T) {
	dir := "testServiceAPIs"
	os.Mkdir(dir, 0775)
	defer os.RemoveAll(dir)

	os.Mkdir(dir+"/pg1", 0775)
	pg1 := &pg{
		Vol: &repository.Vol{
			Name:      "pg1",
			MntPoint:  dir + "/pg1",
			Size:      1024,
			Speed:     repository.High,
			ChunkSize: 10000000,
			NumOfPart: 5,
			SubPartGroup: repository.SubPartGroup{
				Cold: repository.PartGrpInfo{
					NumOfPart: 2,
				},
				Hot: repository.PartGrpInfo{
					NumOfPart: 3,
				},
			},
			ChunkMap: make(map[string]repository.ChunkMap),
			ObjMap:   make(map[string]repository.ObjMap),
		},
	}

	os.Mkdir(dir+"/pg2", 0775)
	pg2 := &pg{
		Vol: &repository.Vol{
			Name:      "pg2",
			MntPoint:  dir + "/pg2",
			Size:      1024,
			Speed:     repository.High,
			ChunkSize: 10000000,
			NumOfPart: 5,
			SubPartGroup: repository.SubPartGroup{
				Cold: repository.PartGrpInfo{
					DiskSched: 0,
					NumOfPart: 2,
				},
				Hot: repository.PartGrpInfo{
					DiskSched: 0,
					NumOfPart: 3,
				},
			},
			ChunkMap: make(map[string]repository.ChunkMap),
			ObjMap:   make(map[string]repository.ObjMap),
		},
	}

	s := newService(dir)
	s.pgs[pg1.Name] = pg1
	s.pgs[pg2.Name] = pg2

	go s.Run()
	runtime.Gosched()

	testCases := []struct {
		op                 repository.Operation
		pg, lgid, oid, cid string
		typ                string
		size               int64
		content            string
		result             error
	}{
		{repository.Read, "pg1", "lg1", "banana", "chunk1", "Parity", int64(len("banana is good\n")), "banana is good\n",
			fmt.Errorf("no such object: banana"),
		},
		{repository.Write, "pg1", "lg1", "banana", "chunk1", "Parity", int64(len("banana is good\n")), "banana is good\n", nil},
		{repository.Read, "pg1", "lg1", "banana", "chunk1", "Data", int64(len("banana is good\n")), "banana is good\n", nil},
		{repository.Delete, "pg2", "lg2", "apple", "chunk1", "Data", int64(len("apple is sweet\n")), "apple is sweet\n",
			fmt.Errorf("no such object: apple"),
		},
		{repository.Read, "pg2", "lg2", "apple", "chunk2", "Data", int64(len("apple is sweet\n")), "apple is sweet\n",
			fmt.Errorf("no such object: apple"),
		},
		{repository.Write, "pg2", "lg2", "apple", "chunk2", "Parity", int64(len("apple is sweet\n")), "apple is sweet\n", nil},
		{repository.Write, "pg2", "lg2", "grapefruit", "chunk2", "Parity", int64(len("grapefruit is sweet\n")), "grapefruit is sweet\n", nil},
		{repository.Write, "pg2", "lg2", "melon", "chunk2", "Parity", int64(len("melon is sweet\n")), "melon is sweet\n", nil},
		{repository.Read, "pg2", "lg2", "apple", "chunk2", "Parity", int64(len("apple is sweet\n")), "apple is sweet\n", nil},
		{repository.Delete, "pg2", "lg2", "apple", "chunk2", "Parity", int64(len("apple is sweet\n")), "apple is sweet\n", nil},
		{repository.Write, "pg1", "lg1", "banana", "chunk1", "Parity", int64(len("banana is good\n")), "banana is good\n", nil},
		{repository.Write, "pg1", "lg1", "banana", "chunk2", "Parity", int64(len("banana is good\n")), "banana is good\n", nil},
		{repository.Write, "pg1", "lg2", "banana", "chunk7", "Parity", int64(len("banana is good\n")), "banana is good\n", nil},
		{repository.Write, "pg1", "lg1", "pineapple", "chunk1", "Parity", int64(len("pineapple is good\n")), "pineapple is good\n", nil},
		{repository.Write, "pg1", "lg1", "watermelon", "chunk1", "Parity", int64(len("watermelon is good\n")), "watermelon is good\n", nil},
		{repository.Write, "pg1", "lg1", "apple1", "chunk3", "Parity", int64(len("apple is sweet\n")), "apple is sweet\n", nil},
		{repository.Write, "pg1", "lg1", "banana", "chunk2", "Parity", int64(len("banana is good\n")), "banana is good\n", nil},
		{repository.Write, "pg1", "lg1", "orange", "chunk1", "Parity", int64(len("orange is good\n")), "orange is good\n", nil},
		{repository.Write, "pg1", "lg1", "banana", "chunk4", "Parity", int64(len("banana is good\n")), "banana is good\n", nil},
		{repository.Read, "pg1", "lg1", "orange", "chunk1", "Parity", int64(len("orange is good\n")), "orange is good\n", nil},
		{repository.Write, "pg1", "lg1", "banana", "chunk5", "Parity", int64(len("banana is good\n")), "banana is good\n", nil},
		{repository.Write, "pg1", "lg1", "banana", "chunk6", "Parity", int64(len("banana is good\n")), "banana is good\n", nil},
		{repository.DeleteReal, "pg1", "lg1", "orange", "", "Parity", int64(len("orange is good\n")), "orange is good\n", nil},
		{repository.DeleteReal, "pg1", "lg1", "pineapple", "", "Parity", int64(len("pineapple is good\n")), "pineapple is good\n",
			fmt.Errorf("can remove only a last object of a chunk")},
		{repository.DeleteReal, "pg1", "lg1", "watermelon", "", "Parity", int64(len("watermelon is good\n")), "watermelon is good\n", nil},
		{repository.DeleteReal, "pg1", "lg1", "", "chunk2", "Parity", int64(len("pineapple is good\n")), "pineapple is good\n", nil},
	}

	for _, c := range testCases {
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)

		r := &repository.Request{
			Op:     c.op,
			Vol:    c.pg,
			LocGid: c.lgid,
			Oid:    c.oid,
			Cid:    c.cid,
			Osize:  c.size,
			In:     strings.NewReader(c.content),
			Out:    writer,
		}

		s.Push(r)

		err := r.Wait()
		if err != nil && c.result != nil {
			if err.Error() != c.result.Error() {
				t.Errorf("%v %s/%s: expected response %v, got %v", c.op, c.pg, c.oid, c.result, err)
			}
			continue
		} else if err != c.result {
			t.Errorf("%v %s/%s: expected response %v, got %v", c.op, c.pg, c.oid, c.result, err)
			continue
		}

		if c.op == repository.Read && b.String() != c.content {
			t.Errorf("%v %s/%s: expected data %v, got %v", c.op, c.pg, c.oid, c.content, b.String())
			continue
		}
	}

	s.RenameChunk("chunk3", "L_chunk3", "pg1", "lg1")
	s.RenameChunk("chunk4", "L_chunk4", "pg1", "lg1")
	s.RenameChunk("L_chunk4", "G_chunk4", "pg1", "lg1")
	s.RenameChunk("chunk2", "G_chunk2", "pg2", "lg2")

	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	r := &repository.Request{
		Op:     repository.Read,
		Vol:    "pg1",
		LocGid: "lg1",
		Oid:    "apple1",
		Osize:  int64(len("apple is sweet")),
		Out:    writer,
	}
	s.Push(r)
	err := r.Wait()
	if err != nil {
		t.Errorf("%+v", err)
	}
	t.Logf("readed: %s", b.String())

	s.GetNonCodedChunk("pg1", "lg1")

	//time.Sleep(13 * time.Second)
	s.MigrateData()
}
