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
	"bufio"
	"errors"
	"io"
	"strings"
)

var (
	newlineByte = []byte("\n")
)

type Header struct {
	kv map[string]string
}

func NewHeader() *Header {
	return &Header{
		kv: map[string]string{},
	}
}

func ParseHeaderFrom(r io.Reader) (*Header, error) {
	h := NewHeader()

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		if line == "" {
			break
		}

		separator := strings.IndexByte(line, ':')
		if separator < 0 {
			return nil, errors.New("invalid line")
		}

		if err := h.Set(line[:separator], line[separator+1:]); err != nil {
			return nil, err
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return h, nil
}

func (h *Header) Len() int {
	return len(h.kv)
}

func (h *Header) Get(key string) string {
	return h.kv[key]
}

func (h *Header) Set(key, value string) error {
	// TODO check for \n and :
	h.kv[key] = value
	return nil
}

func (h *Header) Delete(key string) {
	delete(h.kv, key)
}

func (h *Header) WriteTo(w io.Writer) (int64, error) {
	var written int64

	for k, v := range h.kv {
		n, err := io.WriteString(w, k+":"+v+"\n")
		written += int64(n)

		if err != nil {
			return written, err
		}
	}

	n, err := w.Write(newlineByte)
	written += int64(n)

	return written, err
}
