package mysql

import "github.com/chanyoung/nil/app/mds/usecase/gencoding"

type gencodingStore struct {
	*Store
}

// NewGencodingRepository returns a new instance of a gencoding repository.
func NewGencodingRepository(s *Store) gencoding.Repository {
	return &gencodingStore{
		Store: s,
	}
}
