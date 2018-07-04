package cred

import "errors"
import "time"

// ErrInvalidKey is used when a key is not valid.
var ErrInvalidKey = errors.New("invalid key")

// ErrNoSuchCred is used when the repository has no matched credential
// with the given access key.
var ErrNoSuchCred = errors.New("no such credential with the given access key")

// Cred is an entity which has information related to the user's credentials.
type Cred struct {
	access Key
	secret Key
	expire time.Time
}

const defaultCredExpireTime = 10 * time.Minute

// AccessKey returns an access key of credential.
func (c *Cred) AccessKey() Key {
	return c.access
}

// SecretKey returns a secret key of credential.
func (c *Cred) SecretKey() Key {
	return c.secret
}

// Auth tries to authenticate user with the given secret key.
func (c *Cred) Auth(given Key) bool {
	return c.secret.equal(given)
}

// IsExpired returns the credential is expired or not.
func (c *Cred) IsExpired() bool {
	return time.Now().After(c.expire)
}

// New creates a new user credential with the given keys.
func New(access, secret Key) (*Cred, error) {
	if !access.isValid() || !secret.isValid() {
		return nil, ErrInvalidKey
	}

	return &Cred{
		access: access,
		secret: secret,
		expire: time.Now().Add(defaultCredExpireTime),
	}, nil
}

// Key is the type of access key or secret key.
type Key string

// maxKeyLength is the maximum length of keys.
const maxKeyLength = 30

func (k Key) String() string {
	return string(k)
}

func (k Key) isValid() bool {
	if len(k) > maxKeyLength {
		return false
	}
	return true
}

func (k Key) equal(other Key) bool {
	return k == other
}

// Repository provides access a cred store.
type Repository interface {
	Store(c *Cred)
	Find(access Key) (*Cred, error)
}
