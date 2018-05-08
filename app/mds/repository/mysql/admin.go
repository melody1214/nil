package mysql

import (
	"github.com/chanyoung/nil/app/mds/usecase/admin"
)

type adminStore struct {
	*Store
}

// NewAdminRepository returns a new instance of a mysql admin repository.
func NewAdminRepository(s *Store) admin.Repository {
	return &adminStore{
		Store: s,
	}
}
