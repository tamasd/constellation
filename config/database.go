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

package config

import (
	"encoding/json"

	"github.com/tamasd/constellation/database"
	"github.com/tamasd/constellation/logger"
)

type Database struct {
	conn     database.Connection
	readOnly bool
}

func NewDatabase(conn database.Connection, readOnly bool) *Database {
	return &Database{
		conn:     conn,
		readOnly: readOnly,
	}
}

func (d *Database) Load(name string) (*Collection, error) {
	c := NewCollection()
	c.SetTemporary(true)
	c.AddProviders(NewDatabaseConfigProvider(d.conn, name, d.readOnly))

	return c, nil
}

func (d *Database) Name() string {
	return "config-database"
}

func (d *Database) Migrations() database.MigrationsProvider {
	return database.DefineMigrations(
		"config",
		func(l logger.Logger, conn database.Connection) error {
			_, err := conn.Exec(`
				CREATE TABLE namespace (
					namespace character varying NOT NULL,
					CONSTRAINT namespace_pkey PRIMARY KEY (namespace)
				);

				CREATE TABLE config (
					namespace character varying NOT NULL,
					name character varying NOT NULL,
					value jsonb NOT NULL,
					CONSTRAINT config_pkey PRIMARY KEY (namespace, name)
				);
			`)
			return err
		},
	)
}

type DatabaseConfigProvider struct {
	conn      database.Connection
	namespace string
	readOnly  bool
}

func NewDatabaseConfigProvider(conn database.Connection, namespace string, readOnly bool) *DatabaseConfigProvider {
	return &DatabaseConfigProvider{
		conn:      conn,
		namespace: namespace,
		readOnly:  readOnly,
	}
}

func (p *DatabaseConfigProvider) Has(key string) bool {
	var found bool
	err := p.conn.QueryRow("SELECT true FROM config WHERE namespace = $1 AND name = $2", p.namespace, key).Scan(&found)
	return err == nil && found
}

func (p *DatabaseConfigProvider) Unmarshal(key string, v interface{}) error {
	var jv string
	if err := p.conn.QueryRow(`SELECT value FROM config WHERE namespace = $1 AND name = $2`, p.namespace, key).Scan(&jv); err != nil {
		return err
	}

	if jv == "" {
		return nil
	}

	return json.Unmarshal([]byte(jv), v)
}

func (p *DatabaseConfigProvider) CanSave(key string) bool {
	return !p.readOnly
}

func (p *DatabaseConfigProvider) Save(key string, v interface{}) error {
	jv, _ := json.Marshal(v)
	_, err := p.conn.Exec(`
		INSERT INTO config(namespace, name, value)
			VALUES($1, $2, $3)
			ON CONFLICT (config_pkey)
			DO UPDATE SET value = $3
	`, p.namespace, key, string(jv))

	return err
}
