package uma

import (
	"net/http"
	"time"

	"github.com/EOEPCA/uma-user-agent/src/config"
)

var HttpClient = &http.Client{
	Timeout: time.Second * config.Config.HttpTimeout,
}
