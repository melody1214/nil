package mysql

import (
	"database/sql"
	"fmt"

	"github.com/chanyoung/nil/app/mds/domain/model/region"
	"github.com/chanyoung/nil/app/mds/infrastructure/repository"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
)

type regionRepository struct {
	s *Store
}

// NewRegionRepository returns a new instance of a mysql region repository.
func NewRegionRepository(s *Store) region.Repository {
	return &regionRepository{
		s: s,
	}
}

func (r *regionRepository) FindByID(id region.ID) (*region.Region, error) {
	ctxLogger := mlog.GetMethodLogger(logger, "regionRepository.FindByID")

	q := fmt.Sprintf(
		`
		SELECT
			rg_id, rg_name, rg_end_point
		FROM
			region
		WHERE
			rg_id='%s'
		`, id.String(),
	)

	rg := &region.Region{}
	err := r.s.QueryRow(repository.NotTx, q).Scan(&rg.ID, &rg.Name, &rg.EndPoint)
	if err == sql.ErrNoRows {
		err = region.ErrNotExist
	} else if err != nil {
		ctxLogger.Error(errors.Wrapf(err, "failed to find user by ID: %s", id.String()))
		err = region.ErrInternal
	}

	return rg, err
}

func (r *regionRepository) FindByName(name region.Name) (*region.Region, error) {
	ctxLogger := mlog.GetMethodLogger(logger, "regionRepository.FindByID")

	q := fmt.Sprintf(
		`
		SELECT
			rg_id, rg_name, rg_end_point
		FROM
			region
		WHERE
			rg_name='%s'
		`, name.String(),
	)

	rg := &region.Region{}
	err := r.s.QueryRow(repository.NotTx, q).Scan(&rg.ID, &rg.Name, &rg.EndPoint)
	if err == sql.ErrNoRows {
		err = region.ErrNotExist
	} else if err != nil {
		ctxLogger.Error(errors.Wrapf(err, "failed to find user by name: %s", name.String()))
		err = region.ErrInternal
	}

	return rg, err
}

func (r *regionRepository) Create(rg *region.Region) error {
	ctxLogger := mlog.GetMethodLogger(logger, "regionRepository.Create")

	q := fmt.Sprintf(
		`
		INSERT INTO region (rg_name, rg_end_point)
		SELECT * FROM (SELECT '%s' AS rn, '%s' AS ep) AS tmp
		WHERE NOT EXISTS (
			SELECT rg_name FROM region WHERE rg_name='%s'
		) LIMIT 1;
		`, rg.Name.String(), rg.EndPoint.String(), rg.Name.String(),
	)

	_, err := r.s.PublishCommand("execute", q)
	if err != nil {
		ctxLogger.Error(errors.Wrap(err, "failed to publish command"))
		err = region.ErrInternal
	}

	return err
}
