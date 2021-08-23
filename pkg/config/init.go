package config

import log "github.com/sirupsen/logrus"

func init() {
	// NEW - file-based config approach
	viperInit()

	// Logger
	initLogger()
}

func initLogger() {
	// log level
	log.SetLevel(GetLogLevel())

	// log format
	logFormatter := new(log.TextFormatter)
	logFormatter.TimestampFormat = "2006-01-02 15:04:05.000"
	log.SetFormatter(logFormatter)
	logFormatter.FullTimestamp = true
}
