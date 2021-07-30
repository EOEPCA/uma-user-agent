package config

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type Configuration struct {
	LogLevel         log.Level
	PepUrl           string
	Port             int
	UserIdCookieName string
}

var defaults = Configuration{
	LogLevel:         log.InfoLevel,
	PepUrl:           "http://pep",
	Port:             80,
	UserIdCookieName: "auth_user_id",
}

func getLogLevel() log.Level {
	logLevel := defaults.LogLevel
	val, ok := os.LookupEnv("LOG_LEVEL")
	if ok {
		if len(val) > 0 {
			l, err := log.ParseLevel(val)
			if err != nil {
				log.Warning(fmt.Sprintf("Bad log level '%v' specified, using default '%v'", val, logLevel.String()))
			} else {
				logLevel = l
			}
		}
	}
	log.Info("Using log level: ", logLevel.String())
	return logLevel
}

func getPepUrl() string {
	url := defaults.PepUrl
	val, ok := os.LookupEnv("PEP_URL")
	if ok {
		if len(val) > 0 {
			url = val
		}
	}
	return url
}

func getPort() int {
	port := defaults.Port
	val, ok := os.LookupEnv("PORT")
	if ok {
		if len(val) > 0 {
			ival, err := strconv.Atoi(val)
			if err == nil {
				port = ival
			}
		}
	}
	return port
}

func getUserIdCookieName() string {
	userIdCookieName := defaults.UserIdCookieName
	val, ok := os.LookupEnv("USER_ID_COOKIE_NAME")
	if ok {
		if len(val) > 0 {
			userIdCookieName = val
		}
	}
	return userIdCookieName
}

func (config *Configuration) init() {
	config.LogLevel = getLogLevel()
	config.PepUrl = getPepUrl()
	config.Port = getPort()
	config.UserIdCookieName = getUserIdCookieName()
}
