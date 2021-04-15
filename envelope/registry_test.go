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

package envelope_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tamasd/constellation/envelope"
)

func TestRegistry_Create(t *testing.T) {
	r := envelope.NewRegistry()

	r.Register("data", reflect.TypeOf(&data{}))

	require.NotNil(t, r.Create("data"))
}

func TestRegistry_Create_Invalid(t *testing.T) {
	r := envelope.NewRegistry()

	require.Nil(t, r.Create("data"))
}

func TestRegistry_MessageType(t *testing.T) {
	table := []struct {
		name           string
		registeredType reflect.Type
		value          interface{}
	}{
		{"data-s-s", reflect.TypeOf(data{}), data{}},
		{"data-p-s", reflect.TypeOf(&data{}), data{}},
		{"data-s-p", reflect.TypeOf(data{}), &data{}},
		{"data-p-p", reflect.TypeOf(&data{}), &data{}},
	}

	for _, row := range table {
		t.Run(row.name, func(t *testing.T) {
			r := envelope.NewRegistry()
			r.Register(row.name, row.registeredType)
			require.Equal(t, row.name, r.MessageType(row.value))
		})
	}
}
