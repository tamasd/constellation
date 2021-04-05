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
	"os"
	"strings"

	"github.com/tamasd/constellation/config/env"
)

var _ Provider = &EnvConfigProvider{}

type EnvConfigProvider struct {
	Prefix    string
	Separator string
	variables map[string]string
}

func NewEnvConfigProvider() *EnvConfigProvider {
	return &EnvConfigProvider{
		Prefix:    "",
		Separator: "_",
		variables: nil,
	}
}

func (e *EnvConfigProvider) maybeInitializeVariables() {
	if e.variables != nil {
		return
	}

	e.variables = make(map[string]string)
	for _, ev := range os.Environ() {
		parts := strings.SplitN(ev, "=", 2)
		e.variables[parts[0]] = parts[1]
	}
}

func (e *EnvConfigProvider) Reset() {
	e.variables = nil
}

func (e *EnvConfigProvider) prefixedKey(key string) string {
	key = strings.ToUpper(key)
	if e.Prefix == "" {
		return key
	}
	return e.Prefix + e.Separator + key
}

func (e *EnvConfigProvider) loader(key string) (string, bool) {
	val, found := e.variables[e.prefixedKey(key)]
	return val, found
}

func (e *EnvConfigProvider) Has(key string) bool {
	e.maybeInitializeVariables()
	key = e.prefixedKey(key)
	for k := range e.variables {
		if strings.HasPrefix(k, key) {
			return true
		}
	}

	return false
}

func (e *EnvConfigProvider) Unmarshal(key string, v interface{}) error {
	e.maybeInitializeVariables()
	u := env.NewUnmarshaler()
	u.Prefix = key
	u.Separator = e.Separator
	u.Loader = e.loader

	return u.Unmarshal(v)
}
