package recovery

import (
	"database/sql"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/cmap"
)

// Repository provides access to repository database.
type Repository interface {
	PublishCommand(op, query string) (result sql.Result, err error)
	Query(txid repository.TxID, query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(txid repository.TxID, query string, args ...interface{}) *sql.Row
	Execute(txid repository.TxID, query string) (sql.Result, error)
	FindAllVolumes(txid repository.TxID) ([]*Volume, error)
	MakeNewEncodingGroup(txid repository.TxID, encGrp *cmap.EncodingGroup) error
	Begin() (repository.TxID, error)
	Rollback(repository.TxID) error
	Commit(repository.TxID) error
}
