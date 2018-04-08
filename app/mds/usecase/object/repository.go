package object

import "database/sql"

// Repository provides access to object database.
type Repository interface {
	Execute(query string) (sql.Result, error)
}
