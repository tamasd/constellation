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
	"bytes"
	"errors"
	"io"
	"sync"
)

const (
	TypeHeaderName        = "type"
	MessageTypeHeaderName = "mtyp"

	JsonType = "json"
	GobType  = "gob"
	XmlType  = "xml"
)

var (
	ErrMessageTooSmall    = errors.New("message is too small")
	ErrInvalidMessageType = errors.New("invalid message type")
)

var (
	bufPool = sync.Pool{New: func() interface{} {
		buf := bytes.NewBuffer(make([]byte, 2*1024*1024))
		buf.Reset()

		return buf
	}}
)

type Serializer struct {
	registry    *Registry
	codec       *Codec
	defaultType string
}

func NewSerializer(registry *Registry, codec *Codec, defaultType string) *Serializer {
	return &Serializer{
		registry:    registry,
		codec:       codec,
		defaultType: defaultType,
	}
}

func (s *Serializer) Parse(msg []byte) (*Envelope, error) {
	split := findSplit(msg[:])
	if split < 1 {
		return nil, ErrMessageTooSmall
	}

	header, err := ParseHeaderFrom(bytes.NewBuffer(msg[:split]))
	if err != nil {
		return nil, err
	}

	v := s.registry.Create(header.Get(MessageTypeHeaderName))
	if v == nil {
		return nil, ErrInvalidMessageType
	}

	if err = s.getDecoder(header.Get(TypeHeaderName), bytes.NewBuffer(msg[split:])).Decode(v); err != nil {
		return nil, err
	}

	return NewEnvelope(header, v), nil
}

func (s *Serializer) Serialize(e *Envelope) ([]byte, error) {
	buf := bufPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufPool.Put(buf)
	}()

	if e.Header().Get(MessageTypeHeaderName) == "" {
		if err := e.header.Set(MessageTypeHeaderName, s.registry.MessageType(e.body)); err != nil {
			return nil, err
		}
	}

	if e.Header().Get(TypeHeaderName) == "" {
		if err := e.header.Set(TypeHeaderName, s.defaultType); err != nil {
			return nil, err
		}
	}

	_, err := e.Header().WriteTo(buf)
	if err != nil {
		return nil, err
	}

	if err = s.getEncoder(e.Header().Get(TypeHeaderName), buf).Encode(e.Body()); err != nil {
		return nil, err
	}

	result := make([]byte, buf.Len())
	copy(result, buf.Bytes())

	return result, nil
}

func (s *Serializer) getEncoder(t string, w io.Writer) Encoder {
	enc := s.codec.Encoder(t, w)

	if enc == nil {
		return s.codec.Encoder(s.defaultType, w)
	}

	return enc
}

func (s *Serializer) getDecoder(t string, r io.Reader) Decoder {
	dec := s.codec.Decoder(t, r)

	if dec == nil {
		return s.codec.Decoder(s.defaultType, r)
	}

	return dec
}

func findSplit(data []byte) int {
	if len(data) == 0 {
		return -1
	}

	for i, max := 1, len(data); i < max; i++ {
		if data[i-1] == '\n' && data[i] == '\n' {
			return i + 1
		}
	}

	return -1
}
