package auth

import (
	"database/sql"

	"github.com/chanyoung/nil/app/mds/repository"
)

// Repository provides access to auth database.
type Repository interface {
	QueryRow(txid repository.TxID, query string, args ...interface{}) *sql.Row
}
