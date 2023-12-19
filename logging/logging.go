// This file is part of go-auto-wlan.
//
// go-auto-wlan is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-auto-wlan is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-auto-wlan. If not, see <http://www.gnu.org/licenses/>.
//
// Copyright 2023 Manuel Koch
package logging

import (
	"fmt"
	"io"
	"os"
	"path"

	log "github.com/sirupsen/logrus"
)

// ConfigureLogging will setup logging
// to log using named log level
// to log to optional path.
func ConfigueLogging(jsonFormat bool, logLevel string, logPath string) {
	if jsonFormat {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{DisableLevelTruncation: true, PadLevelText: true})
	}

	level, err := log.ParseLevel(logLevel)
	if err != nil {
		log.Fatal(fmt.Sprintf("Invalid log level '%s', use 'WARN', 'INFO' or 'DEBUG'\n", level))
	}

	var writer io.Writer = os.Stdout
	if len(logPath) > 0 {
		logDir := path.Dir(logPath)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			if err := os.MkdirAll(logDir, 0777); err != nil {
				log.Fatal(fmt.Sprintf("Failed to create log directory: %v\n", err))
			}
		}
		if logFile, err := os.Create(logPath); err != nil {
			log.Fatal(fmt.Sprintf("Failed to create log file: %v\n", err))
		} else {
			writer = logFile
		}
	}

	log.SetLevel(level)
	log.SetOutput(writer)
}
