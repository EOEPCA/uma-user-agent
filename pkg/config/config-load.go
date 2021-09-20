package config

import (
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var clientConfig = viper.New()
var appConfig = viper.New()

const defaultConfigDir = "/app/config"

type configKey struct {
	key    string
	defval interface{}
}

// Config keys with default values
var keyClientId = configKey{"client-id", ""}
var keyClientSecret = configKey{"client-secret", ""}
var keyLoggingLevel = configKey{"logging.level", logrus.InfoLevel}
var keyHttpTimeout = configKey{"network.httpTimeout", 10}
var keyListenPort = configKey{"network.listenPort", 80}
var keyPepUrl = configKey{"pep.url", "http://pep"}
var keyUserIdCookieName = configKey{"userIdCookieName", "auth_user_id"}
var keyUnauthorizedResponse = configKey{"unauthorizedResponse", "Please login to access the resource"}
var keyOpenAccess = configKey{"openAccess", false}

// Client config
var clientConfigKeys = []configKey{keyClientId, keyClientSecret}

// App config
var appConfigKeys = []configKey{}

// Init
func configInit() {
	logrus.Info("Initialising the configuration from file")

	// Get config directory from env
	configDir := defaultConfigDir
	val, ok := os.LookupEnv("CONFIG_DIR")
	if ok {
		if len(val) > 0 {
			configDir = val
		}
	}

	// Init config from files
	clientConfigLoaded := make(chan bool)
	appConfigLoaded := make(chan bool)
	go configInitFromFile(clientConfig, "client", configDir, clientConfigKeys, clientConfigLoaded)
	go configInitFromFile(appConfig, "config", configDir, appConfigKeys, appConfigLoaded)
	if <-clientConfigLoaded {
		logrus.Info("Client configuration loaded successfully")
	}
	if <-appConfigLoaded {
		logrus.Info("Application configuration loaded successfully")
	}
}

// Init config from files
func configInitFromFile(v *viper.Viper, configName string, configDir string, configKeys []configKey, loaded chan bool) {
	// File location
	v.SetConfigName(configName)
	v.AddConfigPath(configDir)

	// Defaults
	for _, key := range configKeys {
		v.SetDefault(key.key, key.defval)
	}

	// Load
	var err error
	for err = v.ReadInConfig(); err != nil; {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			logrus.Warn(err)
			time.Sleep(time.Second * 5)
			err = v.ReadInConfig()
		} else {
			logrus.Error(err)
			break
		}
	}

	// Signal the loaded status
	loaded <- (err == nil)

	// Watch
	if err == nil {
		v.OnConfigChange(func(in fsnotify.Event) {
			// Need this throttling trick to avoid double load events
			timeDelay := 100 * time.Millisecond
			if changeThrottleTimer == nil {
				changeThrottleTimer = time.AfterFunc(timeDelay, handleConfigChange)
			} else {
				changeThrottleTimer.Reset(timeDelay)
			}
		})
		v.WatchConfig()
	}
}

var changeThrottleTimer *time.Timer
