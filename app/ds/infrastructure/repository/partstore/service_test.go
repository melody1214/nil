package partstore

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/chanyoung/nil/app/ds/infrastructure/repository"
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
			Size:      50000000,
			Speed:     repository.High,
			ChunkSize: 10000000,
			NumOfPart: 5,
			SubPartGroup: repository.SubPartGroup{
				Cold: repository.PartGrpInfo{
					NumOfPart: 2,
					PartInfo:  make(map[string]repository.PartInfo),
				},
				Hot: repository.PartGrpInfo{
					NumOfPart: 3,
					PartInfo:  make(map[string]repository.PartInfo),
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
			Size:      50000000,
			Speed:     repository.High,
			ChunkSize: 10000000,
			NumOfPart: 5,
			SubPartGroup: repository.SubPartGroup{
				Cold: repository.PartGrpInfo{
					DiskSched: 0,
					NumOfPart: 2,
					PartInfo:  make(map[string]repository.PartInfo),
				},
				Hot: repository.PartGrpInfo{
					DiskSched: 0,
					NumOfPart: 3,
					PartInfo:  make(map[string]repository.PartInfo),
				},
			},
			ChunkMap: make(map[string]repository.ChunkMap),
			ObjMap:   make(map[string]repository.ObjMap),
		},
	}

	s := newService(dir)
	s.pgs[pg1.Name] = pg1
	s.pgs[pg2.Name] = pg2

	s.devs["cold_part1"] = &dev{
		Name:      "cold_dev1",
		State:     "Standby",
		Timestamp: uint(time.Now().Nanosecond()),
		Size:      50000000,
		Free:      50000000,
		Used:      0,
	}
	s.devs["cold_part2"] = &dev{
		Name:      "cold_dev2",
		State:     "Standby",
		Timestamp: uint(time.Now().Nanosecond()),
		Size:      50000000,
		Free:      50000000,
		Used:      0,
	}
	s.devs["hot_part1"] = &dev{
		Name:      "hot_dev1",
		State:     "Active",
		Timestamp: uint(time.Now().Nanosecond()),
		Size:      50000000,
		Free:      50000000,
		Used:      0,
	}

	s.devs["hot_part2"] = &dev{
		Name:      "hot_dev2",
		State:     "Active",
		Timestamp: uint(time.Now().Nanosecond()),
		Size:      50000000,
		Free:      50000000,
		Used:      0,
	}
	s.devs["hot_part3"] = &dev{
		Name:      "hot_dev3",
		State:     "Active",
		Timestamp: uint(time.Now().Nanosecond()),
		Size:      50000000,
		Free:      50000000,
		Used:      0,
	}

	go s.Run()
	runtime.Gosched()

	/*
		f, err := os.OpenFile("dummy.txt", os.O_RDONLY, 0700)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		r := &repository.Request{
			Op:     repository.Write,
			Vol:    "pg1",
			LocGid: "lg1",
			Oid:    "dummy.txt",
			Cid:    "chunk2",
			Type:   "Parity",
			Osize:  int64(64),
			In:     f,
			Out:    nil,
		}
		s.Push(r)
		err = r.Wait()
		if err != nil {
			t.Error(err)
		}

		var b bytes.Buffer
		writer := bufio.NewWriter(&b)
		r = &repository.Request{
			Op:     repository.Read,
			Vol:    "pg1",
			LocGid: "lg1",
			Oid:    "dummy.txt",
			Cid:    "chunk2",
			Type:   "Parity",
			Osize:  int64(64),
			In:     nil,
			Out:    writer,
		}
		s.Push(r)
		err = r.Wait()
		if err != nil {
			t.Error(err)
		}
		t.Errorf("%+v\n", hex.Dump(b.Bytes()))
	*/
	testCases := []struct {
		op                 repository.Operation
		pg, lgid, oid, cid string
		typ                string
		size               int64
		Md5                string
		content            string
		result             error
	}{
		{repository.Read, "pg1", "lg1", "banana", "", "Parity", int64(len("banana is good\n")), "banana is good\n", "",
			fmt.Errorf("no such object: banana"),
		},
		{repository.Write, "pg1", "lg1", "banana", "chunk1", "Parity", int64(len("banana is good\n")), "", "banana is good\n", nil},
		{repository.Read, "pg1", "lg1", "banana", "", "Parity", int64(len("banana is good\n")), "", "banana is good\n", nil},
		{repository.Delete, "pg2", "lg2", "apple", "chunk1", "Data", int64(len("apple is sweet\n")), "", "apple is sweet\n",
			fmt.Errorf("no such object: apple"),
		},
		{repository.Read, "pg2", "lg2", "apple", "", "Data", int64(len("apple is sweet\n")), "", "apple is sweet\n",
			fmt.Errorf("no such object: apple"),
		},
		{repository.Write, "pg2", "lg2", "apple", "chunk2", "Parity", int64(len("apple is sweet\n")), "", "apple is sweet\n", nil},
		{repository.Write, "pg2", "lg2", "grapefruit", "chunk2", "Parity", int64(len("grapefruit is sweet\n")), "", "grapefruit is sweet\n", nil},
		{repository.Write, "pg2", "lg2", "melon", "chunk2", "Parity", int64(len("melon is sweet\n")), "", "melon is sweet\n", nil},
		{repository.Read, "pg2", "lg2", "apple", "", "Parity", int64(len("apple is sweet\n")), "", "apple is sweet\n", nil},
		{repository.Delete, "pg2", "lg2", "apple", "chunk2", "Parity", int64(len("apple is sweet\n")), "", "apple is sweet\n", nil},
		{repository.Write, "pg1", "lg1", "blueberry", "chunk1", "Parity", int64(len("blueberry is good\n")), "", "blueberry is good\n", nil},
		{repository.Write, "pg2", "lg2", "cherry", "chunk2", "Parity", int64(len("cherry is good\n")), "", "cherry is good\n", nil},
		{repository.Write, "pg1", "lg2", "strawberry", "chunk7", "Parity", int64(len("strawberry is good\n")), "", "strawberry is good\n", nil},
		{repository.Write, "pg1", "lg1", "pineapple", "chunk1", "Parity", int64(len("pineapple is good\n")), "", "pineapple is good\n", nil},
		{repository.Write, "pg1", "lg1", "watermelon", "chunk1", "Parity", int64(len("watermelon is good\n")), "", "watermelon is good\n", nil},
		{repository.Write, "pg1", "lg1", "apple1", "chunk3", "Parity", int64(len("apple is sweet\n")), "", "apple is sweet\n", nil},
		{repository.Write, "pg1", "lg1", "mango", "chunk2", "Parity", int64(len("mango is good\n")), "", "mango is good\n", nil},
		{repository.Write, "pg1", "lg1", "orange", "chunk1", "Parity", int64(len("orange is good\n")), "", "orange is good\n", nil},
		{repository.Write, "pg1", "lg1", "bit", "chunk4", "Parity", int64(len("bit is good\n")), "", "bit is good\n", nil},
		{repository.Read, "pg1", "lg1", "orange", "", "Parity", int64(len("orange is good\n")), "", "orange is good\n", nil},
		{repository.Write, "pg1", "lg1", "redmango", "chunk5", "Parity", int64(len("redmango is good\n")), "", "redmango is good\n", nil},
		{repository.Write, "pg1", "lg1", "kiwi", "chunk6", "Parity", int64(len("kiwi is good\n")), "", "kiwi is good\n", nil},
		{repository.DeleteReal, "pg1", "lg1", "orange", "", "Parity", int64(len("orange is good\n")), "", "orange is good\n", nil},
		{repository.DeleteReal, "pg1", "lg1", "pineapple", "", "Parity", int64(len("pineapple is good\n")), "", "pineapple is good\n",
			fmt.Errorf("can remove only a last object of a chunk")},
		{repository.DeleteReal, "pg1", "lg1", "watermelon", "", "Parity", int64(len("watermelon is good\n")), "", "watermelon is good\n", nil},
		{repository.DeleteReal, "pg1", "lg1", "", "chunk2", "Parity", int64(len("pineapple is good\n")), "", "pineapple is good\n", nil},
		{repository.Read, "pg1", "lg1", "redmango", "", "Parity", int64(len("redmango is good\n")), "", "redmango is good\n", nil},
		{repository.Read, "pg1", "lg1", "bit", "", "Parity", int64(len("bit is good\n")), "", "bit is good\n", nil},
		{repository.Read, "pg1", "lg1", "pineapple", "", "Parity", int64(len("pineapple is good\n")), "", "pineapple is good\n", nil},
		{repository.Read, "pg1", "lg1", "banana", "", "Parity", int64(len("banana is good\n")), "", "banana is good\n", nil},
		{repository.Read, "pg1", "lg1", "blueberry", "", "Parity", int64(len("blueberry is good\n")), "", "blueberry is good\n", nil},
		{repository.Read, "pg2", "lg2", "melon", "", "Parity", int64(len("melon is sweet\n")), "", "melon is sweet\n", nil},
		{repository.Read, "pg2", "lg2", "grapefruit", "", "Parity", int64(len("grapefruit is sweet\n")), "", "grapefruit is sweet\n", nil},
		{repository.Read, "pg2", "lg2", "cherry", "", "Parity", int64(len("cherry is good\n")), "", "cherry is good\n", nil},
		{repository.Read, "pg2", "lg1", "watermelon", "", "Parity", int64(len("watermelon is good\n")), "", "watermelon is good\n",
			fmt.Errorf("no such object: watermelon")},
	}

	h := md5.New()
	for _, c := range testCases {
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)

		io.WriteString(h, c.content)
		md5 := string(h.Sum(nil))

		r := &repository.Request{
			Op:     c.op,
			Vol:    c.pg,
			LocGid: c.lgid,
			Oid:    c.oid,
			Cid:    c.cid,
			Osize:  c.size,
			Md5:    md5,
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
	s.RenameChunk("chunk2", "L_chunk2", "pg2", "lg2")

	testCases = []struct {
		op                 repository.Operation
		pg, lgid, oid, cid string
		typ                string
		size               int64
		Md5                string
		content            string
		result             error
	}{
		{repository.Read, "pg1", "lg1", "banana", "", "Parity", int64(len("banana is good\n")), "", "banana is good\n", nil},
		{repository.Read, "pg1", "lg1", "apple1", "", "Parity", int64(len("apple is sweet\n")), "", "apple is sweet\n", nil},
		{repository.Read, "pg1", "lg1", "bit", "", "Parity", int64(len("bit is good\n")), "", "bit is good\n", nil},
		{repository.Read, "pg2", "lg2", "melon", "", "Parity", int64(len("melon is sweet\n")), "", "melon is sweet\n", nil},
		{repository.Read, "pg2", "lg2", "cherry", "", "Parity", int64(len("cherry is good\n")), "", "cherry is good\n", nil},
		{repository.Read, "pg2", "lg2", "grapefruit", "", "Parity", int64(len("grapefruit is sweet\n")), "", "grapefruit is sweet\n", nil},
		{repository.Read, "pg2", "lg2", "melon", "", "Parity", int64(len("melon is sweet\n")), "", "melon is sweet\n", nil},
		{repository.Delete, "pg2", "lg2", "grapefruit", "", "Parity", int64(len("grapefruit is sweet\n")), "", "grapefruit is sweet\n", nil},
		{repository.Delete, "pg2", "lg2", "melon", "", "Parity", int64(len("melon is sweet\n")), "", "melon is sweet\n", nil},
		{repository.Delete, "pg2", "lg2", "cherry", "", "Parity", int64(len("cherry is good\n")), "", "cherry is good\n", nil},
	}

	for _, c := range testCases {
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)

		io.WriteString(h, c.content)
		md5 := string(h.Sum(nil))

		r := &repository.Request{
			Op:     c.op,
			Vol:    c.pg,
			LocGid: c.lgid,
			Oid:    c.oid,
			Cid:    c.cid,
			Osize:  c.size,
			Md5:    md5,
			In:     strings.NewReader(c.content),
			Out:    writer,
		}

		s.Push(r)

		err := r.Wait()

		if r.Op == repository.Read {
			t.Logf("readed: %s, length: %d", b.String(), b.Len())
			t.Logf("readed: %v", b.Bytes())
		}

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

	err := s.BuildObjectMap("pg2", "L_chunk2")
	if err != nil {
		t.Errorf("%+v", err)
	}

	//time.Sleep(3 * time.Second)

	testCases = []struct {
		op                 repository.Operation
		pg, lgid, oid, cid string
		typ                string
		size               int64
		Md5                string
		content            string
		result             error
	}{
		{repository.Read, "pg2", "lg2", "apple", "", "Parity", int64(len("apple is sweet\n")), "", "apple is sweet\n", nil},
		{repository.Read, "pg2", "lg2", "cherry", "", "Parity", int64(len("cherry is good\n")), "", "cherry is good\n", nil},
		{repository.Read, "pg2", "lg2", "grapefruit", "", "Parity", int64(len("grapefruit is sweet\n")), "", "grapefruit is sweet\n", nil},
		{repository.Read, "pg2", "lg2", "melon", "", "Parity", int64(len("melon is sweet\n")), "", "melon is sweet\n", nil},
	}

	for _, c := range testCases {
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)

		io.WriteString(h, c.content)
		md5 := string(h.Sum(nil))

		r := &repository.Request{
			Op:     c.op,
			Vol:    c.pg,
			LocGid: c.lgid,
			Oid:    c.oid,
			Cid:    c.cid,
			Osize:  c.size,
			Md5:    md5,
			In:     strings.NewReader(c.content),
			Out:    writer,
		}

		s.Push(r)

		err := r.Wait()

		if r.Op == repository.Read {
			t.Logf("readed: %s, length: %d", b.String(), b.Len())
			t.Logf("readed: %v", b.Bytes())
		}

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

	s.GetNonCodedChunk("pg1", "lg1")

	s.devLock.RLock()
	coldDev1 := s.devs["cold_part1"]
	coldDev2 := s.devs["cold_part2"]
	hotDev1 := s.devs["hot_part1"]
	hotDev2 := s.devs["hot_part2"]
	hotDev3 := s.devs["hot_part3"]
	s.devLock.RUnlock()

	s.devLock.Lock()
	coldDev1.State = "Active"
	s.devLock.Unlock()

	time.Sleep(25 * time.Second)

	s.devLock.RLock()
	t.Logf("cold_dev1, State: %s, ActiveTime: %d, StandbyTime: %d, TotalIO: %d, Free: %d, Used: %d",
		coldDev1.State, coldDev1.ActiveTime, coldDev1.StandbyTime, coldDev1.TotalIO, coldDev1.Free, coldDev1.Used)
	t.Logf("cold_dev2, State: %s, ActiveTime: %d, StandbyTime: %d, TotalIO: %d, Free: %d, Used: %d",
		coldDev2.State, coldDev2.ActiveTime, coldDev2.StandbyTime, coldDev2.TotalIO, coldDev2.Free, coldDev2.Used)
	t.Logf("hot_dev1, State: %s, ActiveTime: %d, StandbyTime: %d, TotalIO: %d, Free: %d, Used: %d",
		hotDev1.State, hotDev1.ActiveTime, hotDev1.StandbyTime, hotDev1.TotalIO, hotDev1.Free, hotDev1.Used)
	t.Logf("hot_dev2, State: %s, ActiveTime: %d, StandbyTime: %d, TotalIO: %d, Free: %d, Used: %d",
		hotDev2.State, hotDev2.ActiveTime, hotDev2.StandbyTime, hotDev2.TotalIO, hotDev2.Free, hotDev2.Used)
	t.Logf("hot_dev3, State: %s, ActiveTime: %d, StandbyTime: %d, TotalIO: %d, Free: %d, Used: %d",
		hotDev3.State, hotDev3.ActiveTime, hotDev3.StandbyTime, hotDev3.TotalIO, hotDev3.Free, hotDev3.Used)
	s.devLock.RUnlock()

	//s.MigrateData("cold_part2")

}
