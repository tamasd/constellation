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

package matcher

import (
	"strings"
)

type Matcher struct {
	separator string
	tree      *item
}

func NewMatcher(separator string) *Matcher {
	return &Matcher{
		separator: separator,
		tree:      newItem(),
	}
}

func (m *Matcher) get(path string, create bool) *item {
	parts := strings.Split(path, m.separator)
	return m.tree.get(parts, create)
}

func (m *Matcher) Get(path string) interface{} {
	item := m.get(path, false)
	if item == nil {
		return nil
	}

	return item.content
}

func (m *Matcher) Set(path string, content interface{}) {
	item := m.get(path, true)
	item.content = content
}

type item struct {
	children map[string]*item
	wildcard *item
	content  interface{}
}

func newItem() *item {
	return &item{
		children: make(map[string]*item),
		wildcard: nil,
	}
}

func (i *item) get(path []string, create bool) *item {
	if len(path) == 0 {
		return i
	}

	current := path[0]
	if current == "*" {
		if i.wildcard != nil {
			return i.wildcard.get(path[1:], create)
		}
		if create {
			i.wildcard = newItem()
			return i.wildcard.get(path[1:], create)
		}
	}

	if childItem, found := i.children[current]; found {
		return childItem.get(path[1:], create)
	}

	if create {
		i.children[current] = newItem()
		return i.children[current].get(path[1:], create)
	}

	if i.wildcard != nil {
		return i.wildcard.get(path[1:], create)
	}

	return nil
}
