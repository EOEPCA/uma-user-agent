package main

import (
	log "github.com/sirupsen/logrus"
)

func init() {
	// log level
	log.SetLevel(log.TraceLevel)

	// log format
	logFormatter := new(log.TextFormatter)
	logFormatter.TimestampFormat = "2006-01-02 15:04:05.000"
	log.SetFormatter(logFormatter)
	logFormatter.FullTimestamp = true
}
