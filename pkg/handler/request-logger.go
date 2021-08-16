package handler

import (
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
)

// RequestLogger provides a request logging middleware.
// To avoid the liveness/readiness probes flooding the log, the URL path '/status/'
// is suppressed from logging, unless the log level is set to 'trace'.
func RequestLogger(h http.Handler) (handler http.Handler) {
	handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Suppress logging the path '/status/' (unless we have verbose logging enabled)
		if log.GetLevel() == log.TraceLevel || !strings.HasPrefix(r.URL.Path, "/status/") {
			handlers.CombinedLoggingHandler(os.Stdout, h).ServeHTTP(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
	return
}
