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
