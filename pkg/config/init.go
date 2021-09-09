package config

import log "github.com/sirupsen/logrus"

func init() {
	// Logger
	initLogger()

	// NEW - file-based config approach
	viperInit()

	// log level - from config-specified value
	log.SetLevel(GetLogLevel())
}

func initLogger() {
	// log format
	logFormatter := new(log.TextFormatter)
	logFormatter.TimestampFormat = "2006-01-02 15:04:05.000"
	log.SetFormatter(logFormatter)
	logFormatter.FullTimestamp = true
}
