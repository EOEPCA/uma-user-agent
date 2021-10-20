package uma

import (
	"net/http"
	"net/url"
	"time"

	"github.com/EOEPCA/uma-user-agent/pkg/config"
	"github.com/sirupsen/logrus"
)

var HttpClient = &http.Client{}

func configChangeHandler() {
	HttpClient.Timeout = time.Second * config.GetHttpTimeout()
	logrus.Info("Initialised Http Client with timeout: ", HttpClient.Timeout)
}

func MakeResilentRequest(req *http.Request, requestLogger *logrus.Entry, reason string) (response *http.Response, err error) {
	for attempts := 0; attempts <= 1; attempts++ {
		response, err = HttpClient.Do(req)
		// There are a couple of conditions that will cause us to retry:
		// * the response code is 500+
		// * there is an error due to http timeout
		// In these cases we will 'continue' to repeat the loop
		if err == nil {
			// Bad status code
			if response.StatusCode >= 500 {
				requestLogger.Warnf("[%s] retrying request due to bad response code: %d", reason, response.StatusCode)
				continue
			}
		} else {
			response = nil
			// Error is a timeout
			if urlErr, ok := err.(*url.Error); ok && urlErr.Timeout() {
				requestLogger.Warnf("[%s] retrying request due to timeout", reason)
				continue
			}
		}
		break
	}
	return
}
