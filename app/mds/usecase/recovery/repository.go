package recovery

import "database/sql"

// Repository provides access to repository database.
type Repository interface {
	PublishCommand(op, query string) (result sql.Result, err error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
	Execute(query string) (sql.Result, error)
}
