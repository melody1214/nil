package server

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/cmap"
)

func (s *Server) updateClusterMap() (*cmap.CMap, error) {
	// 1. Get new map version.
	ver, err := s.getNewClusterMapVer()
	if err != nil {
		return nil, err
	}

	// 2. Create a cluster map with the new version.
	return s.createClusterMap(ver)
}

func (s *Server) getNewClusterMapVer() (int64, error) {
	q := fmt.Sprintf(
		`
		INSERT INTO cmap (cmap_id)
		VALUES (NULL)
		`,
	)

	res, err := s.store.Execute(q)
	if err != nil {
		return -1, err
	}

	ver, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return ver, nil
}

func (s *Server) createClusterMap(ver int64) (*cmap.CMap, error) {
	q := fmt.Sprintf(
		`
		SELECT
            node_type,
			node_status,
			node_address
		FROM
			node
		`,
	)

	rows, err := s.store.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := &cmap.CMap{
		Version: ver,
		Nodes:   make([]cmap.Node, 0),
	}

	for rows.Next() {
		n := cmap.Node{}

		if err := rows.Scan(&n.Type, &n.Stat, &n.Addr); err != nil {
			log.Error(err)
		}

		m.Nodes = append(m.Nodes, n)
	}

	return m, nil
}
