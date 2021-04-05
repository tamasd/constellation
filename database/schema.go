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
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/tamasd/constellation/logger"
)

// BootstrapSchema is the schema of the migrations table.
//
// Changing this is a major change, since "normal" updates won't work after it,
// so special upgrade steps will be needed.
const BootstrapSchema = `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
	CREATE EXTENSION IF NOT EXISTS "btree_gist";

    CREATE TABLE IF NOT EXISTS migrations(
        name CHARACTER VARYING(255) NOT NULL PRIMARY KEY,
        current INT NOT NULL DEFAULT -1
    );
`

type simpleMigrationsProvider struct {
	name       string
	migrations Migrations
}

func (p *simpleMigrationsProvider) Name() string {
	return p.name
}

func (p *simpleMigrationsProvider) Migrations() Migrations {
	return p.migrations
}

type simpleMigrationCheckDecorator struct {
	MigrationsProvider
	checks Checks
}

func (d *simpleMigrationCheckDecorator) Checks() Checks {
	return d.checks
}

type MigrationsProvider interface {
	Name() string
	Migrations() Migrations
}

type MigrationCheckProvider interface {
	MigrationsProvider
	Checks() Checks
}

func DecorateMigrationsWithChecks(mp MigrationsProvider, checks ...Check) MigrationCheckProvider {
	return &simpleMigrationCheckDecorator{
		MigrationsProvider: mp,
		checks:             checks,
	}
}

type Migration func(logger logger.Logger, conn Connection) error

func DefineMigrations(name string, gens ...Migration) MigrationsProvider {
	return &simpleMigrationsProvider{
		name:       name,
		migrations: gens,
	}
}

type Migrations []Migration

func (g Migrations) UpgradeFrom(last int, logger logger.Logger, conn Connection) (int, error) {
	for next := last + 1; next < len(g); next++ {
		if err := g[next](logger, conn); err != nil {
			return next - 1, err
		}
	}

	return len(g) - 1, nil
}

func MigrationVersion(conn Connection, name string) (int, error) {
	var current int
	err := conn.QueryRow(`
        SELECT COALESCE(
            (SELECT current FROM migrations WHERE name = $1),
            -1
        )
    `, name).Scan(&current)

	return current, err
}

func SaveMigrationVersion(conn Connection, name string, version int) error {
	_, err := conn.Exec(`
        INSERT INTO migrations (name, current) VALUES ($1, $2)
        ON CONFLICT (name) DO UPDATE SET current = $2
    `, name, version)

	return err
}

type Checks []Check

type Check func(logger logger.Logger, conn Connection) error

func MigrateSchema(logger logger.Logger, conn Connection, providers ...MigrationsProvider) error {
	migrationStart := time.Now()

	_, err := conn.Exec(BootstrapSchema)
	if err != nil {
		return errors.Wrap(err, "failed to ensure bootstrap schema")
	}

	checkers := make([]MigrationCheckProvider, 0, len(providers))

	for _, p := range providers {
		l := logger.WithField("migration-name", p.Name())
		start := time.Now()

		if err = updateSchema(l, conn, p); err != nil {
			l.WithError(err).Errorln("failed to execute migration")
			return errors.Wrap(err, "failed to execute migration: "+p.Name())
		}
		l.WithField("duration", time.Since(start)).Infoln("migration complete")

		if chk, ok := p.(MigrationCheckProvider); ok {
			checkers = append(checkers, chk)
		}
	}

	for _, checker := range checkers {
		l := logger.WithField("check-name", checker.Name())
		start := time.Now()

		for i, check := range checker.Checks() {
			if err = check(l, conn); err != nil {
				l.WithError(err).Errorln("failed to execute check")
				return errors.Wrap(err, "failed to execute check: "+checker.Name()+"/"+strconv.Itoa(i))
			}
		}
		l.WithField("duration", time.Since(start)).Infoln("checks complete")
	}

	logger.WithField("duration", time.Since(migrationStart)).Infoln("installation complete")

	return nil
}

func updateSchema(logger logger.Logger, conn Connection, provider MigrationsProvider) error {
	name := provider.Name()
	tx, err := MaybeBegin(conn)
	if err != nil {
		return errors.Wrap(err, "failed to open transaction")
	}

	defer func() {
		if err = MaybeRollback(tx); err != nil {
			logger.Errorln(err)
		}
	}()

	oldversion, err := MigrationVersion(tx, name)
	if err != nil {
		return errors.Wrap(err, "failed to load migration version")
	}

	logger = logger.WithField("old-version", oldversion)
	logger.Debugln("loaded old version")

	newversion, err := provider.Migrations().UpgradeFrom(oldversion, logger, tx)
	if err != nil {
		return errors.Wrap(err, "failed to upgrade version")
	}
	logger = logger.WithField("new-version", newversion)
	logger.Debugln("migrations finished")

	if err = SaveMigrationVersion(tx, name, newversion); err != nil {
		return errors.Wrap(err, "failed to save version information")
	}
	logger.Debugln("migrations table updated")

	return MaybeCommit(conn)
}
