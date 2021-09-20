package uma

import (
	"net/http"
	"time"

	"github.com/EOEPCA/uma-user-agent/pkg/config"
	"github.com/sirupsen/logrus"
)

var HttpClient = &http.Client{}

func configChangeHandler() {
	HttpClient.Timeout = time.Second * config.GetHttpTimeout()
	logrus.Info("Initialised Http Client with timeout: ", HttpClient.Timeout)
}
