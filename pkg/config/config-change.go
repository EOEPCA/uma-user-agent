package config

import (
	"sync"

	"github.com/sirupsen/logrus"
)

type ConfigChangeHandler func()

var configChangeHandlers = struct {
	handlers []ConfigChangeHandler
	mutex    sync.RWMutex
}{}

func AddConfigChangeHandler(h ConfigChangeHandler) {
	configChangeHandlers.mutex.Lock()
	defer configChangeHandlers.mutex.Unlock()
	configChangeHandlers.handlers = append(configChangeHandlers.handlers, h)
}

func TriggerConfigChangeHandlers() {
	configChangeHandlers.mutex.RLock()
	defer configChangeHandlers.mutex.RUnlock()
	for _, h := range configChangeHandlers.handlers {
		h()
	}
}

func handleConfigChange() {
	logrus.Warn("Config has changed")
	TriggerConfigChangeHandlers()
}
