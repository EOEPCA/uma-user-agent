package uma

import (
	"github.com/EOEPCA/uma-user-agent/pkg/config"
)

func init() {
	initHttpClient()
	config.AddConfigChangeHandler(configChangeHandler)
}
