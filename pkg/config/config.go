package config

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

func IsReady() (isReady bool) {
	isReady = true &&
		len(GetClientId()) > 0 &&
		len(GetClientSecret()) > 0
	return
}

func GetClientId() string {
	return clientConfig.GetString(keyClientId.key)
}

func GetClientSecret() string {
	return clientConfig.GetString(keyClientSecret.key)
}

func GetHttpTimeout() time.Duration {
	return appConfig.GetDuration(keyHttpTimeout.key)
}

func GetLogLevel() logrus.Level {
	// default
	var logLevel logrus.Level
	var ok bool
	if logLevel, ok = keyLoggingLevel.defval.(logrus.Level); !ok {
		logLevel = logrus.InfoLevel
	}

	// read from config
	val := appConfig.GetString(keyLoggingLevel.key)
	l, err := logrus.ParseLevel(val)
	if err != nil {
		logrus.Warning(fmt.Sprintf("Bad log level '%v' specified, using default '%v'", val, logLevel.String()))
	} else {
		logLevel = l
	}

	logrus.Info("Using log level: ", logLevel.String())
	return logLevel
}

func GetPepUrl() string {
	return appConfig.GetString(keyPepUrl.key)
}

func GetPort() int {
	return appConfig.GetInt(keyListenPort.key)
}

func GetUserIdCookieName() string {
	return appConfig.GetString(keyUserIdCookieName.key)
}

func GetAuthRptCookieName() string {
	return appConfig.GetString(keyAuthRptCookieName.key)
}

func GetAuthRptCookieMaxAge() int {
	return appConfig.GetInt(keyAuthRptCookieMaxAge.key)
}

func GetUnauthorizedResponse() string {
	return appConfig.GetString(keyUnauthorizedResponse.key)
}

func GetRetriesAuthorizationAttempt() int {
	return appConfig.GetInt(keyRetriesAuthorizationAttempt.key)
}

func GetRetriesHttpRequest() int {
	return appConfig.GetInt(keyRetriesHttpRequest.key)
}

func IsOpenAccess() bool {
	return appConfig.GetBool(keyOpenAccess.key)
}

func AllowInsecureTlsSkipVerify() bool {
	return appConfig.GetBool(keyInsecureTlsSkipVerify.key)
}
