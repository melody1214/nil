package admin

import (
	"database/sql"

	"github.com/chanyoung/nil/app/mds/repository"
)

// Repository provides access to admin database.
type Repository interface {
	Join(nodeID, addr string) error
	PublishCommand(op, query string) (result sql.Result, err error)
	Query(txid repository.TxID, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(txid repository.TxID, query string, args ...interface{}) *sql.Row
	Execute(txid repository.TxID, query string) (sql.Result, error)
}
