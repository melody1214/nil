package repository

import (
	"database/sql"

	"github.com/chanyoung/nil/pkg/nilmux"
)

// Store is a persistent data store for mds.
type Store interface {
	Join(nodeID, addr string) error
	PublishCommand(op, query string) (result sql.Result, err error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Execute(query string) (sql.Result, error)
	Open(raftL *nilmux.Layer) error
	Close() error
}
