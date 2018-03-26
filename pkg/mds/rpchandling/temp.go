package rpchandling

import (
	"fmt"

	"github.com/chanyoung/nil/pkg/nilrpc"
)

// GetLocalChain : Test code. Will be removed soon.
func (h *Handler) GetLocalChain(req *nilrpc.GetLocalChainRequest, res *nilrpc.GetLocalChainResponse) error {
	q := fmt.Sprintf(
		`
		SELECT
            local_chain_id, parity_volume_id
		FROM
            local_chain
        ORDER BY rand() limit 1;
		`,
	)

	row := h.store.QueryRow(q)
	if row == nil {
		return fmt.Errorf("mysql not connected yet")
	}

	err := row.Scan(&res.LocalChainID, &res.ParityVolumeID)
	if err != nil {
		return err
	}

	q = fmt.Sprintf(
		`
		SELECT
			node_id
			FROM
			volume
 		WHERE
			volume_id = '%d'
		`, res.ParityVolumeID,
	)

	row = h.store.QueryRow(q)
	if row == nil {
		return fmt.Errorf("mysql not connected yet")
	}

	return row.Scan(&res.ParityNodeID)
}

func (h *Handler) GetAllChain(req *nilrpc.GetAllChainRequest, res *nilrpc.GetAllChainResponse) error {
	q := fmt.Sprintf(
		`
		SELECT
			local_chain_id, first_volume_id, second_volume_id, third_volume_id, parity_volume_id
		FROM
			local_chain
		`,
	)

	rows, err := h.store.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		c := nilrpc.Chain{}

		if err := rows.Scan(&c.ID, &c.FirstVolumeID, &c.SecondVolumeID, &c.ThirdVolumeID, &c.ParityVolumeID); err != nil {
			return err
		}

		res.Chains = append(res.Chains, c)
	}

	return nil
}

func (h *Handler) GetAllVolume(req *nilrpc.GetAllVolumeRequest, res *nilrpc.GetAllVolumeResponse) error {
	q := fmt.Sprintf(
		`
		SELECT
			volume_id, node_id
		FROM
			volume
		`,
	)

	rows, err := h.store.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		v := nilrpc.Volume{}

		if err := rows.Scan(&v.ID, &v.NodeID); err != nil {
			return err
		}

		res.Volumes = append(res.Volumes, v)
	}

	return nil
}
