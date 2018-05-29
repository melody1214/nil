package mysql

import (
	"database/sql"
	"fmt"
	"sync"

	"github.com/chanyoung/nil/app/mds/repository"
	"github.com/chanyoung/nil/pkg/util/config"
	"github.com/chanyoung/nil/pkg/util/uuid"
	_ "github.com/go-sql-driver/mysql"
)

// mySQL is the handle of MySQL client.
type mySQL struct {
	cfg *config.Mds
	db  *sql.DB
	txs map[repository.TxID]*sql.Tx
	txl sync.RWMutex
}

// newMySQL returns MySQL handle with the opened db.
func newMySQL(cfg *config.Mds) (*mySQL, error) {
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

	m := &mySQL{
		cfg: cfg,
		db:  db,
		txs: make(map[repository.TxID]*sql.Tx),
	}
	if err = m.init(); err != nil {
		m.db.Close()
		return nil, err
	}

	return m, nil
}

// Close closes mysql database.
func (m *mySQL) close() {
	m.db.Close()
}

func (m *mySQL) init() error {
	// Generates base tables.
	for _, q := range generateSQLBase {
		if _, err := m.db.Exec(q); err != nil {
			return err
		}
	}

	return nil
}

// Execute executes query.
func (m *mySQL) execute(txid repository.TxID, query string) (sql.Result, error) {
	if txid != "" {
		return m.db.Exec(query)
	}

	tx, err := m.getTx(txid)
	if err != nil {
		return nil, err
	}
	return tx.Exec(query)
}

// QueryRow executes a query that is expected to return at most one row.
func (m *mySQL) queryRow(txid repository.TxID, query string, args ...interface{}) *sql.Row {
	if txid != "" {
		return m.db.QueryRow(query, args...)
	}

	tx, err := m.getTx(txid)
	if err != nil {
		return nil
	}
	return tx.QueryRow(query, args...)
}

// Query executes a query that returns rows.
func (m *mySQL) query(txid repository.TxID, query string, args ...interface{}) (*sql.Rows, error) {
	if txid != "" {
		return m.db.Query(query, args...)
	}

	tx, err := m.getTx(txid)
	if err != nil {
		return nil, err
	}
	return tx.Query(query, args...)
}

func (m *mySQL) begin() (txid repository.TxID, err error) {
	tx, err := m.db.Begin()
	if err != nil {
		return "", err
	}

	m.txl.Lock()
	defer m.txl.Unlock()

	for {
		txid = repository.TxID(uuid.Gen())
		if _, ok := m.txs[txid]; ok {
			continue
		}
		break
	}

	m.txs[txid] = tx
	return
}

func (m *mySQL) getTx(txid repository.TxID) (*sql.Tx, error) {
	tx, ok := m.txs[txid]
	if ok == false {
		return nil, fmt.Errorf("no such tx with matching txid: %s", txid)
	}
	return tx, nil
}

func (m *mySQL) rollback(txid repository.TxID) error {
	m.txl.Lock()
	defer m.txl.Unlock()
	defer delete(m.txs, txid)

	tx, err := m.getTx(txid)
	if err != nil {
		return err
	}

	return tx.Rollback()
}

func (m *mySQL) commit(txid repository.TxID) error {
	m.txl.Lock()
	defer m.txl.Unlock()

	tx, err := m.getTx(txid)
	if err != nil {
		return err
	}

	// Delete Tx from the map only when succeess the transaction.
	err = tx.Commit()
	if err == nil {
		defer delete(m.txs, txid)
	}
	return err
}
