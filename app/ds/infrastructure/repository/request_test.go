package repository

import (
	"testing"
)

func TestRequests(t *testing.T) {
	testCases := []Request{
		{Op: -9},
		{Op: Read, Vol: ""},
		{Op: Read, Oid: ""},
		{Op: Read, Out: nil},
		{Op: Write, Vol: ""},
		{Op: Write, Oid: ""},
		{Op: Write, In: nil},
		{Op: Delete, Vol: ""},
		{Op: Delete, Oid: ""},
	}

	for _, c := range testCases {
		if c.Verify() == nil {
			t.Errorf("%v: expected error, got nil", c)
		}
	}
}
