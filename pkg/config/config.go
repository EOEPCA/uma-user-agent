package config

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

func IsReady() (isReady bool) {
	isReady = true &&
		len(GetClientId()) > 0 &&
		len(GetClientSecret()) > 0
	return
}

func GetClientId() string {
	return clientConfig.GetString(keyClientId.key)
}

func GetClientSecret() string {
	return clientConfig.GetString(keyClientSecret.key)
}

func GetHttpTimeout() time.Duration {
	return appConfig.GetDuration(keyHttpTimeout.key)
}

func GetLogLevel() log.Level {
	// default
	var logLevel log.Level
	var ok bool
	if logLevel, ok = keyLoggingLevel.defval.(log.Level); !ok {
		logLevel = log.InfoLevel
	}

	// read from config
	val := appConfig.GetString(keyLoggingLevel.key)
	l, err := log.ParseLevel(val)
	if err != nil {
		log.Warning(fmt.Sprintf("Bad log level '%v' specified, using default '%v'", val, logLevel.String()))
	} else {
		logLevel = l
	}

	log.Info("Using log level: ", logLevel.String())
	return logLevel
}

func GetPepUrl() string {
	return appConfig.GetString(keyPepUrl.key)
}

func GetPort() int {
	return appConfig.GetInt(keyListenPort.key)
}

func GetUserIdCookieName() string {
	return appConfig.GetString(keyUserIdCookieName.key)
}

func GetUnauthorizedResponse() string {
	return appConfig.GetString(keyUnauthorizedResponse.key)
}
