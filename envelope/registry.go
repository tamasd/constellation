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

package envelope

import (
	"reflect"
	"strings"
)

type Registry struct {
	nameLookup map[string]reflect.Type
	typeLookup map[string]string
}

func NewRegistry() *Registry {
	return &Registry{
		nameLookup: map[string]reflect.Type{},
		typeLookup: map[string]string{},
	}
}

func (mtr *Registry) Register(typeName string, typ reflect.Type) {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	mtr.nameLookup[typeName] = typ
	mtr.typeLookup[typ.String()] = typeName
}

func (mtr *Registry) Create(typeName string) interface{} {
	typ := mtr.nameLookup[typeName]
	if typ == nil {
		return nil
	}

	return reflect.New(typ).Interface()
}

func (mtr *Registry) MessageType(v interface{}) string {
	name := reflect.TypeOf(v).String()
	if strings.HasPrefix(name, "*") {
		name = name[1:]
	}
	return mtr.typeLookup[name]
}
