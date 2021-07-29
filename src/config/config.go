package config

import (
	"os"
	"strconv"
)

type Configuration struct {
	Port             int
	UserIdCookieName string
}

var Config = Configuration{}

var defaults = Configuration{
	Port:             80,
	UserIdCookieName: "auth_user_id",
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
	config.Port = getPort()
	config.UserIdCookieName = getUserIdCookieName()
}

func init() {
	Config.init()
}
