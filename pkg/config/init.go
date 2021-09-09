package config

import "github.com/sirupsen/logrus"

func init() {
	// Logger
	initLogger()

	// Load config from files
	configInit()

	// log level - from config-specified value
	logrus.SetLevel(GetLogLevel())
}

func initLogger() {
	// log format
	logFormatter := new(logrus.TextFormatter)
	logFormatter.TimestampFormat = "2006-01-02 15:04:05.000"
	logrus.SetFormatter(logFormatter)
	logFormatter.FullTimestamp = true
}
