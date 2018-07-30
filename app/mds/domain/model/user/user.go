package user

import (
	"errors"
	"math/rand"
	"strconv"
	"time"
)

var (
	// ErrNotExist is used when there is no matched user with the search condition.
	ErrNotExist = errors.New("no users match with the given condition")

	// ErrInternal is used when the internal error is occured.
	ErrInternal = errors.New("internal error")
)

// User is an entity of user information.
type User struct {
	ID     ID
	Name   Name
	Access Key
	Secret Key
}

// ID is the ID of user.
type ID int64

func (i ID) String() string {
	return strconv.FormatInt(int64(i), 10)
}

// Name is the name of user.
type Name string

func (n Name) String() string {
	return string(n)
}

// Key is the type of access or secret key.
type Key string

func (k Key) String() string {
	return string(k)
}

const (
	keyLength = 20
	letters   = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

// GenKey generates new access key.
func GenKey() Key {
	b := make([]byte, keyLength)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return Key(string(b))
}

// Repository provides to access user database.
type Repository interface {
	FindByID(ID) (*User, error)
	FindByAk(access Key) (*User, error)
	Save(*User) error
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
