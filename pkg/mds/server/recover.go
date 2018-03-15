package server

import (
	"database/sql"
	"fmt"

	"github.com/chanyoung/nil/pkg/cmap"
	"github.com/chanyoung/nil/pkg/swim"
	"github.com/sirupsen/logrus"
)

func (s *Server) recover(pe swim.PingError) {
	// Logging the error.
	log.WithFields(logrus.Fields{
		"server":       "swim",
		"message type": pe.Type,
		"destID":       pe.DestID,
	}).Warn(pe.Err)

	// Updates membership.
	s.updateMembership()
	m, err := cmap.GetLatest(s.cfg.ServerAddr + ":" + s.cfg.ServerPort)
	if err != nil {
		log.Error(err)
	}
	log.Error(m)

	// If the error message is occured because just simple membership
	// changed, then finish the recover routine here.
	if pe.Err == swim.ErrChanged {
		return
	}

	// TODO: recovery routine.
}

func (s *Server) updateMembership() {
	membership := s.swimSrv.GetMap()
	for _, m := range membership {
		// Currently we only cares ds.
		if m.Type == swim.DS || m.Type == swim.MDS {
			s.doUpdateMembership(m)
		}
	}
}

func (s *Server) doUpdateMembership(m swim.Member) {
	q := fmt.Sprintf(
		`
		SELECT
			node_status,
			node_address
		FROM
			node
		WHERE
			node_name = '%s'
		`, m.ID,
	)

	var oldStat, oldAddr string
	err := s.store.QueryRow(q).Scan(&oldStat, &oldAddr)
	if err == nil {
		// Member exists, compare if some fields are changed.
		if m.Status.String() != oldStat || string(m.Address) != oldAddr {
			s.updateMember(m)
		}
	} else if err == sql.ErrNoRows {
		// Member not exists, add into the database.
		s.insertNewMember(m)
	} else {
		log.Error(err)
		return
	}
}

func (s *Server) insertNewMember(m swim.Member) {
	log.Infof("insert a new member %v", m)

	q := fmt.Sprintf(
		`
		INSERT INTO node (node_name, node_type, node_status, node_address)
		VALUES ('%s', '%s', '%s', '%s')
		`, string(m.ID), m.Type.String(), m.Status.String(), string(m.Address),
	)

	_, err := s.store.Execute(q)
	if err != nil {
		log.Error(err)
	}
}

func (s *Server) updateMember(m swim.Member) {
	log.Infof("update a member %v", m)

	q := fmt.Sprintf(
		`
		UPDATE node
		SET node_status='%s', node_address='%s'
		WHERE node_name in ('%s')
		`, m.Status.String(), string(m.Address), string(m.ID),
	)

	_, err := s.store.Execute(q)
	if err != nil {
		log.Error(err)
	}
}
