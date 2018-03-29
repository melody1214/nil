package mysql

import (
	"database/sql"
	"fmt"

	"github.com/chanyoung/nil/pkg/util/config"
	_ "github.com/go-sql-driver/mysql"
)

// MySQL is the handle of MySQL client.
type MySQL struct {
	cfg *config.Mds
	db  *sql.DB
}

// New returns MySQL handle with the opened db.
func New(cfg *config.Mds) (*MySQL, error) {
	db, err := sql.Open(
		"mysql",
		fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			cfg.MySQLUser,
			cfg.MySQLPassword,
			cfg.MySQLHost,
			cfg.MySQLPort,
			cfg.MySQLDatabase,
		),
	)
	if err != nil {
		return nil, err
	}

	m := &MySQL{
		cfg: cfg,
		db:  db,
	}
	if err = m.init(); err != nil {
		m.db.Close()
		return nil, err
	}

	return m, nil
}

// Close closes mysql database.
func (m *MySQL) Close() {
	m.db.Close()
}

func (m *MySQL) init() error {
	// Generates base tables.
	for _, q := range generateSQLBase {
		if _, err := m.db.Exec(q); err != nil {
			return err
		}
	}

	return nil
}

// Execute executes query.
func (m *MySQL) Execute(query string) (sql.Result, error) {
	return m.db.Exec(query)
}

// QueryRow executes a query that is expected to return at most one row.
func (m *MySQL) QueryRow(query string, args ...interface{}) *sql.Row {
	return m.db.QueryRow(query, args...)
}

// Query executes a query that returns rows.
func (m *MySQL) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return m.db.Query(query, args...)
}
