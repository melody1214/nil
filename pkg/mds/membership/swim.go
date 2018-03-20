package membership

import (
	"database/sql"
	"fmt"

	"github.com/chanyoung/nil/pkg/swim"
	"github.com/sirupsen/logrus"
)

func (h *Handler) Recover(pe swim.PingError) {
	// 1. Logging the error.
	log.WithFields(logrus.Fields{
		"server":       "swim",
		"message type": pe.Type,
		"destID":       pe.DestID,
	}).Warn(pe.Err)

	// 2. Updates membership.
	h.updateMembership()

	// 3. Get the new version of cluster map.
	newCMap, err := h.updateClusterMap()
	if err != nil {
		log.Error(err)
	}

	// 4. Save the new cluster map.
	err = newCMap.Save()
	if err != nil {
		log.Error(err)
	}

	// 5. If the error message is occured because just simple membership
	// changed, then finish the recover routine here.
	if pe.Err == swim.ErrChanged {
		return
	}

	// TODO: recovery routine.
}

func (h *Handler) updateMembership() {
	membership := h.swimSrv.GetMap()
	for _, member := range membership {
		// Currently we only cares ds.
		if member.Type == swim.DS || member.Type == swim.MDS {
			h.doUpdateMembership(member)
		}
	}
}

func (h *Handler) doUpdateMembership(sm swim.Member) {
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

func (h *Handler) insertNewMember(sm swim.Member) {
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

func (h *Handler) updateMember(sm swim.Member) {
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
