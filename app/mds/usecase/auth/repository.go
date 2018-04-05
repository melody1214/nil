package auth

import "database/sql"

// Repository provides access to auth database.
type Repository interface {
	QueryRow(query string, args ...interface{}) *sql.Row
}
