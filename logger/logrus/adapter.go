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

package logrus

import (
	"github.com/sirupsen/logrus"
	"github.com/tamasd/constellation/logger"
)

type Logger struct {
	*logrus.Entry
}

func NewLogger(entry *logrus.Entry) Logger {
	return Logger{Entry: entry}
}

func (l Logger) WithField(key string, value interface{}) logger.Logger {
	return NewLogger(l.Entry.WithField(key, value))
}

func (l Logger) WithFields(fields logger.Fields) logger.Logger {
	return NewLogger(l.Entry.WithFields(logrus.Fields(fields)))
}

func (l Logger) WithError(err error) logger.Logger {
	return NewLogger(l.Entry.WithError(err))
}

func (l Logger) Level() logger.Level {
	return logger.Level(l.Entry.Level)
}
