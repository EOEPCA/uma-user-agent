package handler

import (
	"context"
	"net/http"

	"github.com/sirupsen/logrus"
)

type requestContextKey string

var contextKeyOrigUri = requestContextKey("ORIG_URI")
var contextKeyOrigMethod = requestContextKey("ORIG_METHOD")
var contextKeyUserIdToken = requestContextKey("USER_ID_TOKEN")
var contextKeyRpt = requestContextKey("RPT")
var contextKeyTries = requestContextKey("TRIES")

// UpdateRequestWithDetails puts the supplied request details into the request context,
// and returns a new Request object with the updated context
func UpdateRequestWithDetails(r *http.Request, requestDetails *ClientRequestDetails) *http.Request {
	c := r.Context()
	c = context.WithValue(c, contextKeyOrigUri, requestDetails.OrigUri)
	c = context.WithValue(c, contextKeyOrigMethod, requestDetails.OrigMethod)
	c = context.WithValue(c, contextKeyUserIdToken, requestDetails.UserIdToken)
	c = context.WithValue(c, contextKeyRpt, requestDetails.Rpt)
	c = context.WithValue(c, contextKeyTries, requestDetails.Tries)
	return r.WithContext(c)
}

// GetRequestLogger returns a logger with fields set from the supplied request context
func GetRequestLogger(c context.Context) *logrus.Entry {
	return logrus.StandardLogger().WithFields(logrus.Fields{
		"origUri":    c.Value(contextKeyOrigUri),
		"origMethod": c.Value(contextKeyOrigMethod),
		"attempt":    c.Value(contextKeyTries),
	})
}
