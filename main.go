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
package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/manuel-koch/go-auto-wlan/app"
	"github.com/manuel-koch/go-auto-wlan/logging"
	log "github.com/sirupsen/logrus"
)

var (
	// these vars will be set on build time
	versionTag  string
	versionSha1 string
	buildDate   string

	logLevel string
	logPath  string
)

func main() {
	flag.StringVar(&logLevel, "log-level", "INFO", "Select the log level: DEBUG, INFO, WARN")
	flag.StringVar(&logPath, "log-path", "", "Log to file at given path")
	flag.Parse()

	logging.ConfigueLogging(false, logLevel, logPath)

	app := app.NewApp(versionTag, versionSha1, buildDate)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Info("Received shutdown signal")
		app.Shutdown()
	}()

	app.Run()
}
