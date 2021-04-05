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

type ConstraintType string

const (
	ConstraintTypePrimary   ConstraintType = "p"
	ConstraintTypeUnique    ConstraintType = "u"
	ConstraintTypeCheck     ConstraintType = "c"
	ConstraintTypeForeign   ConstraintType = "f"
	ConstraintTypeExclusion ConstraintType = "x"

	ConstraintTypeUniqueIndex ConstraintType = "iu"
)

func (c ConstraintType) DropDefinition(name string) string {
	switch c {
	case ConstraintTypeUniqueIndex:
		return "DROP INDEX " + name + " CASCADE"
	default:
		return "ALTER TABLE attribute DROP CONSTRAINT " + name
	}
}

type Constraint struct {
	Name string
	Type ConstraintType
}

func LoadConstraints(conn Connection, relname, prefix string) ([]Constraint, error) {
	var args []interface{}
	constraintQuery := `
		SELECT con.conname, con.contype::text
		FROM pg_catalog.pg_constraint con
		INNER JOIN pg_catalog.pg_class rel ON rel.oid = con.conrelid
		INNER JOIN pg_catalog.pg_namespace nsp ON nsp.oid = con.connamespace
		WHERE
			nsp.nspname = current_schema() AND
			rel.relname = $1`

	indexQuery := `
		SELECT idx.relname as index_name, 'iu'
		FROM pg_index pgi
		INNER JOIN pg_class idx ON idx.oid = pgi.indexrelid
		INNER JOIN pg_namespace insp ON insp.oid = idx.relnamespace
		INNER JOIN pg_class tbl ON tbl.oid = pgi.indrelid
		INNER JOIN pg_namespace tnsp ON tnsp.oid = tbl.relnamespace
		WHERE
			pgi.indisunique AND
			tnsp.nspname = current_schema() AND
			tbl.relname = $1`

	if prefix != "" {
		constraintQuery += ` AND
			con.conname LIKE $2`
		indexQuery += ` AND
			idx.relname LIKE $2`
		args = []interface{}{
			relname,
			prefix + "%",
		}
	} else {
		args = []interface{}{
			relname,
		}
	}

	rows, err := conn.Query(constraintQuery+" UNION "+indexQuery, args...)
	if err != nil {
		return nil, err
	}

	var ret []Constraint

	for rows.Next() {
		constraint := Constraint{}
		if err = rows.Scan(&constraint.Name, &constraint.Type); err != nil {
			return nil, err
		}
		ret = append(ret, constraint)
	}

	return ret, nil
}
