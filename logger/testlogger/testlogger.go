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

package testlogger

import (
	"bytes"

	logrusorig "github.com/sirupsen/logrus"
	"github.com/tamasd/constellation/logger"
	"github.com/tamasd/constellation/logger/logrus"
)

type Logger struct {
	logger.Logger
	Buffer *bytes.Buffer
}

func NewLogger(logger logger.Logger, buf *bytes.Buffer) *Logger {
	return &Logger{
		Logger: logger,
		Buffer: buf,
	}
}

func TestLogger() *Logger {
	buf := bytes.NewBuffer(nil)
	l := logrusorig.New()
	l.Out = buf
	l.Level = logrusorig.TraceLevel

	return NewLogger(logrus.NewLogger(l.WithFields(nil)), buf)
}
