package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hashicorp/raft"
	"github.com/pkg/errors"
)

type command struct {
	Op    string `json:"op,omitempty"`
	Query string `json:"query,omitempty"`
}

// PublishCommand publish a command across the cluster.
func (s *Store) PublishCommand(op, query string) (result sql.Result, err error) {
	if s.raft.State() != raft.Leader {
		return nil, errors.New("not leader")
	}

	b, err := json.Marshal(&command{
		Op:    op,
		Query: query,
	})
	if err != nil {
		return nil, err
	}

	f := s.raft.Apply(b, 3*time.Second)
	if f.Error() != nil {
		return nil, f.Error()
	}

	r := f.Response().(*fsmExecuteResponse)
	return r.result, r.err
}

// QueryRow executes a query that is expected to return at most one row.
func (s *Store) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.db.QueryRow(query, args...)
}

func (s *Store) addRegion(region, addr string) error {
	q := fmt.Sprintf(
		`
		INSERT INTO region (region_name, end_point)
		SELECT * FROM (SELECT '%s' AS rn, '%s' AS ep) AS tmp
		WHERE NOT EXISTS (
			SELECT region_name FROM region WHERE region_name = '%s'
		) LIMIT 1;
		`, region, addr, region,
	)
	_, err := s.PublishCommand("execute", q)

	return err
}
