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

package database_test

import (
	"crypto/rand"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tamasd/constellation/database"
	"github.com/tamasd/constellation/logger"
	"github.com/tamasd/constellation/logger/testlogger"
	"github.com/tamasd/constellation/util"
	"github.com/tamasd/constellation/uuid"
)

const testSchema = `
	CREATE TABLE test (
		id UUID NOT NULL,
		data text NOT NULL,
		PRIMARY KEY(id)
	);
`

func TestMigrations(t *testing.T) {
	conn, l := getConnection(t)

	m := database.DecorateMigrationsWithChecks(
		database.DefineMigrations(
			"test",
			func(logger logger.Logger, conn database.Connection) error {
				_, err := conn.Exec(testSchema)
				return err
			},
		),
		func(logger logger.Logger, conn database.Connection) error {
			constraints, err := database.LoadConstraints(conn, "test", "")
			require.NoError(t, err)

			require.Equal(t, []database.Constraint{
				{"test_pkey", database.ConstraintTypePrimary},
			}, constraints)

			return nil
		},
	)

	err := database.MigrateSchema(l, conn, m)
	require.NoError(t, err)
}

func TestDatabase(t *testing.T) {
	conn, l := setupTest(t)
	key := genKey()

	t.Run("insert record", func(t *testing.T) {
		id := uuid.Generate(key)
		data := util.RandomHexString(4096)
		_, err := conn.Exec(`INSERT INTO test(id, data) VALUES($1, $2)`, id, data)
		require.NoError(t, err)
		assertLog(t, l)
	})

	t.Run("start transaction, delete row, rollback", func(t *testing.T) {
		tx, err := database.MaybeBegin(conn)
		require.NoError(t, err)
		require.NotEqual(t, conn, tx)
		assertLog(t, l)

		_, err = tx.Exec(`DELETE FROM test`)
		require.NoError(t, err)
		assertLog(t, l)

		assertTestTableRowCount(t, tx, 0)
		assertLog(t, l)

		require.NoError(t, database.MaybeRollback(tx))
		assertLog(t, l)

		assertTestTableRowCount(t, conn, 1)
		assertLog(t, l)
	})

	t.Run("start a transaction, update row, commit", func(t *testing.T) {
		tx, err := database.MaybeBegin(conn)
		require.NoError(t, err)
		require.NotEqual(t, conn, tx)
		assertLog(t, l)

		data := util.RandomHexString(2048)
		_, err = tx.Exec(`UPDATE test SET data = $1`, data)
		require.NoError(t, err)
		assertLog(t, l)

		require.NoError(t, database.MaybeCommit(tx))
		assertLog(t, l)

		rows, err := conn.Query(`SELECT data FROM test`)
		require.NoError(t, err)
		assertLog(t, l)
		for rows.Next() {
			var scanned string
			err = rows.Scan(&scanned)
			require.NoError(t, err)
			require.Equal(t, data, scanned)
		}
	})
}

func TestMaybeBegin(t *testing.T) {
	tf := &fakeTransactionFactory{}
	conn := &fakeConnection{}

	tx, err := database.MaybeBegin(tf)
	require.NoError(t, err)
	require.NotNil(t, tx)

	c, err := database.MaybeBegin(conn)
	require.NoError(t, err)
	require.Equal(t, conn, c)
}

func TestMaybeRollback(t *testing.T) {
	errMsg := util.RandomHexString(8)
	tf := &fakeTransactionFactory{
		err: errors.New(errMsg),
	}
	tx, _ := tf.Begin()

	err := database.MaybeRollback(tx)
	require.Error(t, err)
	require.Equal(t, errMsg, err.Error())

	tf.err = nil
	tx, _ = tf.Begin()
	require.NoError(t, database.MaybeRollback(tx))

	require.NoError(t, database.MaybeRollback(&fakeConnection{}))
}

func TestMaybeCommit(t *testing.T) {
	errMsg := util.RandomHexString(8)
	tf := &fakeTransactionFactory{
		err: errors.New(errMsg),
	}
	tx, _ := tf.Begin()

	err := database.MaybeCommit(tx)
	require.Error(t, err)
	require.Equal(t, errMsg, err.Error())

	tf.err = nil
	tx, _ = tf.Begin()
	require.NoError(t, database.MaybeCommit(tx))

	require.NoError(t, database.MaybeCommit(&fakeConnection{}))
}

type fakeTransactionFactory struct {
	fakeConnection
	err error
}

func (f *fakeTransactionFactory) Begin() (database.Transaction, error) {
	return &fakeTransaction{
		fakeConnection: f.fakeConnection,
		err:            f.err,
	}, nil
}

type fakeTransaction struct {
	fakeConnection
	err error
}

func (t *fakeTransaction) Commit() error {
	return t.err
}

func (t *fakeTransaction) Rollback() error {
	return t.err
}

type fakeConnection struct {
}

func (c *fakeConnection) Exec(_ string, _ ...interface{}) (sql.Result, error) {
	panic("implement me")
}

func (c *fakeConnection) Query(_ string, _ ...interface{}) (*sql.Rows, error) {
	panic("implement me")
}

func (c *fakeConnection) QueryRow(_ string, _ ...interface{}) *sql.Row {
	panic("implement me")
}

func assertTestTableRowCount(t *testing.T, conn database.Connection, expected int) {
	count := -1
	require.NoError(t, conn.QueryRow(`SELECT COUNT(*) FROM test`).Scan(&count))
	require.Equal(t, expected, count)
}

func setupTest(t *testing.T) (database.Connection, *testlogger.Logger) {
	conn, l := getConnection(t)

	_, err := conn.Exec(testSchema)
	require.NoError(t, err)
	assertLog(t, l)

	return conn, l
}

func assertLog(t *testing.T, l *testlogger.Logger) {
	require.NotZero(t, l.Buffer.Len())
	l.Buffer.Reset()
}

func getConnection(t *testing.T) (database.Connection, *testlogger.Logger) {
	conn, cleanup := database.TestConnect(os.Getenv("DATABASE_URL"))
	t.Cleanup(cleanup)
	l := testlogger.TestLogger()

	return database.NewLoggerDB(l, conn), l
}

func genKey() []byte {
	key := make([]byte, 64)
	_, _ = rand.Read(key)

	return key
}
