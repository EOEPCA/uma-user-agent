package uma

import (
	"net/http"
	"time"

	"github.com/EOEPCA/uma-user-agent/pkg/config"
)

var HttpClient = &http.Client{
	Timeout: time.Second * config.GetHttpTimeout(),
}
