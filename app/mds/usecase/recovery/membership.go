package recovery

import (
	"log"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
)

// func (w *worker) updateMembership() {
// 	ctxLogger := mlog.GetMethodLogger(logger, "worker.updateMembership")

// 	conn, err := nilrpc.Dial(w.cfg.ServerAddr+":"+w.cfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
// 	if err != nil {
// 		ctxLogger.Error(err)
// 		return
// 	}
// 	defer conn.Close()

// 	req := &nilrpc.GetMembershipListRequest{}
// 	res := &nilrpc.GetMembershipListResponse{}

// 	cli := rpc.NewClient(conn)
// 	if err := cli.Call(nilrpc.MdsMembershipGetMembershipList.String(), req, res); err != nil {
// 		ctxLogger.Error(err)
// 	}

// 	membership := res.Nodes
// 	for _, member := range membership {
// 		// Currently we only cares ds.
// 		if member.Type == swim.DS || member.Type == swim.MDS {
// 			w.doUpdateMembership(member)
// 		}
// 	}
// }

// func (w *worker) doUpdateMembership(sm swim.Member) {
// 	ctxLogger := mlog.GetMethodLogger(logger, "worker.doUpdateMembership")

// 	q := fmt.Sprintf(
// 		`
// 		SELECT
// 			node_status,
// 			node_address
// 		FROM
// 			node
// 		WHERE
// 			node_name = '%s'
// 		`, sm.ID,
// 	)

// 	var oldStat, oldAddr string
// 	row := w.store.QueryRow(repository.NotTx, q)
// 	if row == nil {
// 		ctxLogger.Error("mysql is not connected yet")
// 		return
// 	}

// 	err := row.Scan(&oldStat, &oldAddr)
// 	if err == nil {
// 		// Member exists, compare if some fields are changed.
// 		if sm.Status.String() != oldStat || string(sm.Address) != oldAddr {
// 			w.updateMember(sm)
// 		}
// 	} else if err == sql.ErrNoRows {
// 		// Member not exists, add into the database.
// 		w.insertNewMember(sm)
// 	} else {
// 		ctxLogger.Error(err)
// 		return
// 	}
// }

// func (w *worker) insertNewMember(sm swim.Member) {
// 	ctxLogger := mlog.GetMethodLogger(logger, "worker.insertNewMember")
// 	ctxLogger.Infof("insert a new member %v", sm)

// 	q := fmt.Sprintf(
// 		`
// 		INSERT INTO node (node_name, node_type, node_status, node_address)
// 		VALUES ('%s', '%s', '%s', '%s')
// 		`, string(sm.ID), sm.Type.String(), sm.Status.String(), string(sm.Address),
// 	)

// 	_, err := w.store.Execute(repository.NotTx, q)
// 	if err != nil {
// 		ctxLogger.Error(err)
// 	}
// }

// func (w *worker) updateMember(sm swim.Member) {
// 	ctxLogger := mlog.GetMethodLogger(logger, "worker.updateMember")
// 	ctxLogger.Infof("update a member %v", sm)

// 	q := fmt.Sprintf(
// 		`
// 		UPDATE node
// 		SET node_status='%s', node_address='%s'
// 		WHERE node_name in ('%s')
// 		`, sm.Status.String(), string(sm.Address), string(sm.ID),
// 	)

// 	_, err := w.store.Execute(repository.NotTx, q)
// 	if err != nil {
// 		ctxLogger.Error(err)
// 	}
// }

func (w *recoveryWorker) updateClusterMap() error {
	conn, err := nilrpc.Dial(w.cfg.ServerAddr+":"+w.cfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	req := &nilrpc.MCLUpdateClusterMapRequest{}
	res := &nilrpc.MCLUpdateClusterMapResponse{}

	cli := rpc.NewClient(conn)
	defer cli.Close()

	return cli.Call(nilrpc.MdsClusterUpdateClusterMap.String(), req, res)
}
