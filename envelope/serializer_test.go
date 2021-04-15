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

package envelope_test

import (
	"bytes"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tamasd/constellation/envelope"
)

func TestSerializationDeserialization(t *testing.T) {
	table := []struct {
		name    string
		d       *data
		headers []string
		body    []byte
		typ     string
	}{
		{
			"json",
			&data{5, "x"},
			[]string{"mtyp:data", "type:json"},
			[]byte("{\"A\":5,\"B\":\"x\"}\n"),
			envelope.JsonType,
		},
		{
			"xml",
			&data{5, "x"},
			[]string{"mtyp:data", "type:xml"},
			[]byte("<data><A>5</A><B>x</B></data>"),
			envelope.XmlType,
		},
		{
			"gob",
			&data{5, "x"},
			[]string{"mtyp:data", "type:gob"},
			[]byte{0x1e, 0xff, 0x81, 0x3, 0x1, 0x1, 0x4, 0x64, 0x61, 0x74, 0x61, 0x1, 0xff, 0x82, 0x0, 0x1, 0x2, 0x1, 0x1, 0x41, 0x1, 0x4, 0x0, 0x1, 0x1, 0x42, 0x1, 0xc, 0x0, 0x0, 0x0, 0x8, 0xff, 0x82, 0x1, 0xa, 0x1, 0x1, 0x78, 0x0},
			envelope.GobType,
		},
	}

	s := serializer()

	for _, row := range table {
		t.Run(row.name, func(t *testing.T) {
			h := envelope.NewHeader()
			err := h.Set(envelope.TypeHeaderName, row.typ)
			require.Nil(t, err)
			e := envelope.NewEnvelope(h, row.d)
			serialized, err := s.Serialize(e)
			require.Nil(t, err)

			headers, body := parseSerialized(t, serialized)
			require.Equal(t, row.headers, headers)
			require.Equal(t, row.body, body)

			ed, err := s.Parse(append([]byte(strings.Join(row.headers, "\n")+"\n\n"), row.body...))
			require.Nil(t, err)

			require.Equal(t, e, ed)
		})
	}
}

func parseSerialized(t *testing.T, serialized []byte) ([]string, []byte) {
	idx := bytes.Index(serialized, []byte{0xa, 0xa})
	require.GreaterOrEqual(t, idx, 0)

	headerBytes, body := serialized[:idx], serialized[idx+2:]
	header := strings.Split(string(headerBytes), "\n")

	sort.Strings(header)

	return header, body
}

func serializer() *envelope.Serializer {
	registry := envelope.NewRegistry()
	registry.Register("data", reflect.TypeOf(data{}))

	codec := envelope.NewCodec()

	js := envelope.NewJson()
	codec.AddEncoder(envelope.JsonType, js)
	codec.AddDecoder(envelope.JsonType, js)

	gob := envelope.NewGob()
	codec.AddEncoder(envelope.GobType, gob)
	codec.AddDecoder(envelope.GobType, gob)

	xml := envelope.NewXml()
	codec.AddEncoder(envelope.XmlType, xml)
	codec.AddDecoder(envelope.XmlType, xml)

	return envelope.NewSerializer(registry, codec, envelope.JsonType)
}

type data struct {
	A int
	B string
}
