package handler

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
)

type requestContextKey string

var keyOrigUri = requestContextKey("ORIG_URI")
var keyOrigMethod = requestContextKey("ORIG_METHOD")
var keyUserIdToken = requestContextKey("USER_ID_TOKEN")

// UpdateRequestWithDetails puts the supplied request details into the request context,
// and returns a new Request object with the updated context
func UpdateRequestWithDetails(r *http.Request, requestDetails ClientRequestDetails) *http.Request {
	c := r.Context()
	c = context.WithValue(c, keyOrigUri, requestDetails.OrigUri)
	c = context.WithValue(c, keyOrigMethod, requestDetails.OrigMethod)
	c = context.WithValue(c, keyUserIdToken, requestDetails.UserIdToken)
	return r.WithContext(c)
}

// GetRequestLogger returns a logger with fields set from the supplied request context
func GetRequestLogger(c context.Context) *logrus.Entry {
	return logrus.StandardLogger().WithFields(logrus.Fields{
		"origUri":    c.Value(keyOrigUri),
		"origMethod": c.Value(keyOrigMethod),
	})
}
