package inmem

import (
	"testing"
)

func TestAuth(t *testing.T) {
	r := NewAuthRepository()

	testCases := []struct {
		accesskey string
		secretKey string
	}{
		{"accessKey1", "secretKey1"},
		{"accessKey2", "secretKey2"},
	}

	for _, c := range testCases {
		if sk, ok := r.Find(c.accesskey); ok {
			t.Errorf("expected failed to find secret key but found %+v", sk)
		}

		r.Add(c.accesskey, c.secretKey)
		if sk, ok := r.Find(c.accesskey); ok == false {
			t.Errorf("expected to find secret key but found nothing")
		} else if sk != c.secretKey {
			t.Errorf("expected secret key value %+v, but got %+v", c.secretKey, sk)
		}
	}
}
