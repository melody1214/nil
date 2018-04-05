package bucket

import "database/sql"

// Repository provides access to bucket database.
type Repository interface {
	PublishCommand(op, query string) (result sql.Result, err error)
}
