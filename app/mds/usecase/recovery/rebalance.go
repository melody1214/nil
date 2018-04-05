package recovery

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sort"
)

func (h *handlers) needRebalance() bool {
	q := fmt.Sprintf(
		`
		SELECT
			chain,
			max_chain
		FROM
			volume
		`,
	)

	rows, err := h.store.Query(q)
	if err != nil {
		log.Error(err)
		return false
	}
	defer rows.Close()

	for rows.Next() {
		var chain, maxChain int
		if err = rows.Scan(&chain, &maxChain); err != nil {
			log.Error(err)
			return false
		}

		if isVolumeUnbalanced(chain, maxChain) {
			return true
		}
	}

	return false
}

func isVolumeUnbalanced(chain, maxChain int) bool {
	if maxChain == 0 {
		return false
	}

	if chain == 0 {
		return true
	}

	return (chain*100)/maxChain < 70
}

func (h *handlers) rebalance() error {
	speedLv := []string{"low", "mid", "high"}

	for _, speed := range speedLv {
		if err := h.rebalanceSpeedGroup(speed); err != nil {
			log.Error(err)
		}
	}

	return nil
}

func (h *handlers) rebalanceSpeedGroup(speed string) error {
	q := fmt.Sprintf(
		`
		SELECT
			volume_id,
			volume_status,
			node_id,
			used,
			chain,
			max_chain
		FROM
			volume
		WHERE
			speed = '%s'
		ORDER BY rand();
		`, speed,
	)

	rows, err := h.store.Query(q)
	if err != nil {
		return err
	}
	defer rows.Close()

	vols := make([]*Volume, 0)
	for rows.Next() {
		vol := &Volume{
			Unbalanced: false,
		}
		if err = rows.Scan(
			&vol.ID,
			&vol.Status,
			&vol.NodeID,
			&vol.Used,
			&vol.Chain,
			&vol.MaxChain,
		); err != nil {
			return err
		}

		if isVolumeUnbalanced(vol.Chain, vol.MaxChain) {
			vol.Unbalanced = true
		}
		vols = append(vols, vol)
	}

	for _, vol := range vols {
		if vol.Unbalanced {
			if err := h.doRebalance(vol, vols); err != nil {
				log.Error(err)
			}
		}
	}

	return nil
}

func (h *handlers) doRebalance(target *Volume, group []*Volume) error {
	const localChainNum = 4

	perm := rand.Perm(len(group))
	shuffledGroup := make([]*Volume, len(group))
	for i, v := range perm {
		shuffledGroup[v] = group[i]
	}

	if len(shuffledGroup) < localChainNum {
		return fmt.Errorf("lack of volumes for rebalancing: %+v", group)
	}
	sort.Sort(ByFreeChain(shuffledGroup))

	return h.newLocalChain(target, shuffledGroup)
}

func (h *handlers) newLocalChain(primary *Volume, vols []*Volume) error {
	const localChainNum = 4
	if len(vols) < localChainNum {
		return fmt.Errorf("lack of volumes for make local chain: %+v", vols)
	}

	selected := make([]*Volume, 0)
	selected = append(selected, primary)
	c := localChain{
		status:    "alive",
		parityVol: primary.ID,
		firstVol:  -1,
		secondVol: -1,
		thirdVol:  -1,
	}
	for _, v := range vols {
		if v.ID == primary.ID {
			continue
		}

		if v.Chain >= v.MaxChain {
			continue
		}

		if c.firstVol < 0 {
			sameNode := false
			for _, s := range selected {
				if v.NodeID == s.NodeID {
					sameNode = true
					break
				}
			}
			if sameNode {
				continue
			}

			selected = append(selected, v)
			c.firstVol = v.ID
			continue
		}

		if c.secondVol < 0 {
			sameNode := false
			for _, s := range selected {
				if v.NodeID == s.NodeID {
					sameNode = true
					break
				}
			}
			if sameNode {
				continue
			}

			selected = append(selected, v)
			c.secondVol = v.ID
			continue
		}

		if c.thirdVol < 0 {
			sameNode := false
			for _, s := range selected {
				if v.NodeID == s.NodeID {
					sameNode = true
					break
				}
			}
			if sameNode {
				continue
			}

			selected = append(selected, v)
			c.thirdVol = v.ID
			continue
		}

		break
	}

	if c.parityVol < 0 || c.firstVol < 0 || c.secondVol < 0 || c.thirdVol < 0 {
		return fmt.Errorf("not enough free volumes to make local chain")
	}

	q := fmt.Sprintf(
		`
		SELECT
			local_chain_id
		FROM
			local_chain
		WHERE
			first_volume_id = '%d' and
			second_volume_id = '%d' and
			third_volume_id = '%d' and
			parity_volume_id = '%d'
		`, c.firstVol, c.secondVol, c.thirdVol, c.parityVol,
	)

	var exist int
	err := h.store.QueryRow(q).Scan(&exist)
	if err == nil {
		return fmt.Errorf("already have same local chain: %+v", c)
	} else if err == sql.ErrNoRows {
		// There is no duplicate local chain.
	} else {
		return err
	}

	q = fmt.Sprintf(
		`
		INSERT INTO local_chain (local_chain_status, first_volume_id, second_volume_id, third_volume_id, parity_volume_id)
		VALUES ('%s', '%d', '%d', '%d', '%d')
		`, c.status, c.firstVol, c.secondVol, c.thirdVol, c.parityVol,
	)
	_, err = h.store.Execute(q)
	if err != nil {
		return err
	}

	log.Errorf("%+v", selected)
	for _, v := range selected {
		q := fmt.Sprintf(
			`
		UPDATE volume
		SET chain=chain+1
		WHERE volume_id in ('%d')
		`, v.ID,
		)

		_, err := h.store.Execute(q)
		if err != nil {
			log.Error(err)
		}

		// v.Chain++
		v.Chain = v.Chain + 1
	}

	log.Infof("create local chain %+v", c)

	return nil
}
