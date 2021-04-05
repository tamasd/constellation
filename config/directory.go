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
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/tamasd/constellation/util"
)

type Directory struct {
	base     string
	conf     map[string]string
	readOnly bool
}

func NewDirectory(base string, conf map[string]string, readOnly bool) *Directory {
	return &Directory{
		base:     base,
		conf:     conf,
		readOnly: readOnly,
	}
}

func (d *Directory) Load(name string) (*Collection, error) {
	if alias, found := d.conf[name]; found {
		name = alias
	}

	dir := filepath.Join(d.base, name)

	info, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, CollectionNotFoundError{Name: dir}
	}

	c := NewCollection()
	c.SetTemporary(true)

	e := NewEnvConfigProvider()
	e.Prefix = "NS_" + strings.ToUpper(name)

	p := NewDirectoryConfigProvider(dir, d.readOnly)
	p.RegisterFiletype(&JSON{})
	p.RegisterFiletype(&YAML{})
	p.RegisterFiletype(&TOML{})
	p.RegisterFiletype(&XML{})

	c.AddProviders(e, p)

	return c, nil
}

var _ WritableProvider = &DirectoryConfigProvider{}

type FileType interface {
	Extensions() []string
	Unmarshal(stream io.Reader, v interface{}) error
	Marshal(stream io.Writer, v interface{}) error
}

type DirectoryConfigProvider struct {
	base      string
	readOnly  bool
	fileTypes []FileType
}

func NewDirectoryConfigProvider(base string, readOnly bool) *DirectoryConfigProvider {
	return &DirectoryConfigProvider{
		base:     base,
		readOnly: readOnly,
	}
}

func (d *DirectoryConfigProvider) RegisterFiletype(t FileType) {
	d.fileTypes = append(d.fileTypes, t)
}

func (d *DirectoryConfigProvider) basenameForKey(key string) string {
	return filepath.FromSlash(filepath.Join(d.base, key))
}

func (d *DirectoryConfigProvider) exists(key string) (FileType, string) {
	name := d.basenameForKey(key)
	for _, t := range d.fileTypes {
		for _, ext := range t.Extensions() {
			fn := name + "." + ext
			if _, err := os.Stat(fn); err == nil {
				return t, fn
			}
		}
	}

	return nil, ""
}

func (d *DirectoryConfigProvider) Has(key string) bool {
	_, fn := d.exists(key)
	return fn != ""
}

func (d *DirectoryConfigProvider) Unmarshal(key string, v interface{}) error {
	ft, fn := d.exists(key)
	f, err := os.Open(fn)
	if err != nil {
		return err
	}
	defer util.MustClose(f)

	return ft.Unmarshal(f, v)
}

func (d *DirectoryConfigProvider) CanSave(_ string) bool {
	return !d.readOnly
}

func (d *DirectoryConfigProvider) Save(key string, v interface{}) error {
	var f *os.File
	var err error

	ft, fn := d.exists(key)
	if fn == "" { // file does not exists
		if len(d.fileTypes) == 0 {
			return errors.New("no configured file type for this directory config provider")
		}
		name := d.basenameForKey(key) + "." + d.fileTypes[0].Extensions()[0]
		ft = d.fileTypes[0]
		f, err = os.Create(name)
	} else { // file exists
		f, err = os.OpenFile(fn, os.O_RDWR, 0)
	}
	if err != nil {
		return err
	}
	defer util.MustClose(f)

	return ft.Marshal(f, v)
}
