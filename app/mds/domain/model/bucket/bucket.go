package bucket

import (
	"errors"
	"strconv"
)

var (
	// ErrDuplicateEntry is used when they try to save already existed.
	ErrDuplicateEntry = errors.New("duplicated entry exists")
)

// Bucket is an entity of bucket.
type Bucket struct {
	ID     ID
	Name   Name
	User   ID
	Region ID
}

// ID is the ID of bucket, user, region.
type ID int64

func (i ID) String() string {
	return strconv.FormatInt(int64(i), 10)
}

// Name is the name of bucket.
type Name string

func (n Name) String() string {
	return string(n)
}

// Repository provides to access bucket databse.
type Repository interface {
	Save(*Bucket) error
}
