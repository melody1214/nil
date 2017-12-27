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

// AddRegion inserts a new region into the region table.
func (m *MySQL) AddRegion(region, addr string) error {
	_, err := m.db.Exec(
		`
		INSERT INTO region (region_name, end_point)
		SELECT * FROM (SELECT ? AS rn, ? AS ep) AS tmp
		WHERE NOT EXISTS ( 
			SELECT region_name FROM region WHERE region_name = ?
		) LIMIT 1;
		`,
		m.cfg.Raft.LocalClusterRegion,
		m.cfg.Raft.LocalClusterAddr,
		m.cfg.Raft.LocalClusterRegion,
	)

	return err
}
