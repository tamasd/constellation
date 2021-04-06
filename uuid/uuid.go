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

package uuid

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql/driver"

	gouuid "github.com/satori/go.uuid"
	"github.com/tamasd/constellation/util"
)

type UUID [16]byte

var Nil = UUID{}

// Generates a signed UUID.
//
// The key must be 64 bytes long.
func Generate(key []byte) UUID {
	u := UUID(gouuid.NewV4())

	sum := hmacsum(u[:12], key)
	copy(u[12:], sum)

	return u
}

func hmacsum(msg []byte, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	_, err := mac.Write(msg)
	util.Must(err)
	return mac.Sum(nil)
}

// Verifies a signed UUID.
//
// The key must be 64 bytes long.
func (u UUID) Verify(key []byte) bool {
	sum := hmacsum(u[:12], key)
	return hmac.Equal(u[12:], sum[:4])
}

// Parses and verifies a signed UUID.
//
// The key must be 64 bytes long.
func ParseAndVerify(key []byte, input string) bool {
	u, uerr := gouuid.FromString(input)
	if uerr != nil {
		return false
	}
	return UUID(u).Verify(key)
}

func (u UUID) IsNil() bool {
	return Equal(u, Nil)
}

func Equal(u1, u2 UUID) bool {
	return gouuid.Equal(gouuid.UUID(u1), gouuid.UUID(u2))
}

func FromBytes(input []byte) (UUID, error) {
	u, err := gouuid.FromBytes(input)
	return UUID(u), err
}

func FromByteOrNil(input []byte) UUID {
	return UUID(gouuid.FromBytesOrNil(input))
}

func FromString(input string) (UUID, error) {
	u, err := gouuid.FromString(input)
	return UUID(u), err
}

func FromStringOrNil(input string) UUID {
	return UUID(gouuid.FromStringOrNil(input))
}

func (u UUID) Bytes() []byte {
	return u[:]
}

func (u UUID) MarshalBinary() ([]byte, error) {
	return gouuid.UUID(u).MarshalBinary()
}

func (u UUID) MarshalText() ([]byte, error) {
	return gouuid.UUID(u).MarshalText()
}

func (u *UUID) Scan(src interface{}) error {
	return (*gouuid.UUID)(u).Scan(src)
}

func (u UUID) String() string {
	return gouuid.UUID(u).String()
}

func (u *UUID) UnmarshalBinary(data []byte) error {
	return (*gouuid.UUID)(u).UnmarshalBinary(data)
}

func (u *UUID) UnmarshalText(text []byte) error {
	return (*gouuid.UUID)(u).UnmarshalText(text)
}

func (u UUID) Value() (driver.Value, error) {
	return gouuid.UUID(u).Value()
}

func (u UUID) Variant() byte {
	return gouuid.UUID(u).Variant()
}

func (u UUID) Version() byte {
	return gouuid.UUID(u).Version()
}
