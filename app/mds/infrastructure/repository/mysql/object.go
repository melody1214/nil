package mysql

import (
	"fmt"
	"strconv"

	"github.com/chanyoung/nil/app/mds/application/object"
	"github.com/chanyoung/nil/app/mds/infrastructure/repository"
	"github.com/chanyoung/nil/pkg/cmap"
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
		SELECT
			egv_role
		FROM
			encoding_group_volume
		WHERE
			egv_volume=%d AND egv_encoding_group=%d
		`, o.Vol, o.EncGrp,
	)

	var role int
	if err := s.QueryRow(repository.NotTx, q).Scan(&role); err != nil {
		return err
	}

	q = fmt.Sprintf(
		`
		INSERT INTO object (obj_name, obj_bucket, obj_encoding_group, obj_role)
		SELECT '%s', b.bk_id, '%d', '%d'
		FROM bucket b
		WHERE bk_name = '%s'
		`, o.Name, o.EncGrp, role, o.Bucket,
	)

	_, err := s.Execute(repository.NotTx, q)
	return err
}

func (s *objectStore) Get(name string) (*object.ObjInfo, error) {
	o := &object.ObjInfo{}

	q := fmt.Sprintf(
		`
		SELECT
			obj_encoding_group, obj_role
		FROM
			object
		WHERE
			obj_name = '%s'
		`, name,
	)

	var role int
	row := s.QueryRow(repository.NotTx, q)
	if row == nil {
		return nil, fmt.Errorf("mysql not connected yet")
	}

	err := row.Scan(&o.EncGrp, &role)
	if err != nil {
		return nil, err
	}

	q = fmt.Sprintf(
		`
		SELECT
			egv_volume
		FROM
			encoding_group_volume
		WHERE
			egv_encoding_group=%d AND egv_role=%d
		`, o.EncGrp, role,
	)

	row = s.QueryRow(repository.NotTx, q)
	if row == nil {
		return nil, fmt.Errorf("mysql not connected yet")
	}

	err = row.Scan(&o.Vol)
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

func (s *objectStore) GetChunk(eg cmap.ID) (cID string, err error) {
	q := fmt.Sprintf(
		`
		INSERT INTO chunk (chk_encoding_group, chk_status)
		VALUES (%d, '%s')
		`, eg, "W",
	)
	r, err := s.Store.Execute(repository.NotTx, q)
	if err != nil {
		return "", err
	}
	id, err := r.LastInsertId()
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(id, 10), nil
}

func (s *objectStore) SetChunk(cID string, egID cmap.ID, status string) error {
	q := fmt.Sprintf(
		`
		UPDATE chunk
		SET chk_encoding_group=%d, chk_status='%s'
		WHERE chk_id=%s
		`, egID, status, cID,
	)
	_, err := s.Store.Execute(repository.NotTx, q)
	return err
}
