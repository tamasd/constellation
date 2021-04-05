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
	"encoding/gob"
	"encoding/json"
	"encoding/xml"
	"io"
)

type Encoder interface {
	Encode(v interface{}) error
}

type Decoder interface {
	Decode(v interface{}) error
}

type EncoderFactory interface {
	NewEncoder(w io.Writer) Encoder
}

type DecoderFactory interface {
	NewDecoder(r io.Reader) Decoder
}

type Codec struct {
	encoders map[string]EncoderFactory
	decoders map[string]DecoderFactory
}

func NewCodec() *Codec {
	return &Codec{
		encoders: map[string]EncoderFactory{},
		decoders: map[string]DecoderFactory{},
	}
}

func (c *Codec) AddEncoder(contentType string, encoder EncoderFactory) {
	c.encoders[contentType] = encoder
}

func (c *Codec) AddDecoder(contentType string, decoder DecoderFactory) {
	c.decoders[contentType] = decoder
}

func (c *Codec) Encoder(contentType string, w io.Writer) Encoder {
	factory := c.encoders[contentType]
	if factory != nil {
		return factory.NewEncoder(w)
	}

	return nil
}

func (c *Codec) Decoder(contentType string, r io.Reader) Decoder {
	factory := c.decoders[contentType]
	if factory != nil {
		return factory.NewDecoder(r)
	}

	return nil
}

type Json struct {
}

func NewJson() *Json {
	return &Json{}
}

func (j *Json) NewDecoder(r io.Reader) Decoder {
	return json.NewDecoder(r)
}

func (j *Json) NewEncoder(w io.Writer) Encoder {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	return enc
}

type Gob struct {
}

func NewGob() *Gob {
	return &Gob{}
}

func (g *Gob) NewDecoder(r io.Reader) Decoder {
	return gob.NewDecoder(r)
}

func (g *Gob) NewEncoder(w io.Writer) Encoder {
	return gob.NewEncoder(w)
}

type Xml struct {
}

func NewXml() *Xml {
	return &Xml{}
}

func (x *Xml) NewDecoder(r io.Reader) Decoder {
	return xml.NewDecoder(r)
}

func (x *Xml) NewEncoder(w io.Writer) Encoder {
	return xml.NewEncoder(w)
}
