package store

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"
)

func TestServiceAPIs(t *testing.T) {
	dir := "testServiceAPIs"
	os.Mkdir(dir, 0775)
	defer os.RemoveAll(dir)

	os.Mkdir(dir+"/lv1", 0775)
	lv1 := &LV{
		Name:     "lv1",
		MntPoint: dir + "/lv1",
		Size:     1024,
		Speed:    High,
	}

	os.Mkdir(dir+"/lv2", 0775)
	lv2 := &LV{
		Name:     "lv2",
		MntPoint: dir + "/lv2",
		Size:     1024,
		Speed:    High,
	}

	s := NewService(dir)
	s.lvs[lv1.Name] = lv1
	s.lvs[lv2.Name] = lv2

	go s.Run()
	runtime.Gosched()

	testCases := []struct {
		op      Operation
		lv, oid string
		content string
		result  error
	}{
		{Read, "lv1", "banana", "banana is good\n",
			fmt.Errorf("%s %s: %s", "open", dir+"/lv1/banana", "no such file or directory"),
		},
		{Write, "lv1", "banana", "banana is good\n", nil},
		{Read, "lv1", "banana", "banana is good\n", nil},
		{Delete, "lv2", "apple", "apple is sweet\n",
			fmt.Errorf("%s %s: %s", "remove", dir+"/lv2/apple", "no such file or directory"),
		},
		{Write, "lv2", "apple", "apple is sweet\n", nil},
		{Read, "lv2", "apple", "apple is sweet\n", nil},
		{Delete, "lv2", "apple", "apple is sweet\n", nil},
		{Read, "lv2", "apple", "apple is sweet\n",
			fmt.Errorf("%s %s: %s", "open", dir+"/lv2/apple", "no such file or directory"),
		},
	}

	for _, c := range testCases {
		var b bytes.Buffer
		writer := bufio.NewWriter(&b)

		call := s.Call().
			Operation(c.op).
			Lv(c.lv).
			ObjectID(c.oid).
			InputStream(strings.NewReader(c.content)).
			OutputStream(writer)

		err := call.Do()
		if err != nil && c.result != nil {
			if err.Error() != c.result.Error() {
				t.Errorf("%s %s/%s: expected response %v, got %v", c.op, c.lv, c.oid, c.result, err)
			}
			continue
		} else if err != c.result {
			t.Errorf("%s %s/%s: expected response %v, got %v", c.op, c.lv, c.oid, c.result, err)
			continue
		}

		if c.op == Read && b.String() != c.content {
			t.Errorf("%s %s/%s: expected data %v, got %v", c.op, c.lv, c.oid, c.content, b.String())
			continue
		}
	}
}
