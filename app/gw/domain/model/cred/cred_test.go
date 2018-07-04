package cred

import (
	"testing"
)

func TestAuth(t *testing.T) {
	access := Key("accessKey")
	secret := Key("secretKey")

	c, err := New(access, secret)
	if err != nil {
		t.Fatal(err)
	}

	if !c.AccessKey().equal(access) {
		t.Errorf("expected the access key is %+v, but got %+v", access, c.AccessKey())
	}

	if !c.Auth(secret) {
		t.Error("expected the authenticate result is true, but false")
	}
}
