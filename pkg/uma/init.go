package uma

import (
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.Info("Initialising Http Client with timeout: ", HttpClient.Timeout)
}
