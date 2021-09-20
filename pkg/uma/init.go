package uma

import (
	"github.com/EOEPCA/uma-user-agent/pkg/config"
)

func init() {
	config.AddConfigChangeHandler(configChangeHandler)
}
