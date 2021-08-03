package uma

import (
	log "github.com/sirupsen/logrus"
)

func init() {
	log.Info("Initialising Http Client with timeout: ", HttpClient.Timeout)
}
