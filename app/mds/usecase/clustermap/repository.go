package clustermap

import "database/sql"

// Repository provides access to cluster map database.
type Repository interface {
	Execute(query string) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
}
