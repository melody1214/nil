package object

import (
	"database/sql"

	"github.com/chanyoung/nil/app/mds/repository"
)

// Repository provides access to object database.
type Repository interface {
	Execute(txid repository.TxID, query string) (sql.Result, error)
	QueryRow(txid repository.TxID, query string, args ...interface{}) *sql.Row
}
