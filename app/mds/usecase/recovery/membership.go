package recovery

import (
	"database/sql"
	"fmt"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/swim"
	"github.com/chanyoung/nil/pkg/util/mlog"
)

func (h *handlers) updateMembership() {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.updateMembership")

	conn, err := nilrpc.Dial(h.cfg.ServerAddr+":"+h.cfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		ctxLogger.Error(err)
		return
	}
	defer conn.Close()

	req := &nilrpc.GetMembershipListRequest{}
	res := &nilrpc.GetMembershipListResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsMembershipGetMembershipList.String(), req, res); err != nil {
		ctxLogger.Error(err)
	}

	membership := res.Nodes
	for _, member := range membership {
		// Currently we only cares ds.
		if member.Type == swim.DS || member.Type == swim.MDS {
			h.doUpdateMembership(member)
		}
	}
}

func (h *handlers) doUpdateMembership(sm swim.Member) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.doUpdateMembership")

	q := fmt.Sprintf(
		`
		SELECT
			node_status,
			node_address
		FROM
			node
		WHERE
			node_name = '%s'
		`, sm.ID,
	)

	var oldStat, oldAddr string
	row := h.store.QueryRow(repository.NotTx, q)
	if row == nil {
		ctxLogger.Error("mysql is not connected yet")
		return
	}

	err := row.Scan(&oldStat, &oldAddr)
	if err == nil {
		// Member exists, compare if some fields are changed.
		if sm.Status.String() != oldStat || string(sm.Address) != oldAddr {
			h.updateMember(sm)
		}
	} else if err == sql.ErrNoRows {
		// Member not exists, add into the database.
		h.insertNewMember(sm)
	} else {
		ctxLogger.Error(err)
		return
	}
}

func (h *handlers) insertNewMember(sm swim.Member) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.insertNewMember")
	ctxLogger.Infof("insert a new member %v", sm)

	q := fmt.Sprintf(
		`
		INSERT INTO node (node_name, node_type, node_status, node_address)
		VALUES ('%s', '%s', '%s', '%s')
		`, string(sm.ID), sm.Type.String(), sm.Status.String(), string(sm.Address),
	)

	_, err := h.store.Execute(repository.NotTx, q)
	if err != nil {
		ctxLogger.Error(err)
	}
}

func (h *handlers) updateMember(sm swim.Member) {
	ctxLogger := mlog.GetMethodLogger(logger, "handlers.updateMember")
	ctxLogger.Infof("update a member %v", sm)

	q := fmt.Sprintf(
		`
		UPDATE node
		SET node_status='%s', node_address='%s'
		WHERE node_name in ('%s')
		`, sm.Status.String(), string(sm.Address), string(sm.ID),
	)

	_, err := h.store.Execute(repository.NotTx, q)
	if err != nil {
		ctxLogger.Error(err)
	}
}
