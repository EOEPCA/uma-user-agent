package config

import (
	"github.com/sirupsen/logrus"
)

func init() {
	// Logger
	initLogger()

	// Load config from files
	configInit()

	// Trigger config change handlers
	go TriggerConfigChangeHandlers()
}

func initLogger() {
	// log format
	logFormatter := new(logrus.TextFormatter)
	logFormatter.TimestampFormat = "2006-01-02 15:04:05.000"
	logrus.SetFormatter(logFormatter)
	logFormatter.FullTimestamp = true
	// config change handler
	AddConfigChangeHandler(updateLoggerConfig)
}

func updateLoggerConfig() {
	logrus.SetLevel(GetLogLevel())
}
