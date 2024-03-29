package uma

import (
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	"github.com/EOEPCA/uma-user-agent/pkg/config"
	"github.com/sirupsen/logrus"
)

var HttpClient *http.Client

func initHttpClient() {
	if config.AllowInsecureTlsSkipVerify() {
		transport := http.DefaultTransport.(*http.Transport)
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		HttpClient = &http.Client{Transport: transport}
	} else {
		HttpClient = &http.Client{}
	}
}

func configChangeHandler() {
	initHttpClient()
	HttpClient.Timeout = time.Second * config.GetHttpTimeout()
	logrus.Infof("Initialised Http Client: timeout=%v, insecure-tls=%v", HttpClient.Timeout, config.AllowInsecureTlsSkipVerify())
}

// MakeResilentRequest makes the provided http request with additional logic to perform
// a configurable number of retries.
// Conditions that will cause us to retry...
// * the response code is 5xx
// * there is an error due to http timeout
func MakeResilentRequest(req *http.Request, requestLogger *logrus.Entry, reason string) (response *http.Response, err error) {
	for attempts := 0; attempts <= config.GetRetriesHttpRequest(); attempts++ {
		response, err = HttpClient.Do(req)
		// Check if conditions are met for a retry and, of so, 'continue' to repeat the loop and so make another attempt
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
