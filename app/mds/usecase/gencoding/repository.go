package gencoding

import (
	"github.com/chanyoung/nil/app/mds/repository"
)

// Repository provides access to gencoding database.
type Repository interface {
	Begin() (repository.TxID, error)
	Rollback(repository.TxID) error
	Commit(repository.TxID) error
}
