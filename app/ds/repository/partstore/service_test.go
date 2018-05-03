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

	os.Mkdir(dir+"/part1", 0775)
	part1 := &part{
		Vol: &repository.Vol{
			Name:      "part1",
			MntPoint:  dir + "/part1",
			Size:      1024,
			Speed:     repository.High,
			ChunkSize: 10000000,
			Obj:       make(map[string]repository.Object),
		},
	}

	os.Mkdir(dir+"/part2", 0775)
	part2 := &part{
		Vol: &repository.Vol{
			Name:      "part2",
			MntPoint:  dir + "/part2",
			Size:      1024,
			Speed:     repository.High,
			ChunkSize: 10000000,
			Obj:       make(map[string]repository.Object),
		},
	}

	s := newService(dir)
	s.parts[part1.Name] = part1
	s.parts[part2.Name] = part2

	go s.Run()
	runtime.Gosched()

	testCases := []struct {
		op                   repository.Operation
		part, lgid, oid, cid string
		size                 int64
		content              string
		result               error
	}{
		{repository.Read, "part1", "lg1", "banana", "chunk1", int64(len("banana is good\n")), "banana is good\n",
			fmt.Errorf("no such object: banana"),
		},
		{repository.Write, "part1", "lg1", "banana", "chunk1", int64(len("banana is good\n")), "banana is good\n", nil},
		{repository.Read, "part1", "lg1", "banana", "chunk1", int64(len("banana is good\n")), "banana is good\n", nil},
		{repository.Delete, "part2", "lg2", "apple", "chunk1", int64(len("apple is sweet\n")), "apple is sweet\n",
			fmt.Errorf("no such object: apple"),
		},
		{repository.Read, "part2", "lg2", "apple", "chunk2", int64(len("apple is sweet\n")), "apple is sweet\n",
			fmt.Errorf("no such object: apple"),
		},
		{repository.Write, "part2", "lg2", "apple", "chunk2", int64(len("apple is sweet\n")), "apple is sweet\n", nil},
		{repository.Read, "part2", "lg2", "apple", "chunk2", int64(len("apple is sweet\n")), "apple is sweet\n", nil},
		{repository.Delete, "part2", "lg2", "apple", "chunk2", int64(len("apple is sweet\n")), "apple is sweet\n", nil},
	}

	for _, c := range testCases {
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)

		r := &repository.Request{
			Op:     c.op,
			Vol:    c.part,
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
				t.Errorf("%v %s/%s: expected response %v, got %v", c.op, c.part, c.oid, c.result, err)
			}
			continue
		} else if err != c.result {
			t.Errorf("%v %s/%s: expected response %v, got %v", c.op, c.part, c.oid, c.result, err)
			continue
		}

		if c.op == repository.Read && b.String() != c.content {
			t.Errorf("%v %s/%s: expected data %v, got %v", c.op, c.part, c.oid, c.content, b.String())
			continue
		}
	}
}
