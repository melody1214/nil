package object

// type encodeGroup struct {
// 	cmap.EncodingGroup
// 	nodeIDs   []int64
// 	nodeAddrs []string
// }

// func (e *endec) updateGroup() {
// 	ctxLogger := mlog.GetMethodLogger(logger, "endec.updateGroup")

// 	mds, err := e.cMap.SearchCallNode().Type(cmap.MDS).Do()
// 	if err != nil {
// 		ctxLogger.Error(err)
// 		return
// 	}

// 	conn, err := nilrpc.Dial(mds.Addr, nilrpc.RPCNil, time.Duration(2*time.Second))
// 	if err != nil {
// 		ctxLogger.Error(err)
// 		return
// 	}
// 	defer conn.Close()

// 	vreq := &nilrpc.GetAllVolumeRequest{}
// 	vres := &nilrpc.GetAllVolumeResponse{}

// 	cli := rpc.NewClient(conn)
// 	if err := cli.Call(nilrpc.MdsAdminGetAllVolume.String(), vreq, vres); err != nil {
// 		ctxLogger.Error(err)
// 		return
// 	}

// 	volumeMap := make(map[int64]int64)
// 	for _, v := range vres.Volumes {
// 		volumeMap[v.ID] = v.NodeID
// 	}

// 	creq := &nilrpc.GetAllChainRequest{}
// 	cres := &nilrpc.GetAllChainResponse{}

// 	if err := cli.Call(nilrpc.MdsAdminGetAllChain.String(), creq, cres); err != nil {
// 		ctxLogger.Error(err)
// 		return
// 	}

// 	for _, eg := range cres.EncGrps {
// 		newEg := encodeGroup{
// 			EncodingGroup: eg,
// 			nodeIDs:       make([]int64, len(eg.Vols)),
// 			nodeAddrs:     make([]string, len(eg.Vols)),
// 		}

// 		for i, v := range eg.Vols {
// 			id, ok := volumeMap[v.Int64()]
// 			if ok == false {
// 				ctxLogger.Errorf("no such volume %d", v.Int64())
// 				continue
// 			}
// 			n, err := e.cMap.SearchCallNode().ID(cmap.ID(id)).Do()
// 			if err != nil {
// 				ctxLogger.Error(err)
// 				continue
// 			}
// 			newEg.nodeIDs[i] = id
// 			newEg.nodeAddrs[i] = n.Addr
// 		}

// 		e.emap[newEg.ID.String()] = newEg
// 	}
// }
