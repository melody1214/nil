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
