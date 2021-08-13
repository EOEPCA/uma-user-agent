package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
)

type Configuration struct {
	ClientId         string
	ClientSecret     string
	HttpTimeout      time.Duration
	LogLevel         log.Level
	PepUrl           string
	Port             int
	UserIdCookieName string
}

var defaults = Configuration{
	ClientId:         "",
	ClientSecret:     "",
	HttpTimeout:      10,
	LogLevel:         log.InfoLevel,
	PepUrl:           "http://pep",
	Port:             80,
	UserIdCookieName: "auth_user_id",
}

func (config *Configuration) init() {
	log.Info("Initialising the configuration")
	config.ClientId = getClientId()
	config.ClientSecret = getClientSecret()
	config.HttpTimeout = getHttpTimeout()
	config.LogLevel = getLogLevel()
	config.PepUrl = getPepUrl()
	config.Port = getPort()
	config.UserIdCookieName = getUserIdCookieName()
}

func (config *Configuration) ensureReady() {
	interval := 10
	go func() {
		for !config.IsReady() {
			log.Warnf("Config is not ready; retrying in %v seconds", interval)
			time.Sleep(time.Second * time.Duration(interval))
			config.init()
		}
		log.Info("Config is READY")
	}()
}

func (config *Configuration) IsReady() (isReady bool) {
	isReady = true &&
		len(config.ClientId) > 0 &&
		len(config.ClientSecret) > 0
	return
}

func getClientId() string {
	clientId := defaults.ClientId
	val, ok := os.LookupEnv("CLIENT_ID")
	if ok {
		if len(val) > 0 {
			clientId = val
		}
	}
	return clientId
}

func getClientSecret() string {
	clientSecret := defaults.ClientId
	val, ok := os.LookupEnv("CLIENT_SECRET")
	if ok {
		if len(val) > 0 {
			clientSecret = val
		}
	}
	return clientSecret
}

func getHttpTimeout() time.Duration {
	timeout := defaults.HttpTimeout
	val, ok := os.LookupEnv("HTTP_TIMEOUT")
	if ok {
		if len(val) > 0 {
			ival, err := strconv.Atoi(val)
			if err == nil {
				timeout = time.Duration(ival)
			}
		}
	}
	return timeout
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
