package config

import (
	"os"

	log "github.com/sirupsen/logrus"
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
var keyLoggingLevel = configKey{"logging.level", log.InfoLevel}
var keyHttpTimeout = configKey{"network.httpTimeout", 10}
var keyListenPort = configKey{"network.listenPort", 80}
var keyPepUrl = configKey{"pep.url", "http://pep"}
var keyUserIdCookieName = configKey{"userIdCookieName", "auth_user_id"}
var keyUnauthorizedResponse = configKey{"unauthorizedResponse", "Please login to access the resource"}

// Client config
var clientConfigKeys = []configKey{keyClientId, keyClientSecret}

// App config
var appConfigKeys = []configKey{}

// Init
func viperInit() {
	log.Info("Initialising the viper configuration")

	// Get config directory from env
	configDir := defaultConfigDir
	val, ok := os.LookupEnv("CONFIG_DIR")
	if ok {
		if len(val) > 0 {
			configDir = val
		}
	}

	// Init config from files
	configInit(clientConfig, "client", configDir, clientConfigKeys)
	configInit(appConfig, "config", configDir, appConfigKeys)
}

// Init config from files
func configInit(v *viper.Viper, configName string, configDir string, configKeys []configKey) {
	// File location
	v.SetConfigName(configName)
	v.AddConfigPath(configDir)

	// Defaults
	for _, key := range configKeys {
		v.SetDefault(key.key, key.defval)
	}

	// Load
	v.ReadInConfig()

	// Watch
	v.WatchConfig()
}