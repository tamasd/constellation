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

package util_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tamasd/constellation/util"
)

type placeholderPair struct {
	start  int
	length int
}

func TestGeneratePlaceholders(t *testing.T) {
	table := map[placeholderPair]string{
		{0, 0}: "",
		{2, 1}: "$2",
		{1, 2}: "$1, $2",
	}

	for pair, result := range table {
		placeholders := util.GeneratePlaceholders(pair.start, pair.length)
		require.Equal(t, result, placeholders)
	}
}

func TestRandomHexString(t *testing.T) {
	for i := 0; i < 12; i++ {
		require.Len(t, util.RandomHexString(i), i)
	}
}

func TestSetContext(t *testing.T) {
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	require.Nil(t, err)

	key := util.RandomHexString(8)
	value := util.RandomHexString(8)

	r = util.SetContext(r, key, value)

	require.Equal(t, value, r.Context().Value(key))
}

func TestToSnakeCase(t *testing.T) {
	require.Equal(t, "test_uuid_foo", util.ToSnakeCase("TestUUIDFoo"))
}
