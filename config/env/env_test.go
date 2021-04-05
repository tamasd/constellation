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

package env_test

import (
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tamasd/constellation/config/env"
)

type data struct {
	A int
	B string
	C bool
	D uint
	E float64
}

type simpleData struct {
	A int
	B bool
}

type invalidData struct {
	f func()
}

func TestUnmarshaler_Unmarshal(t *testing.T) {
	table := []struct {
		name     string
		env      map[string]string
		expected interface{}
		prefix   string
	}{
		{
			"basic data", map[string]string{
				"FOO_A": "-2",
				"FOO_B": "asdf",
				"FOO_C": "true",
				"FOO_D": "5",
				"FOO_E": "-1.2",
			}, &data{-2, "asdf", true, 5, -1.2}, "FOO",
		},
		{
			"simple data", map[string]string{
				"A": "5",
				"B": "false",
			}, &simpleData{5, false}, "",
		},
	}

	for _, row := range table {
		os.Clearenv()
		t.Run(row.name, func(t *testing.T) {
			for k, v := range row.env {
				err := os.Setenv(k, v)
				require.Nil(t, err)
			}

			v := reflect.New(reflect.Indirect(reflect.ValueOf(row.expected)).Type()).Interface()
			u := env.NewUnmarshaler()
			u.Prefix = row.prefix
			u.Strict = true
			err := u.Unmarshal(v)
			require.Nil(t, err)
			require.Equal(t, row.expected, v)
		})
	}
}

func TestUnmarshaler_Unmarshal_NonPointer(t *testing.T) {
	u := env.NewUnmarshaler()
	d := simpleData{}
	err := u.Unmarshal(d)
	require.NotNil(t, err)
	require.Equal(t, "env: Unmarshal(non-pointer env_test.simpleData)", err.Error())
}

func TestUnmarshaler_Unmarshal_Nil(t *testing.T) {
	u := env.NewUnmarshaler()
	err := u.Unmarshal(nil)
	require.NotNil(t, err)
	require.Equal(t, "env: Unmarshal(nil)", err.Error())
}

func TestUnmarshaler_Unmarshal_InvalidType(t *testing.T) {
	u := env.NewUnmarshaler()
	u.Strict = true
	d := &invalidData{}
	err := u.Unmarshal(d)
	require.NotNil(t, err)
	require.Equal(t, "env: Unmarshal(func())", err.Error())
}
