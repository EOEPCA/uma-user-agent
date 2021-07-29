package handler

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"
)

func NginxAuthRequestHandler(w http.ResponseWriter, r *http.Request) {

	// Gather expected info from headers/cookies
	origUri := r.Header.Get("X-Original-Uri")
	origMethod := r.Header.Get("X-Original-Method")
	userIdToken := r.Header.Get("X-User-Id")
	// If no user ID token in header, then fall back to cookie
	if len(userIdToken) == 0 {
		c, err := r.Cookie("auth_user_id")
		if err == nil {
			userIdToken = c.Value
		}
	}

	if len(origUri) == 0 || len(origMethod) == 0 || len(userIdToken) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Debug(fmt.Sprintf("Handling request: origUri: %v, origMethod: %v, userIdToken: %v", origUri, origMethod, userIdToken))

	fmt.Fprintln(w, "this is the uma-user-agent")
}
