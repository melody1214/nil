package mysql

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chanyoung/nil/app/mds/infrastructure/repository"
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
func (s *Store) QueryRow(txid repository.TxID, query string, args ...interface{}) *sql.Row {
	if s.db == nil {
		return nil
	}
	return s.db.queryRow(txid, query, args...)
}

// Query executes a query that returns rows.
func (s *Store) Query(txid repository.TxID, query string, args ...interface{}) (*sql.Rows, error) {
	if s.db == nil {
		return nil, fmt.Errorf("mysql is not connected yet")
	}
	return s.db.query(txid, query, args...)
}

// Execute executes a query in the local cluster.
func (s *Store) Execute(txid repository.TxID, query string) (sql.Result, error) {
	if s.db == nil {
		return nil, fmt.Errorf("mysql is not connected yet")
	}
	return s.db.execute(txid, query)
}

func (s *Store) addRegion(region, addr string) error {
	q := fmt.Sprintf(
		`
		INSERT INTO region (rg_name, rg_end_point)
		SELECT * FROM (SELECT '%s' AS rn, '%s' AS ep) AS tmp
		WHERE NOT EXISTS (
			SELECT rg_name FROM region WHERE rg_name = '%s'
		) LIMIT 1;
		`, region, addr, region,
	)
	_, err := s.PublishCommand("execute", q)

	return err
}

func (s *Store) setGlobalClusterConf() error {
	q := fmt.Sprintf(
		`
		INSERT INTO cluster (cl_local_parity_shards)
		VALUES ('%s')
		`, s.cfg.LocalParityShards,
	)
	_, err := s.PublishCommand("execute", q)

	return err
}
