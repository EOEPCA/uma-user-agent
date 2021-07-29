package config

import (
	"os"
	"strconv"
)

type Configuration struct {
	Port int
}

var Config = Configuration{}

var defaults = Configuration{
	Port: 80,
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

func (config *Configuration) init() {
	config.Port = getPort()
}

func init() {
	Config.init()
}
