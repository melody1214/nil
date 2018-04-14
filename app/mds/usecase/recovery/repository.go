package recovery

import (
	"database/sql"

	"github.com/chanyoung/nil/app/mds/repository"
)

// Repository provides access to repository database.
type Repository interface {
	PublishCommand(op, query string) (result sql.Result, err error)
	Query(txid repository.TxID, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(txid repository.TxID, query string, args ...interface{}) *sql.Row
	Execute(txid repository.TxID, query string) (sql.Result, error)
	FindAllVolumes(txid repository.TxID) ([]*Volume, error)
}
