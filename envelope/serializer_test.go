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
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tamasd/constellation/envelope"
)

func TestSerializationDeserialization(t *testing.T) {
	table := []struct {
		d          *data
		serialized []byte
		typ        string
	}{
		{&data{5, "x"}, []byte("type:json\n\n{\"A\":5,\"B\":\"x\"}\n"), envelope.JsonType},
		{&data{5, "x"}, []byte("type:xml\n\n<data><A>5</A><B>x</B></data>"), envelope.XmlType},
		{&data{5, "x"}, []byte{0x74, 0x79, 0x70, 0x65, 0x3a, 0x67, 0x6f, 0x62, 0xa, 0xa, 0x1e, 0xff, 0x81, 0x3, 0x1, 0x1, 0x4, 0x64, 0x61, 0x74, 0x61, 0x1, 0xff, 0x82, 0x0, 0x1, 0x2, 0x1, 0x1, 0x41, 0x1, 0x4, 0x0, 0x1, 0x1, 0x42, 0x1, 0xc, 0x0, 0x0, 0x0, 0x8, 0xff, 0x82, 0x1, 0xa, 0x1, 0x1, 0x78, 0x0}, envelope.GobType},
	}

	s := serializer()

	for _, row := range table {
		h := envelope.NewHeader()
		err := h.Set(envelope.TypeHeaderName, row.typ)
		require.Nil(t, err)
		e := envelope.NewEnvelope(h, row.d)
		serialized, err := s.Serialize(e)
		require.Nil(t, err)

		require.Equal(t, row.serialized, serialized)

		deserializedData := &data{}
		ed, err := s.Parse(row.serialized, deserializedData)
		require.Nil(t, err)

		require.Equal(t, e, ed)
	}
}

func serializer() *envelope.Serializer {
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

	return envelope.NewSerializer(codec, envelope.JsonType)
}

type data struct {
	A int
	B string
}
