package mysql

import (
	"fmt"
	"sync"

	"github.com/chanyoung/nil/app/mds/domain/model/clustermap"
	"github.com/chanyoung/nil/app/mds/infrastructure/repository"
	"github.com/chanyoung/nil/pkg/cmap"
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
            INSERT INTO node (node_id, node_name, node_type, node_status, node_address)
            VALUES (%d, '%s', '%s', '%s', '%s') ON DUPLICATE KEY UPDATE node_status='%s'
            `, n.ID.Int64(), n.Name.String(), n.Type.String(), n.Stat.String(), n.Addr.String(), n.Stat.String(),
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
            INSERT INTO node (node_id, node_name, node_type, node_status, node_address)
            VALUES (%d, '%s', '%s', '%s', '%s') ON DUPLICATE KEY UPDATE node_status='%s'
            `, n.ID.Int64(), n.Name.String(), n.Type.String(), n.Stat.String(), n.Addr.String(), n.Stat.String(),
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

	// TODO: rebalancing here.

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

	q := fmt.Sprintf(
		`
        SELECT node_id, node_name, node_type, node_status, node_address
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
			&n.ID, &n.Name, &n.Type, &n.Stat, &n.Addr,
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
