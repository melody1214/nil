package clustermap

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/mlog"
)

// TODO: with transaction.

func (h *handlers) updateClusterMap() error {
	// Get new map version.
	ver, err := h.getNewClusterMapVer()
	if err != nil {
		return err
	}

	// Create a cluster map with the new version.
	cm, err := h.createClusterMap(ver)
	if err != nil {
		return err
	}

	return h.cMap.Update(cmap.WithFile(cm))
}

func (h *handlers) getNewClusterMapVer() (cmap.Version, error) {
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

	return cmap.Version(ver), nil
}

func (h *handlers) createClusterMap(ver cmap.Version) (*cmap.CMap, error) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.createClusterMap")

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
			ctxLogger.Error(err)
		}

		cm.Nodes = append(cm.Nodes, n)
	}

	return cm, nil
}
