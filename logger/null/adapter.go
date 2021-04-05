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

package null

import (
	"github.com/tamasd/constellation/logger"
)

type Logger struct {
}

func NewLogger() *Logger {
	return nil
}

func (l *Logger) WithField(_ string, _ interface{}) logger.Logger {
	return l
}

func (l *Logger) WithFields(_ logger.Fields) logger.Logger {
	return l
}

func (l *Logger) WithError(_ error) logger.Logger {
	return l
}

func (l *Logger) Debugf(_ string, _ ...interface{}) {
}

func (l *Logger) Infof(_ string, _ ...interface{}) {
}

func (l *Logger) Printf(_ string, _ ...interface{}) {
}

func (l *Logger) Warnf(_ string, _ ...interface{}) {
}

func (l *Logger) Warningf(_ string, _ ...interface{}) {
}

func (l *Logger) Errorf(_ string, _ ...interface{}) {
}

func (l *Logger) Fatalf(_ string, _ ...interface{}) {
}

func (l *Logger) Panicf(_ string, _ ...interface{}) {
}

func (l *Logger) Debug(_ ...interface{}) {
}

func (l *Logger) Info(_ ...interface{}) {
}

func (l *Logger) Print(_ ...interface{}) {
}

func (l *Logger) Warn(_ ...interface{}) {
}

func (l *Logger) Warning(_ ...interface{}) {
}

func (l *Logger) Error(_ ...interface{}) {
}

func (l *Logger) Fatal(_ ...interface{}) {
}

func (l *Logger) Panic(_ ...interface{}) {
}

func (l *Logger) Debugln(_ ...interface{}) {
}

func (l *Logger) Infoln(_ ...interface{}) {
}

func (l *Logger) Println(_ ...interface{}) {
}

func (l *Logger) Warnln(_ ...interface{}) {
}

func (l *Logger) Warningln(_ ...interface{}) {
}

func (l *Logger) Errorln(_ ...interface{}) {
}

func (l *Logger) Fatalln(_ ...interface{}) {
}

func (l *Logger) Panicln(_ ...interface{}) {
}

func (l *Logger) Level() logger.Level {
	return logger.ErrorLevel
}
