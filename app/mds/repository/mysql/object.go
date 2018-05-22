package mysql

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/usecase/object"
)

type objectStore struct {
	*Store
}

// NewObjectRepository returns a new instance of a mysql object domain repository.
func NewObjectRepository(s *Store) object.Repository {
	return &objectStore{
		Store: s,
	}
}

func (s *objectStore) Put(o *object.ObjInfo) error {
	q := fmt.Sprintf(
		`
		INSERT INTO object (obj_name, obj_bucket, obj_encoding_group, obj_volume)
		SELECT '%s', b.bk_id, '%d', '%d'
		FROM bucket b
		WHERE bk_name = '%s'
		`, o.Name, o.EncGrp, o.Vol, o.Bucket,
	)

	_, err := s.Execute(repository.NotTx, q)
	return err
}

func (s *objectStore) Get(name string) (*object.ObjInfo, error) {
	o := &object.ObjInfo{}

	q := fmt.Sprintf(
		`
		SELECT
			obj_encoding_group, obj_volume
		FROM
			object
		WHERE
			obj_name = '%s'
		`, name,
	)

	row := s.QueryRow(repository.NotTx, q)
	if row == nil {
		return nil, fmt.Errorf("mysql not connected yet")
	}

	err := row.Scan(&o.EncGrp, &o.Vol)
	if err != nil {
		return nil, err
	}

	q = fmt.Sprintf(
		`
		SELECT
			vl_node
		FROM
			volume
		WHERE
			vl_id = '%d'
		`, o.Vol,
	)

	row = s.QueryRow(repository.NotTx, q)
	if row == nil {
		return nil, fmt.Errorf("mysql not connected yet")
	}

	err = row.Scan(&o.Node)
	if err != nil {
		return nil, err
	}

	return o, nil
}
