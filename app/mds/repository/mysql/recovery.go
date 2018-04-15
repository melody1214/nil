package mysql

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/usecase/recovery"
	"github.com/pkg/errors"
)

type recoveryStore struct {
	*Store
}

// NewRecoveryRepository returns a new instance of a mysql recovery repository.
func NewRecoveryRepository(s *Store) recovery.Repository {
	return &recoveryStore{
		Store: s,
	}
}

func (s *recoveryStore) FindAllVolumes(txid repository.TxID) ([]*recovery.Volume, error) {
	q := fmt.Sprintf(
		`
		SELECT
			vl_id,
			vl_status,
			vl_node,
			vl_used,
			vl_encoding_group,
			vl_max_encoding_group,
			vl_speed
		FROM
			volume
		`,
	)

	rows, err := s.Query(txid, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var vols []*recovery.Volume
	for rows.Next() {
		v := &recovery.Volume{}

		if err = rows.Scan(&v.ID, &v.Status, &v.NodeID, &v.Used, &v.Chain, &v.MaxChain, &v.Speed); err != nil {
			return nil, err
		}

		vols = append(vols, v)
	}

	return vols, nil
}

func (s recoveryStore) MakeNewEncodingGroup(txid repository.TxID, encGrp *recovery.EncodingGroup) error {
	// Make new encoding group.
	q := fmt.Sprintf(
		`
		INSERT INTO encoding_group (eg_status)
		VALUES ('%s')
		`, encGrp.Status.String(),
	)
	r, err := s.Store.Execute(txid, q)
	if err != nil {
		return errors.Wrap(err, "failed to create encoding group")
	}
	egID, err := r.LastInsertId()
	if err != nil {
		return errors.Wrap(err, "failed to create encoding group")
	}

	// Register each volumes.
	for role, v := range encGrp.Vols {
		q = fmt.Sprintf(
			`
			INSERT INTO encoding_group_volume (egv_encoding_group, egv_volume, egv_role)
			VALUES ('%d', '%d', '%d')
			`, egID, v.Int64(), role,
		)
		_, err = s.Store.Execute(txid, q)
		if err != nil {
			return errors.Wrap(err, "failed to create volume in encoding group table")
		}

		q = fmt.Sprintf(
			`
			UPDATE volume
			SET vl_encoding_group=vl_encoding_group+1
			WHERE vl_id in ('%d')
			`, v.Int64(),
		)
		_, err = s.Store.Execute(txid, q)
		if err != nil {
			return errors.Wrap(err, "failed to increase encoding group counting in volume")
		}
	}

	return nil
}
