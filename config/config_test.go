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

package config_test

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/tamasd/constellation/config"
	"github.com/tamasd/constellation/database"
	"github.com/tamasd/constellation/logger/null"
	"github.com/tamasd/constellation/util"
)

var (
	_ config.FileType         = &config.JSON{}
	_ config.FileType         = &config.YAML{}
	_ config.FileType         = &config.TOML{}
	_ config.FileType         = &config.XML{}
	_ config.CollectionLoader = &config.Directory{}
	_ config.CollectionLoader = &config.Database{}
	_ config.WritableProvider = &config.DatabaseConfigProvider{}
	_ config.WritableProvider = &errorProvider{}
)

type errorProvider struct {
	OnRead  bool
	OnWrite bool
}

func (p *errorProvider) CanSave(key string) bool {
	return true
}

func (p *errorProvider) Save(key string, v interface{}) error {
	if p.OnWrite {
		return errors.New("")
	}

	return nil
}

func (p *errorProvider) Has(key string) bool {
	return true
}

func (p *errorProvider) Unmarshal(key string, v interface{}) error {
	if p.OnRead {
		return errors.New("")
	}

	return nil
}

type test struct {
	A int
	B string
	C bool
	D struct {
		E int
		F float64
	}
	G string
}

func TestConfig(t *testing.T) {
	expected := testExample()

	c := config.NewStore(null.NewLogger())
	c.RegisterSchema("test.*", reflect.TypeOf(test{}))

	entries := []string{"test.0", "test.1", "test.2", "test.3"}
	for _, entry := range entries {
		_ = os.Setenv("CONFIG_"+strings.ToUpper(entry)+"_G", "zxcvbn")
	}

	defaultCollection := config.NewCollection()
	ep := config.NewEnvConfigProvider()
	ep.Prefix = "CONFIG"
	ep.Reset()
	dp := config.NewDirectoryConfigProvider("fixtures/config", true)
	registerFileTypes(dp)
	defaultCollection.AddProviders(ep, dp)
	c.AddCollection("config", defaultCollection)

	for _, entry := range entries {
		t.Run(entry, func(t *testing.T) {
			v, err := c.Get("config").Get(entry)
			require.NoError(t, err)
			require.Equal(t, expected, v)
		})
	}

	require.Nil(t, c.Get("test"))
}

func TestWritableConfig(t *testing.T) {
	entries := []config.FileType{
		&config.JSON{},
		&config.YAML{},
		&config.TOML{},
		&config.XML{},
	}

	for _, entry := range entries {
		t.Run(reflect.TypeOf(entry).Name(), func(t *testing.T) {
			ep := config.NewEnvConfigProvider()
			ep.Prefix = ""
			ep.Reset()

			tmpdir, err := ioutil.TempDir("", "constellationtest")
			require.NoError(t, err)
			defer func() { util.Must(os.RemoveAll(tmpdir)) }()

			dp := config.NewDirectoryConfigProvider("fixtures/config", true)
			registerFileTypes(dp)
			dpw := config.NewDirectoryConfigProvider(tmpdir, false)
			dpw.RegisterFiletype(entry)
			registerFileTypes(dpw)

			collection := config.NewCollection()
			collection.AddProviders(ep, dpw, dp)

			c := config.NewStore(null.NewLogger())
			c.RegisterSchema("test.*", reflect.TypeOf(test{}))
			c.AddCollection("config", collection)

			ft := saveValue(t, c, "qwer")
			checkValue(t, tmpdir, entry, ft)
			collection.ClearCache()

			ft = saveValue(t, c, "asdf")
			checkValue(t, tmpdir, entry, ft)
			collection.ClearCache()
		})
	}
}

func TestReadOnlyProviders(t *testing.T) {
	c := config.NewStore(null.NewLogger())
	c.RegisterSchema("test.*", reflect.TypeOf(test{}))

	collection := config.NewCollection()
	dp := config.NewDirectoryConfigProvider("fixtures/config", true)
	registerFileTypes(dp)
	collection.AddProviders(dp)

	c.AddCollection("config", collection)

	v, saver, err := c.GetWritable("config").GetWritable("test.0")
	data := v.(test)
	require.NoError(t, err)

	err = saver.Save(data)
	require.Error(t, err)
}

func TestConfigSaveError(t *testing.T) {
	c := config.NewStore(null.NewLogger())
	c.RegisterSchema("test.*", reflect.TypeOf(test{}))

	collection := config.NewCollection()
	collection.AddProviders(&errorProvider{
		OnRead:  false,
		OnWrite: true,
	})

	c.AddCollection("config", collection)

	v, saver, err := c.GetWritable("config").GetWritable("test.0")
	require.NoError(t, err)

	err = saver.Save(v)
	require.Error(t, err)
}

func TestCollectionLoaders(t *testing.T) {
	c := config.NewStore(null.NewLogger())
	c.RegisterSchema("test", reflect.TypeOf(test{}))

	mp := config.NewMemoryConfigProvider()
	collection := config.NewCollection()
	collection.SetTemporary(true)
	collection.AddProviders(mp)
	var loader = config.CollectionLoaderFunc(func(name string) (*config.Collection, error) {
		if name == "test" {
			return collection, nil
		}

		return nil, config.CollectionNotFoundError{Name: name}
	})
	c.AddCollectionLoaders(loader)

	clean := func() {
		collection.ClearCache()
		mp.Reset()
		c.RemoveTemporary()
	}

	t.Run("load collection and find value", func(t *testing.T) {
		clean()
		_, saver, err := c.GetWritable("test").GetWritable("test")
		require.NoError(t, err)
		err = saver.Save(testExample())
		require.NoError(t, err)

		collection.ClearCache()

		res, err := c.Get("test").Get("test")
		require.NoError(t, err)
		require.Equal(t, testExample(), res)
	})

	t.Run("do not find value that does not exists", func(t *testing.T) {
		clean()
		res, err := c.Get("test").Get("test")
		require.Nil(t, res)
		require.NoError(t, err)
	})
}

func TestSchemaCannotBeAddedTwice(t *testing.T) {
	c := config.NewStore(null.NewLogger())
	c.RegisterSchema("test", reflect.TypeOf(test{}))

	require.NotPanics(t, func() {
		c.RegisterSchema("test", reflect.TypeOf(test{}))
	})
	require.Panics(t, func() {
		c.RegisterSchema("test", reflect.TypeOf(struct{}{}))
	})
}

func TestCollectionLoader(t *testing.T) {
	conf := config.NewStore(null.NewLogger())
	cl := config.NewDirectory(".", map[string]string{
		"test": "fixtures",
	}, true)
	conf.RegisterSchema("test", reflect.TypeOf(test{}))
	conf.AddCollectionLoaders(cl)

	t.Run("find configuration file", func(t *testing.T) {
		testInterface, err := conf.Get("test").Get("test")
		require.NoError(t, err)
		require.NotNil(t, testInterface)
		testData := testInterface.(test)
		require.Equal(t, 5, testData.A)
		require.Equal(t, "asdf", testData.B)
	})

	t.Run("return nil when directory does not exists", func(t *testing.T) {
		testInterface := conf.Get("asdf")
		require.Nil(t, testInterface)
	})

	t.Run("return nil when it is not a directory", func(t *testing.T) {
		testInterface := conf.Get("config_test.go")
		require.Nil(t, testInterface)
	})
}

func TestDatabaseCollectionLoader(t *testing.T) {
	conf := config.NewStore(null.NewLogger())

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		t.Skip("no database provided")
	}
	conn, cleanup := database.TestConnect(dbUrl)
	t.Cleanup(cleanup)

	cl := config.NewDatabase(conn, false)
	conf.RegisterSchema("test", reflect.TypeOf(test{}))
	conf.AddCollectionLoaders(cl)

	ns0 := util.RandomHexString(12)

	t.Run("create the schema and save test data", func(t *testing.T) {
		_, err := cl.Migrations().Migrations().UpgradeFrom(-1, null.NewLogger(), conn)
		require.NoError(t, err)

		_, err = conn.Exec(`INSERT INTO namespace(namespace) VALUES($1)`, ns0)
		require.NoError(t, err)

		_, err = conn.Exec(`INSERT INTO config(namespace, name, value) VALUES($1, $2, $3)`, ns0, "test", util.JSONString(testExample()))
		require.NoError(t, err)
	})

	t.Run("load test data from the database", func(t *testing.T) {
		v, err := conf.Get(ns0).Get("test")
		require.NoError(t, err)
		require.Equal(t, testExample(), v)
	})
}

func testExample() test {
	example := test{
		A: 5,
		B: "asdf",
		C: true,
		G: "zxcvbn",
	}
	example.D.E = -2
	example.D.F = -1.2

	return example
}

func checkValue(t *testing.T, tmpdir string, ft config.FileType, testData test) {
	fn := filepath.FromSlash(path.Join(tmpdir, "test.0")) + "." + ft.Extensions()[0]
	f, err := os.Open(fn)
	require.NoError(t, err)
	defer func() { _ = f.Close() }()
	ut := test{}
	_ = ft.Unmarshal(f, &ut)
	require.Equal(t, testData, ut)
}

func saveValue(t *testing.T, c *config.Store, value string) test {
	v, saver, err := c.GetWritable("config").GetWritable("test.0")
	require.NoError(t, err)
	data := v.(test)
	data.G = value
	err = saver.Save(data)
	require.NoError(t, err)
	return data
}

func registerFileTypes(dp *config.DirectoryConfigProvider) {
	dp.RegisterFiletype(&config.YAML{})
	dp.RegisterFiletype(&config.JSON{})
	dp.RegisterFiletype(&config.TOML{})
	dp.RegisterFiletype(&config.XML{})
}
