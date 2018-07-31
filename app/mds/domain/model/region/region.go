package region

import (
	"errors"
	"strconv"
)

var (
	// ErrNotExist is used when there is no matched region with the search condition.
	ErrNotExist = errors.New("no region match with the given condition")

	// ErrInternal is used when the internal error is occured.
	ErrInternal = errors.New("internal error")
)

// Region means a single IDC.
type Region struct {
	ID       ID
	Name     Name
	EndPoint EndPoint
}

// ID is the ID of region.
type ID int64

func (i ID) String() string {
	return strconv.FormatInt(int64(i), 10)
}

// Name is the name of user.
type Name string

func (n Name) String() string {
	return string(n)
}

// EndPoint is the endpoint of region.
type EndPoint string

func (p EndPoint) String() string {
	return string(p)
}

// Repository provides to access region database.
type Repository interface {
	FindByID(ID) (*Region, error)
	FindByName(Name) (*Region, error)
	Create(*Region) error
}
