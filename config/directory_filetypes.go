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
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v2"
)

type JSON struct {
	Strict bool
	Prefix string
	Indent string
}

func (t *JSON) Extensions() []string {
	return []string{"json"}
}

func (t *JSON) Unmarshal(stream io.Reader, v interface{}) error {
	dec := json.NewDecoder(stream)
	if t.Strict {
		dec.DisallowUnknownFields()
	}
	return dec.Decode(v)
}

func (t *JSON) Marshal(stream io.Writer, v interface{}) error {
	enc := json.NewEncoder(stream)
	enc.SetIndent(t.Prefix, t.Indent)
	return enc.Encode(v)
}

type YAML struct {
	Strict bool
}

func (t *YAML) Extensions() []string {
	return []string{"yml", "yaml"}
}

func (t *YAML) Unmarshal(stream io.Reader, v interface{}) error {
	data, err := ioutil.ReadAll(stream)
	if err != nil {
		return err
	}

	if t.Strict {
		return yaml.UnmarshalStrict(data, v)
	} else {
		return yaml.Unmarshal(data, v)
	}
}

func (t *YAML) Marshal(stream io.Writer, v interface{}) error {
	data, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	_, err = stream.Write(data)
	return err
}

type TOML struct {
	ArraysWithOneElementPerLine bool
	QuoteMapKeys                bool
}

func (t *TOML) Extensions() []string {
	return []string{"toml"}
}

func (t *TOML) Unmarshal(stream io.Reader, v interface{}) error {
	return toml.NewDecoder(stream).Decode(v)
}

func (t *TOML) Marshal(stream io.Writer, v interface{}) error {
	enc := toml.NewEncoder(stream)
	enc.ArraysWithOneElementPerLine(t.ArraysWithOneElementPerLine)
	enc.QuoteMapKeys(t.QuoteMapKeys)
	return enc.Encode(v)
}

type XML struct {
	Strict        bool
	AutoClose     []string
	Entity        map[string]string
	CharsetReader func(charset string, input io.Reader) (io.Reader, error)
	DefaultSpace  string

	Prefix string
	Indent string
}

func (t *XML) Extensions() []string {
	return []string{"xml"}
}

func (t *XML) Unmarshal(stream io.Reader, v interface{}) error {
	dec := xml.NewDecoder(stream)
	dec.Strict = t.Strict
	dec.AutoClose = t.AutoClose
	dec.Entity = t.Entity
	dec.CharsetReader = t.CharsetReader
	dec.DefaultSpace = t.DefaultSpace
	return dec.Decode(v)
}

func (t *XML) Marshal(stream io.Writer, v interface{}) error {
	enc := xml.NewEncoder(stream)
	enc.Indent(t.Prefix, t.Indent)
	return enc.Encode(v)
}
