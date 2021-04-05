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

package env

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type InvalidUnmarshalError struct {
	Type        reflect.Type
	Unsupported bool
}

func (e *InvalidUnmarshalError) Error() string {
	if e.Type == nil {
		return "env: Unmarshal(nil)"
	}

	if e.Type.Kind() != reflect.Ptr && !e.Unsupported {
		return "env: Unmarshal(non-pointer " + e.Type.String() + ")"
	}

	return "env: Unmarshal(" + e.Type.String() + ")"
}

type Unmarshaler struct {
	NameConverter func(string) string
	Loader        func(string) (string, bool)
	Prefix        string
	Separator     string
	Strict        bool
}

func NewUnmarshaler() *Unmarshaler {
	return &Unmarshaler{
		NameConverter: strings.ToLower,
		Loader:        os.LookupEnv,
		Separator:     "_",
	}
}

func (u *Unmarshaler) Unmarshal(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if rerr, ok := r.(error); ok {
				err = rerr
			} else if str, ok := r.(string); ok {
				err = errors.New(str)
			} else {
				panic(r)
			}
		}
	}()

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &InvalidUnmarshalError{reflect.TypeOf(v), false}
	}
	u.unmarshal(u.Prefix, rv)

	return nil
}

func (u *Unmarshaler) unmarshal(current string, rv reflect.Value) {
	current = strings.ToUpper(current)
	switch rv.Kind() {
	case reflect.Bool:
		if val, found := u.Loader(current); found {
			val = strings.ToLower(val)
			switch val {
			case "true":
				rv.SetBool(true)
			case "false":
				rv.SetBool(false)
			default:
				panic("invalid value")
			}
		}
	case reflect.Int, reflect.Int32, reflect.Int8, reflect.Int16, reflect.Int64:
		if val, found := u.Loader(current); found {
			i, err := strconv.ParseInt(val, 0, 64)
			if err != nil {
				panic(err)
			}
			rv.SetInt(i)
		}
	case reflect.Uint, reflect.Uint32, reflect.Uint8, reflect.Uint16, reflect.Uint64:
		if val, found := u.Loader(current); found {
			i, err := strconv.ParseUint(val, 0, 64)
			if err != nil {
				panic(err)
			}
			rv.SetUint(i)
		}
	case reflect.Float32, reflect.Float64:
		if val, found := u.Loader(current); found {
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				panic(err)
			}
			rv.SetFloat(f)
		}
	case reflect.Ptr:
		u.unmarshal(current, rv.Elem())
	case reflect.String:
		if val, found := u.Loader(current); found {
			rv.SetString(val)
		}
	case reflect.Struct:
		structType := rv.Type()
		for i := 0; i < structType.NumField(); i++ {
			field := structType.Field(i)
			childname := u.childName(current, field.Name)
			u.unmarshal(childname, rv.Field(i))
		}
	default:
		if u.Strict {
			panic(&InvalidUnmarshalError{rv.Type(), true})
		}
	}
}

func (u *Unmarshaler) childName(current, child string) string {
	if u.NameConverter != nil {
		child = u.NameConverter(child)
	}
	if current == "" {
		return child
	}

	return current + u.Separator + child
}
