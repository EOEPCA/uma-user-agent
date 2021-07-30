package config

import (
	"fmt"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type Configuration struct {
	Port             int
	UserIdCookieName string
	LogLevel         log.Level
}

var defaults = Configuration{
	Port:             80,
	UserIdCookieName: "auth_user_id",
	LogLevel:         log.InfoLevel,
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

func (config *Configuration) init() {
	config.Port = getPort()
	config.UserIdCookieName = getUserIdCookieName()
	config.LogLevel = getLogLevel()
}
