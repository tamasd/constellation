/*
 * Copyright Tamás Demeter-Haludka 2021
 *
 * This file is part of Constellation.
 *
 * Constellation is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Constellation is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with Constellation.  If not, see <https://www.gnu.org/licenses/>.
 */

package database

import (
	"database/sql"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/tamasd/constellation/logger"
	"github.com/tamasd/constellation/util"
)

var (
	spaces = regexp.MustCompile(`\s+`)
)

func MaybeBegin(conn Connection) (Connection, error) {
	if f, ok := conn.(TransactionFactory); ok {
		return f.Begin()
	}

	return conn, nil
}

func MaybeRollback(conn Connection) error {
	tx, ok := conn.(Transaction)
	if !ok {
		return nil
	}

	if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
		return err
	}

	return nil
}

func MaybeCommit(conn Connection) error {
	if tx, ok := conn.(Transaction); ok {
		return tx.Commit()
	}

	return nil
}

// Connection represents a database connection.
type Connection interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

type ConfigurableConnection interface {
	Connection

	SetConnMaxLifetime(d time.Duration)
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
}

// Transaction represents a database connection with an active transaction.
type Transaction interface {
	Connection
	Commit() error
	Rollback() error
}

// TransactionFactory can initiate a transaction.
type TransactionFactory interface {
	Begin() (Transaction, error)
}

type dbWrapper struct {
	*sql.DB
}

func (w *dbWrapper) Begin() (Transaction, error) {
	return w.DB.Begin()
}

type loggerDB struct {
	logger logger.Logger
	db     Connection
}

// NewLoggerDB wraps a database connection with a logger.
func NewLoggerDB(logger logger.Logger, db Connection) Connection {
	ldb := loggerDB{
		logger: logger,
		db:     db,
	}

	if _, ok := db.(Transaction); ok {
		return &transactionLoggerDB{ldb}
	}
	if _, ok := db.(TransactionFactory); ok {
		return &transactionFactoryLoggerDB{ldb}
	}

	return &ldb
}

func (d *loggerDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := d.db.Exec(query, args...)
	l := d.logger.WithFields(logger.Fields{
		"query":    cleanSQL(query),
		"args":     args,
		"duration": time.Since(start),
	})
	if err != nil {
		l = l.WithError(err)
	}
	l.Debugln("executing query")
	return res, err
}

func (d *loggerDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := d.db.Query(query, args...)
	l := d.logger.WithFields(logger.Fields{
		"query":    cleanSQL(query),
		"args":     args,
		"duration": time.Since(start),
	})
	if err != nil {
		l = l.WithError(err)
	}
	l.Debugln("running query")

	return rows, err
}

func (d *loggerDB) QueryRow(query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := d.db.QueryRow(query, args...)
	d.logger.WithFields(logger.Fields{
		"query":    cleanSQL(query),
		"args":     args,
		"duration": time.Since(start),
	}).Debugln("running query row")

	return row
}

type transactionFactoryLoggerDB struct {
	loggerDB
}

func (d *transactionFactoryLoggerDB) Begin() (Transaction, error) {
	const msg = "begin transaction"
	if f, ok := d.db.(TransactionFactory); ok {
		start := time.Now()
		tx, err := f.Begin()
		l := d.logger.WithFields(logger.Fields{
			"transaction-id": util.RandomHexString(8),
			"duration":       time.Since(start),
		})
		if err != nil {
			l = l.WithError(err)
			l.Debugln(msg)
			return nil, err
		}
		l.Debugln(msg)

		return NewLoggerDB(l, tx).(Transaction), nil
	}

	return nil, nil
}

type transactionLoggerDB struct {
	loggerDB
}

func (d *transactionLoggerDB) Commit() error {
	start := time.Now()
	err := d.db.(Transaction).Commit()
	l := d.logger.WithFields(logger.Fields{
		"duration": time.Since(start),
	})
	if err != nil && err != sql.ErrTxDone {
		l = l.WithError(err)
	}
	l.Debugln("commit transaction")

	return err
}

func (d *transactionLoggerDB) Rollback() error {
	start := time.Now()
	err := d.db.(Transaction).Rollback()
	l := d.logger.WithFields(logger.Fields{
		"duration": time.Since(start),
	})
	if err != nil && err != sql.ErrTxDone {
		l = l.WithError(err)
	}
	l.Debugln("rollback transaction")

	return err
}

// Connect creates a database connection to a PostgreSQL database.
func Connect(dbUrl string) (ConfigurableConnection, error) {
	conn, err := sql.Open("postgres", dbUrl)
	if err != nil {
		return nil, err
	}

	return &dbWrapper{
		DB: conn,
	}, nil
}

func cleanSQL(query string) string {
	return spaces.ReplaceAllString(strings.TrimSpace(query), " ")
}

// TestConnect creates a connection to a PostgreSQL database for testing.
//
// This will set up a temporary schema, and delete it after the test is run.
//
// The correct usage of this function:
//
//     conn, cleanup := database.TestConnect(os.Getenv("DATABASE_URL"))
//     t.Cleanup(cleanup)
func TestConnect(dbUrl string) (Connection, func()) {
	conn, err := Connect(dbUrl)
	if err != nil {
		panic(err)
	}
	conn.SetConnMaxLifetime(120 * time.Second)
	conn.SetMaxIdleConns(1)
	conn.SetMaxOpenConns(1)

	schema := "ct_" + strings.ToLower(util.RandomHexString(8))
	_, err = conn.Exec("CREATE SCHEMA " + schema)
	if err != nil {
		panic(err)
	}
	setSearchPath(conn, schema)

	return conn, func() {
		_, err := conn.Exec("DROP SCHEMA " + schema + " CASCADE;")
		if err != nil {
			panic(err)
		}
	}
}

func setSearchPath(conn Connection, schema string) {
	if _, err := conn.Exec("SET search_path TO " + schema + ", public;"); err != nil {
		panic(err)
	}
}
