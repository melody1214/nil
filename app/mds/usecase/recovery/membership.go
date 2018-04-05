package recovery

import (
	"database/sql"
	"fmt"
	"net/rpc"
	"time"

	"github.com/chanyoung/nil/pkg/nilrpc"
	"github.com/chanyoung/nil/pkg/swim"
)

func (h *handlers) updateMembership() {
	conn, err := nilrpc.Dial(h.cfg.ServerAddr+":"+h.cfg.ServerPort, nilrpc.RPCNil, time.Duration(2*time.Second))
	if err != nil {
		log.Error(err)
		return
	}
	defer conn.Close()

	req := &nilrpc.GetMembershipListRequest{}
	res := &nilrpc.GetMembershipListResponse{}

	cli := rpc.NewClient(conn)
	if err := cli.Call(nilrpc.MdsMembershipGetMembershipList.String(), req, res); err != nil {
		log.Error(err)
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
	row := h.store.QueryRow(q)
	if row == nil {
		log.WithField("func", "doUpdateMembership").Error("mysql is not connected yet")
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
		log.Error(err)
		return
	}
}

func (h *handlers) insertNewMember(sm swim.Member) {
	log.Infof("insert a new member %v", sm)

	q := fmt.Sprintf(
		`
		INSERT INTO node (node_name, node_type, node_status, node_address)
		VALUES ('%s', '%s', '%s', '%s')
		`, string(sm.ID), sm.Type.String(), sm.Status.String(), string(sm.Address),
	)

	_, err := h.store.Execute(q)
	if err != nil {
		log.WithField("func", "insertNewMember").Error(err)
	}
}

func (h *handlers) updateMember(sm swim.Member) {
	log.Infof("update a member %v", sm)

	q := fmt.Sprintf(
		`
		UPDATE node
		SET node_status='%s', node_address='%s'
		WHERE node_name in ('%s')
		`, sm.Status.String(), string(sm.Address), string(sm.ID),
	)

	_, err := h.store.Execute(q)
	if err != nil {
		log.WithField("func", "updateMember").Error(err)
	}
}
