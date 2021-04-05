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
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

var _ WritableProvider = &MemoryConfigProvider{}

type MemoryConfigProvider struct {
	store map[string]interface{}
}

func NewMemoryConfigProvider() *MemoryConfigProvider {
	m := &MemoryConfigProvider{}
	m.Reset()
	return m
}

func (m *MemoryConfigProvider) Reset() {
	m.store = make(map[string]interface{})
}

func (m *MemoryConfigProvider) CanSave(key string) bool {
	return true
}

func (m *MemoryConfigProvider) Save(key string, v interface{}) error {
	m.store[key] = v
	return nil
}

func (m *MemoryConfigProvider) Has(key string) bool {
	_, found := m.store[key]
	return found
}

func (m *MemoryConfigProvider) Unmarshal(key string, v interface{}) error {
	val, found := m.store[key]
	if found {
		return mergo.Merge(v, val)
	}

	return errors.New("value not found")
}
