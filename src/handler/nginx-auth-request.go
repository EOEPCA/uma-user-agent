package handler

import (
	"fmt"
	"net/http"

	"github.com/EOEPCA/uma-user-agent/src/config"
	log "github.com/sirupsen/logrus"
)

type requestDetails struct {
	OrigUri     string
	OrigMethod  string
	UserIdToken string
}

func NginxAuthRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Gather expected info from headers/cookies
	details, err := processRequestHeaders(w, r)
	if err != nil {
		log.Error("ERROR processing request headers: ", err)
		return
	}

	// Naive call to the PEP
	//
	// Prepare the request
	pepReq, err := http.NewRequest("GET", config.Config.PepUrl, nil)
	if err != nil {
		log.Error("Error establishing request for PEP: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	pepReq.Header.Set("X-Orig-Uri", details.OrigUri)
	pepReq.Header.Set("X-Orig-Method", details.OrigMethod)
	pepReq.Header.Set("X-User-Id", details.UserIdToken)
	//
	// Send the request
	client := http.Client{}
	pepResp, err := client.Do(pepReq)
	if err != nil {
		log.Error("Error requesting auth from PEP: ", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//
	// Handle the response
	switch code := pepResp.StatusCode; {
	case code >= 200 && code <= 299:
		// zzz
	case code == 401:
		// zzz
	case code == 403:
		// zzz
	default:
		// zzz
	}

	fmt.Fprintln(w, "this is the uma-user-agent")
}

func processRequestHeaders(w http.ResponseWriter, r *http.Request) (details requestDetails, err error) {
	details = requestDetails{}
	err = nil

	// Gather expected info from headers/cookies
	details.OrigUri = r.Header.Get("X-Original-Uri")
	details.OrigMethod = r.Header.Get("X-Original-Method")
	details.UserIdToken = r.Header.Get("X-User-Id")
	// If no user ID token in header, then fall back to cookie
	if len(details.UserIdToken) == 0 {
		c, err := r.Cookie(config.Config.UserIdCookieName)
		if err == nil {
			details.UserIdToken = c.Value
		}
	}

	// Check details are complete
	if len(details.OrigUri) == 0 || len(details.OrigMethod) == 0 || len(details.UserIdToken) == 0 {
		err = fmt.Errorf("mandatory header values missing")
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintln(w, "ERROR: Expecting non-zero values for the following data...")
		fmt.Fprintln(w, "  Original URI:    ", details.OrigUri, "\n    [header X-Orig-Uri]")
		fmt.Fprintln(w, "  Original Method: ", details.OrigMethod, "\n    [header X-Orig-Method]")
		fmt.Fprintln(w, "  User ID Token:   ", details.UserIdToken, "\n    [header X-User-Id or cookie '"+config.Config.UserIdCookieName+"']")
		return
	}
	log.Debug(fmt.Sprintf("Handling request: origUri: %v, origMethod: %v, userIdToken: %v", details.OrigUri, details.OrigMethod, details.UserIdToken))

	return
}
