package recovery

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/cmap"
)

func (h *handlers) updateClusterMap() (*cmap.CMap, error) {
	// 1. Get new map version.
	ver, err := h.getNewClusterMapVer()
	if err != nil {
		return nil, err
	}

	// 2. Create a cluster map with the new version.
	return h.createClusterMap(ver)
}

func (h *handlers) getNewClusterMapVer() (int64, error) {
	q := fmt.Sprintf(
		`
		INSERT INTO cmap (cmap_id)
		VALUES (NULL)
		`,
	)

	res, err := h.store.Execute(q)
	if err != nil {
		return -1, err
	}

	ver, err := res.LastInsertId()
	if err != nil {
		return -1, err
	}

	return ver, nil
}

func (h *handlers) createClusterMap(ver int64) (*cmap.CMap, error) {
	q := fmt.Sprintf(
		`
		SELECT
			node_id,
			node_name,
            node_type,
			node_status,
			node_address
		FROM
			node
		`,
	)

	rows, err := h.store.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cm := &cmap.CMap{
		Version: ver,
		Nodes:   make([]cmap.Node, 0),
	}

	for rows.Next() {
		n := cmap.Node{}

		if err := rows.Scan(&n.ID, &n.Name, &n.Type, &n.Stat, &n.Addr); err != nil {
			log.Error(err)
		}

		cm.Nodes = append(cm.Nodes, n)
	}

	return cm, nil
}
