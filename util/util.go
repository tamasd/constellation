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

package util

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap   = regexp.MustCompile("([a-z0-9])([A-Z])")

	humanizerReplacer = strings.NewReplacer(
		"-", " ",
		"_", " ",
	)
)

// ToSnakeCase converts camel case to snake case.
func ToSnakeCase(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

// SetContext sets a value in a request's context.
func SetContext(r *http.Request, key, value interface{}) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), key, value))
}

// RandomHexString returns a random hex string with a given string length.
func RandomHexString(length int) string {
	buflen := length / 2

	if length%2 == 1 {
		buflen++
	}

	// TODO sync.Pool?
	buf := make([]byte, buflen)

	_, err := io.ReadFull(rand.Reader, buf)
	Must(err)

	return hex.EncodeToString(buf)[:length]
}

// GeneratePlaceholders generates placeholders for an SQL query.
func GeneratePlaceholders(start, length int) string {
	if length == 0 {
		return ""
	}

	var str string
	for i := 0; i < length; i++ {
		str += ", $" + strconv.Itoa(i+start)
	}

	return str[2:]
}

func HumanizeIdentifier(id string) string {
	if len(id) == 0 {
		return ""
	}

	id = humanizerReplacer.Replace(id)
	id = strings.ToLower(id)
	return string(unicode.ToUpper(rune(id[0]))) + id[1:]
}

func CloneRequestUrl(r *http.Request) *url.URL {
	u := &url.URL{
		Host:     r.Host,
		Path:     r.URL.Path,
		RawQuery: r.URL.RawQuery,
	}

	return u
}

func WithUrlQuery(u *url.URL, with func(v url.Values)) {
	v := u.Query()
	with(v)
	u.RawQuery = v.Encode()
}

// DisableHTTPClientCache sets headers that tell the client not to cache.
//
// This is a temporary(?) solution until caching is not evaluated.
func DisableHTTPClientCache(header http.Header) {
	header.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	header.Set("Pragma", "no-cache")
	header.Set("Expires", "0")
}

func JSONString(v interface{}) string {
	if v == nil {
		return "{}"
	}

	b, _ := json.Marshal(v)
	return string(b)
}

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustClose(c io.Closer) {
	Must(c.Close())
}
