package config

import log "github.com/sirupsen/logrus"

var Config = Configuration{}

func init() {
	// Load the config values
	Config.init()
	// Logger
	initLogger()
}

func initLogger() {
	// log level
	log.SetLevel(Config.LogLevel)

	// log format
	logFormatter := new(log.TextFormatter)
	logFormatter.TimestampFormat = "2006-01-02 15:04:05.000"
	log.SetFormatter(logFormatter)
	logFormatter.FullTimestamp = true
}
