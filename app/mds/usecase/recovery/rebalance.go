package recovery

import (
	"database/sql"
	"fmt"
	"math/rand"
	"sort"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/util/mlog"
)

func (h *handlers) needRebalance(vols []*Volume) bool {
	for _, v := range vols {
		if v.isUnbalanced() {
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

func (h *handlers) rebalance(vols []*Volume) error {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.rebalance")

	speedLv := []cmap.VolumeSpeed{cmap.Low, cmap.Mid, cmap.High}
	for _, speed := range speedLv {
		sVols := make([]*Volume, 0)
		for _, v := range vols {
			if v.Speed != speed {
				continue
			}

			sVols = append(sVols, v)
		}

		for _, v := range sVols {
			if v.isUnbalanced() == false {
				continue
			}

			if err := h.doRebalance(v, sVols); err != nil {
				ctxLogger.Error(err)
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
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.newLocalChain")

	const localChainNum = 4
	if len(vols) < localChainNum {
		return fmt.Errorf("lack of volumes for make local chain: %+v", vols)
	}

	selected := make([]*Volume, 0)
	selected = append(selected, primary)
	c := localChain{
		EncodingGroup: cmap.EncodingGroup{
			Status: cmap.EGAlive,
		},
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
			eg_id
		FROM
			encoding_group
		WHERE
			eg_first_volume = '%d' and
			eg_second_volume = '%d' and
			eg_third_volume = '%d' and
			eg_parity_volume = '%d'
		`, c.firstVol, c.secondVol, c.thirdVol, c.parityVol,
	)

	var exist int
	err := h.store.QueryRow(repository.NotTx, q).Scan(&exist)
	if err == nil {
		return fmt.Errorf("already have same local chain: %+v", c)
	} else if err == sql.ErrNoRows {
		// There is no duplicate local chain.
	} else {
		return err
	}

	q = fmt.Sprintf(
		`
		INSERT INTO encoding_group (eg_status, eg_first_volume, eg_second_volume, eg_third_volume, eg_parity_volume)
		VALUES ('%s', '%d', '%d', '%d', '%d')
		`, c.Status.String(), c.firstVol, c.secondVol, c.thirdVol, c.parityVol,
	)
	_, err = h.store.Execute(repository.NotTx, q)
	if err != nil {
		return err
	}

	ctxLogger.Errorf("%+v", selected)
	for _, v := range selected {
		q := fmt.Sprintf(
			`
		UPDATE volume
		SET vl_encoding_group=vl_encoding_group+1
		WHERE vl_id in ('%d')
		`, v.ID,
		)

		_, err := h.store.Execute(repository.NotTx, q)
		if err != nil {
			ctxLogger.Error(err)
		}

		// v.Chain++
		v.Chain = v.Chain + 1
	}

	ctxLogger.Infof("create local chain %+v", c)

	return nil
}
