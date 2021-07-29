package handler

import (
	"fmt"
	"net/http"

	"github.com/EOEPCA/uma-user-agent/src/config"
	log "github.com/sirupsen/logrus"
)

func NginxAuthRequestHandler(w http.ResponseWriter, r *http.Request) {

	// Gather expected info from headers/cookies
	origUri := r.Header.Get("X-Original-Uri")
	origMethod := r.Header.Get("X-Original-Method")
	userIdToken := r.Header.Get("X-User-Id")
	// If no user ID token in header, then fall back to cookie
	if len(userIdToken) == 0 {
		c, err := r.Cookie(config.Config.UserIdCookieName)
		if err == nil {
			userIdToken = c.Value
		}
	}

	if len(origUri) == 0 || len(origMethod) == 0 || len(userIdToken) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "ERROR: Expecting non-zero values for the following data...")
		fmt.Fprintln(w, "  Original URI:    ", origUri, "\n    [header X-Orig-Uri]")
		fmt.Fprintln(w, "  Original Method: ", origMethod, "\n    [header X-Orig-Method]")
		fmt.Fprintln(w, "  User ID Token:   ", userIdToken, "\n    [header X-User-Id or cookie '"+config.Config.UserIdCookieName+"']")
		return
	}

	log.Debug(fmt.Sprintf("Handling request: origUri: %v, origMethod: %v, userIdToken: %v", origUri, origMethod, userIdToken))

	fmt.Fprintln(w, "this is the uma-user-agent")
}
