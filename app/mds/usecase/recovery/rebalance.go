package recovery

// func (w *recoveryWorker) needRebalance(vols []*Volume) bool {
// 	for _, v := range vols {
// 		if v.isUnbalanced() {
// 			return true
// 		}
// 	}

// 	return false
// }

// func isVolumeUnbalanced(chain, maxChain int) bool {
// 	if maxChain == 0 {
// 		return false
// 	}

// 	if chain == 0 {
// 		return true
// 	}

// 	return (chain*100)/maxChain < 70
// }

// func (w *recoveryWorker) rebalanceWithinSameVolumeSpeedGroup(vols []*Volume) error {
// 	ctxLogger := mlog.GetMethodLogger(logger, "worker.rebalance")

// 	rebalanced := false
// 	speedLv := []cmap.VolumeSpeed{cmap.Low, cmap.Mid, cmap.High}
// 	for _, speed := range speedLv {
// 		sVols := make([]*Volume, 0)
// 		for _, v := range vols {
// 			if v.Speed != speed {
// 				continue
// 			}

// 			sVols = append(sVols, v)
// 		}

// 		if err := w.rebalanceVolumeGroup(sVols); err != nil {
// 			ctxLogger.Error(err)
// 		} else {
// 			// rebalanced is set when any volume group is rebalanced.
// 			rebalanced = true
// 		}
// 	}

// 	if rebalanced == false {
// 		return fmt.Errorf("there is no rebalanceable volume group")
// 	}
// 	return nil
// }

// func (w *recoveryWorker) rebalanceVolumeGroup(vols []*Volume) error {
// 	doRebalance := false
// 	for _, v := range vols {
// 		if v.isUnbalanced() == false {
// 			continue
// 		}

// 		if err := w.doRebalance(v, vols); err != nil {
// 			return err
// 		}
// 		doRebalance = true
// 	}

// 	if doRebalance == false {
// 		return fmt.Errorf("there is no rebalanceable volume set")
// 	}

// 	return nil
// }

// func (w *recoveryWorker) doRebalance(target *Volume, group []*Volume) error {
// 	perm := rand.Perm(len(group))
// 	shuffledGroup := make([]*Volume, len(group))
// 	for i, v := range perm {
// 		shuffledGroup[v] = group[i]
// 	}

// 	shards, err := strconv.Atoi(w.cfg.LocalParityShards)
// 	if err != nil {
// 		return errors.Wrap(err, "failed to parse local parity shards number")
// 	}

// 	// N shards + 1 xoring result.
// 	if len(shuffledGroup) < shards+1 {
// 		return fmt.Errorf("lack of volumes for rebalancing: %+v", group)
// 	}
// 	sort.Sort(ByFreeChain(shuffledGroup))

// 	return w.newEncodingGroup(target, shuffledGroup, shards)
// }

// func (w *recoveryWorker) pickOneNewEncodingGroupVolume(picked []*Volume, candidates []*Volume) (*Volume, error) {
// 	if cap(picked) == len(picked) {
// 		return nil, fmt.Errorf("selected encoding group is full")
// 	}

// 	for _, c := range candidates {
// 		if isPicked(c, picked) {
// 			continue
// 		}

// 		if c.Chain > c.MaxChain {
// 			continue
// 		}

// 		return c, nil
// 	}

// 	return nil, fmt.Errorf("no available volume in the candidates")
// }

// func isPicked(target *Volume, picked []*Volume) bool {
// 	for _, v := range picked {
// 		if v == nil {
// 			return false
// 		}

// 		if v.ID == target.ID {
// 			return true
// 		}
// 	}
// 	return false
// }

// func (w *recoveryWorker) newEncodingGroup(primary *Volume, vols []*Volume, shards int) error {
// 	ctxLogger := mlog.GetMethodLogger(logger, "worker.newEncodingGroup")

// 	// Pick volumes from candidates.
// 	picked := make([]*Volume, 0, shards+1)
// 	picked = append(picked, primary)
// 	for i := 0; i < shards; i++ {
// 		p, err := w.pickOneNewEncodingGroupVolume(picked, vols)
// 		if err != nil {
// 			return errors.Wrap(err, "failed to pick new volume")
// 		}

// 		picked = append(picked, p)
// 	}

// 	// Make encoding group.
// 	eg := cmap.EncodingGroup{
// 		Stat: cmap.EGAlive,
// 		Vols: make([]cmap.ID, shards+1),
// 	}
// 	for i, p := range picked {
// 		eg.Vols[i] = p.ID
// 	}

// 	// TODO: prevent duplicated encoding group.

// 	// Update repository.
// 	txid, err := w.store.Begin()
// 	if err != nil {
// 		return errors.Wrap(err, "failed to start transaction")
// 	}
// 	err = w.store.MakeNewEncodingGroup(txid, &eg)
// 	if err != nil {
// 		w.store.Rollback(txid)
// 		return errors.Wrap(err, "failed to make new encoding group")
// 	}
// 	if err := w.store.Commit(txid); err != nil {
// 		w.store.Rollback(txid)
// 		return errors.Wrap(err, "failed to commit transaction")
// 	}

// 	ctxLogger.Infof("create encoding group %+v", eg)
// 	return nil
// }
