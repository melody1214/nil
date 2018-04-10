package repository

import "errors"

var (
	// ErrNotExist means there is no rows which are matched conditions.
	ErrNotExist = errors.New("no condition matched rows")

	// ErrDuplicateEntry means store already has a entry which the
	// primary key value is same with you given.
	ErrDuplicateEntry = errors.New("duplicate entry for primary key")
)
