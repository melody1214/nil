package repository

import "errors"

// ErrNotExist means there is no rows which are matched conditions.
var ErrNotExist = errors.New("no condition matched rows")
