package mysql

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/chanyoung/nil/app/mds/domain/model/clustermap"
	"github.com/chanyoung/nil/app/mds/infrastructure/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/matrix"
	"github.com/chanyoung/nil/pkg/util/mlog"
	"github.com/pkg/errors"
)

type clusterMapRepository struct {
	s  *Store
	mu sync.RWMutex
}

// NewClusterMapRepository returns a new instance of a mysql cluster map repository.
func NewClusterMapRepository(s *Store) clustermap.Repository {
	return &clusterMapRepository{
		s: s,
	}
}

func (r *clusterMapRepository) UpdateWhole(m *cmap.CMap) (*cmap.CMap, error) {
	r.mu.Lock()

	tx, err := r.s.Begin()
	if err != nil {
		return nil, err
	}

	for _, n := range m.Nodes {
		q := fmt.Sprintf(
			`
            INSERT INTO node (node_id, node_name, node_type, node_status, node_address, node_size, node_encoding_matrix)
            VALUES (%d, '%s', '%s', '%s', '%s', '%d', '%d') ON DUPLICATE KEY UPDATE node_status='%s', node_size='%d', node_encoding_matrix='%d'
            `, n.ID.Int64(), n.Name.String(), n.Type.String(), n.Stat.String(), n.Addr.String(), n.Size, n.MatrixID, n.Stat.String(), n.Size, n.MatrixID,
		)

		if _, err := r.s.Execute(tx, q); err != nil {
			r.s.Rollback(tx)
			return nil, errors.Wrap(err, "failed to update cluster map")
		}
	}

	_, err = r.incrVersion(m.Time)
	if err != nil {
		r.s.Rollback(tx)
		return nil, errors.Wrap(err, "failed to update cluster map")
	}
	r.s.Commit(tx)

	// TODO: rebalancing here.

	r.mu.Unlock()

	return r.FindLatest()
}

func (r *clusterMapRepository) UpdateNode(n *cmap.Node) (*cmap.CMap, error) {
	r.mu.Lock()

	tx, err := r.s.Begin()
	if err != nil {
		return nil, err
	}

	q := fmt.Sprintf(
		`
        INSERT INTO node (node_id, node_name, node_type, node_status, node_address, node_size, node_encoding_matrix)
        VALUES (%d, '%s', '%s', '%s', '%s', '%d', '%d') ON DUPLICATE KEY UPDATE node_status='%s', node_size='%d', node_encoding_matrix='%d'
        `, n.ID.Int64(), n.Name.String(), n.Type.String(), n.Stat.String(), n.Addr.String(), n.Size, n.MatrixID, n.Stat.String(), n.Size, n.MatrixID,
	)

	if _, err := r.s.Execute(tx, q); err != nil {
		r.s.Rollback(tx)
		return nil, errors.Wrap(err, "failed to update cluster map")
	}

	_, err = r.incrVersion(cmap.Now())
	if err != nil {
		r.s.Rollback(tx)
		return nil, errors.Wrap(err, "failed to update cluster map")
	}
	r.s.Commit(tx)

	r.mu.Unlock()

	return r.FindLatest()
}

func (r *clusterMapRepository) FindLatest() (*cmap.CMap, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m := &cmap.CMap{}

	v, t, err := r.getVersionAndTime()
	if err != nil {
		return nil, err
	}
	m.Version = v
	m.Time = t

	ids, err := r.getEncodingMatricesID()
	if err != nil {
		return nil, err
	}
	m.MatrixIDs = ids

	q := fmt.Sprintf(
		`
        SELECT node_id, node_name, node_type, node_status, node_address, node_size, node_encoding_matrix
        FROM node
        `,
	)

	rows, err := r.s.Query(repository.NotTx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var n cmap.Node
		if err = rows.Scan(
			&n.ID, &n.Name, &n.Type, &n.Stat, &n.Addr, &n.Size, &n.MatrixID,
		); err != nil {
			return nil, err
		}

		m.Nodes = append(m.Nodes, n)
	}

	return m, nil
}

func (r *clusterMapRepository) incrVersion(t cmap.Time) (cmap.Version, error) {
	q := fmt.Sprintf(
		`
		INSERT INTO cmap (cmap_id, cmap_time)
		VALUES (NULL, '%s')
		`, t,
	)

	res, err := r.s.Execute(repository.NotTx, q)
	if err != nil {
		return cmap.Version(-1), err
	}

	ver, err := res.LastInsertId()
	if err != nil {
		return cmap.Version(-1), err
	}

	return cmap.Version(ver), nil
}

func (r *clusterMapRepository) getVersionAndTime() (cmap.Version, cmap.Time, error) {
	q := fmt.Sprintf(
		`
		SELECT cmap_id, cmap_time
        FROM cmap
        ORDER BY cmap_id DESC
		`,
	)

	var (
		v cmap.Version
		t cmap.Time
	)

	err := r.s.QueryRow(repository.NotTx, q).Scan(&v, &t)
	if err != nil {
		return cmap.Version(-1), cmap.Time(""), err
	}

	return v, t, nil
}

// InitEncodingMatricesID initializes encoding matrices id based on the region id.
func (r *clusterMapRepository) InitEncodingMatricesID() error {
	ctxLogger := mlog.GetMethodLogger(logger, "clusterMapRepository.InitEncodingMatricesID")

	q := fmt.Sprintf(
		`
		SELECT rg_id
		FROM region
		WHERE rg_name='%s'
		`, r.s.cfg.Raft.LocalClusterRegion,
	)

	var regionID int
	err := r.s.QueryRow(repository.NotTx, q).Scan(&regionID)
	if err != nil {
		ctxLogger.Error("failed to fetch region id")
		return err
	}

	localEncodingMatrices, _ := strconv.Atoi(r.s.cfg.LocalEncodingMatrices)
	startEncodingMatrixIndex := regionID * localEncodingMatrices

	tx, err := r.s.Begin()
	if err != nil {
		ctxLogger.Error("failed to start transaction")
		return err
	}

	for i := 0; i < localEncodingMatrices; i++ {
		m, err := matrix.FindEncodingMatrixByIndex(startEncodingMatrixIndex + i)
		if err != nil {
			r.s.Rollback(tx)
			ctxLogger.Errorf("failed to find cauchy matrix with the given index: %d", startEncodingMatrixIndex+i)
			return err
		}

		q := fmt.Sprintf(
			`
			INSERT INTO cmap_encoding_matrix (cem_id)
			VALUES (%d)
			`, m.ID.Byte(),
		)

		_, err = r.s.Execute(repository.NotTx, q)
		if err != nil {
			ctxLogger.Error("failed to insert encoding matrix id into the cmap_encoding_matrix")
			r.s.Rollback(tx)
			return err
		}
	}

	r.s.Commit(tx)

	return nil
}

func (r *clusterMapRepository) getEncodingMatricesID() (ids []int, err error) {
	q := fmt.Sprintf(
		`
		SELECT cem_id
        FROM cmap_encoding_matrix
        ORDER BY cem_id ASC
		`,
	)

	rows, err := r.s.Query(repository.NotTx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		if err = rows.Scan(&id); err != nil {
			return nil, err
		}

		ids = append(ids, id)
	}

	return ids, nil
}
