package lvstore

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/chanyoung/nil/pkg/ds/store/request"
	"github.com/chanyoung/nil/pkg/ds/store/volume"
)

func TestServiceAPIs(t *testing.T) {
	dir := "testServiceAPIs"
	os.Mkdir(dir, 0775)
	defer os.RemoveAll(dir)

	os.Mkdir(dir+"/lv1", 0775)
	lv1 := &lv{
		Vol: &volume.Vol{
			Name:     "lv1",
			MntPoint: dir + "/lv1",
			Size:     1024,
			Speed:    volume.High,
		},
	}

	os.Mkdir(dir+"/lv2", 0775)
	lv2 := &lv{
		Vol: &volume.Vol{
			Name:     "lv2",
			MntPoint: dir + "/lv2",
			Size:     1024,
			Speed:    volume.High,
		},
	}

	s := NewService(dir)
	s.lvs[lv1.Name] = lv1
	s.lvs[lv2.Name] = lv2

	go s.Run()
	runtime.Gosched()

	testCases := []struct {
		op      request.Operation
		lv, oid string
		content string
		result  error
	}{
		{request.Read, "lv1", "banana", "banana is good\n",
			fmt.Errorf("%s %s: %s", "open", dir+"/lv1/banana", "no such file or directory"),
		},
		{request.Write, "lv1", "banana", "banana is good\n", nil},
		{request.Read, "lv1", "banana", "banana is good\n", nil},
		{request.Delete, "lv2", "apple", "apple is sweet\n",
			fmt.Errorf("%s %s: %s", "remove", dir+"/lv2/apple", "no such file or directory"),
		},
		{request.Write, "lv2", "apple", "apple is sweet\n", nil},
		{request.Read, "lv2", "apple", "apple is sweet\n", nil},
		{request.Delete, "lv2", "apple", "apple is sweet\n", nil},
		{request.Read, "lv2", "apple", "apple is sweet\n",
			fmt.Errorf("%s %s: %s", "open", dir+"/lv2/apple", "no such file or directory"),
		},
	}

	for _, c := range testCases {
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)

		r := &request.Request{
			Op:  c.op,
			Vol: c.lv,
			Oid: c.oid,
			In:  strings.NewReader(c.content),
			Out: writer,
		}

		s.Push(r)

		err := r.Wait()
		if err != nil && c.result != nil {
			if err.Error() != c.result.Error() {
				t.Errorf("%v %s/%s: expected response %v, got %v", c.op, c.lv, c.oid, c.result, err)
			}
			continue
		} else if err != c.result {
			t.Errorf("%v %s/%s: expected response %v, got %v", c.op, c.lv, c.oid, c.result, err)
			continue
		}

		if c.op == request.Read && b.String() != c.content {
			t.Errorf("%v %s/%s: expected data %v, got %v", c.op, c.lv, c.oid, c.content, b.String())
			continue
		}
	}
}