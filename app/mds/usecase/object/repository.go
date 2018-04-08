package object

import "database/sql"

// Repository provides access to object database.
type Repository interface {
	Execute(query string) (sql.Result, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}
