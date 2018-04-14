package mysql

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/usecase/recovery"
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
