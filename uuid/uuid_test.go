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

package uuid_test

import (
	"crypto/rand"
	"testing"

	gouuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"github.com/tamasd/constellation/uuid"
)

func TestUUID(t *testing.T) {
	key := genKey()

	gen := uuid.Generate(key)
	require.NotZero(t, gen)
	require.True(t, gen.Verify(key))

	require.Equal(t, uint8(4), gen.Version())
	require.Equal(t, uint8(gouuid.VariantRFC4122), gen.Variant())

}

func TestUUID_Tampered(t *testing.T) {
	key := genKey()

	gen := uuid.Generate(key)
	gen[0] = 0

	require.False(t, gen.Verify(key))
}

func TestParseAndVerify(t *testing.T) {
	key := genKey()

	gen := uuid.Generate(key)
	require.NotZero(t, gen)
	require.True(t, uuid.ParseAndVerify(key, gen.String()))
}

func TestParseAndVerify_InvalidInput(t *testing.T) {
	key := genKey()
	require.False(t, uuid.ParseAndVerify(key, ""))
}

func genKey() []byte {
	key := make([]byte, 64)
	_, _ = rand.Read(key)

	return key
}
