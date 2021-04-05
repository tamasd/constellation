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

type Envelope struct {
	header *Header
	body   interface{}
}

func NewEnvelope(header *Header, body interface{}) *Envelope {
	return &Envelope{
		header: header,
		body:   body,
	}
}

func (e *Envelope) Header() *Header {
	return e.header
}

func (e *Envelope) Body() interface{} {
	return e.body
}
