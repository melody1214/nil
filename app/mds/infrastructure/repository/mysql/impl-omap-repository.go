package mysql

import "github.com/chanyoung/nil/app/mds/domain/model/objectmap"
import "github.com/chanyoung/nil/pkg/omap"

type objectMapRepository struct {
	s *Store
}

func NewObjectMapRepository(s *Store) objectmap.Repository {
	return &objectMapRepository{
		s: s,
	}
}

func (r *objectMapRepository) FindByObject(object, size, bucket omap.Name) (*omap.OMap, error) {
	return nil, nil
}

func (r *objectMapRepository) FindByChunk(chunk omap.Name) (*omap.OMap, error) {
	return nil, nil
}

func (r *objectMapRepository) Update(...*omap.OMap) error {
	return nil
}
