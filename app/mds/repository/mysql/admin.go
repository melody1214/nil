package mysql

import (
	"fmt"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/app/mds/usecase/admin"
	"github.com/chanyoung/nil/pkg/cmap"
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

func (s *adminStore) GetAllEncodingGroups(txid repository.TxID) ([]cmap.EncodingGroup, error) {
	q := fmt.Sprintf(
		`
		SELECT
			eg_id
		FROM
			encoding_group
		`,
	)

	rows, err := s.Query(txid, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	egs := make([]cmap.EncodingGroup, 0)
	for rows.Next() {
		eg := cmap.EncodingGroup{
			Vols: make([]cmap.ID, 0),
		}

		if err := rows.Scan(&eg.ID); err != nil {
			return nil, err
		}

		q = fmt.Sprintf(
			`
			SELECT
				egv_volume, egv_role
			FROM
				encoding_group_volume
			WHERE
				egv_encoding_group = '%d'
			ORDER BY
				egv_role
			`, eg.ID.Int64(),
		)

		vrows, err := s.Query(txid, q)
		if err != nil {
			return nil, err
		}

		for vrows.Next() {
			var volID cmap.ID
			var volRole int
			if err := vrows.Scan(&volID, &volRole); err != nil {
				return nil, err
			}

			eg.Vols = append(eg.Vols, volID)
		}
		vrows.Close()

		egs = append(egs, eg)
	}

	return egs, nil
}
