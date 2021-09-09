package handler

import (
	"fmt"
	"net/http"

	"github.com/EOEPCA/uma-user-agent/pkg/config"
	"github.com/EOEPCA/uma-user-agent/pkg/uma"
	"github.com/sirupsen/logrus"
)

// HTTP Header Names
const headerNameXOriginalUri = "X-Original-Uri"
const headerNameXOriginalMethod = "X-Original-Method"
const headerNameXUserId = "X-User-Id"

type ClientRequestDetails struct {
	OrigUri     string
	OrigMethod  string
	UserIdToken string
	Rpt         string
}

type pepResponseHandlerFunc func(ClientRequestDetails, *http.Response, http.ResponseWriter, *http.Request)

func NginxAuthRequestHandler(w http.ResponseWriter, r *http.Request) {
	// Gather expected info from headers/cookies
	clientRequestDetails, err := processRequestHeaders(w, r)
	if err != nil {
		logrus.Error("ERROR processing request headers: ", err)
		return
	}

	// Naive call to the PEP
	pepResponse, err := pepAuthRequest(clientRequestDetails)
	if err != nil {
		msg := "ERROR making naive call to the pep auth_request endpoint"
		logrus.Error(msg, ": ", err)
		WriteHeaderUnauthorized(w)
		fmt.Fprint(w, msg)
		return
	}

	// Handle the response
	handlePepResponse(clientRequestDetails, pepResponse, handlePepNaiveUnauthorized, w, r)
}

func handlePepResponse(clientRequestDetails ClientRequestDetails, pepResponse *http.Response, unauthResponseHandler pepResponseHandlerFunc, w http.ResponseWriter, r *http.Request) {
	switch code := pepResponse.StatusCode; {
	case code >= 200 && code <= 299:
		// AUTHORIZED
		msg := fmt.Sprintf("PEP authorized the request with code: %v", code)
		logrus.Info(msg)
		w.Header().Set(headerNameXUserId, clientRequestDetails.UserIdToken)
		w.WriteHeader(code)
		fmt.Fprint(w, msg)
	case code == 401:
		// UNAUTHORIZED
		if unauthResponseHandler != nil {
			unauthResponseHandler(clientRequestDetails, pepResponse, w, r)
		} else {
			msg := "PEP responded UNAUTHORIZED"
			logrus.Info(msg)
			WriteHeaderUnauthorized(w)
			fmt.Fprint(w, msg)
		}
	case code == 403:
		// FORBIDDEN
		msg := "PEP responded FORBIDDEN"
		logrus.Info(msg)
		w.WriteHeader(code)
		fmt.Fprint(w, msg)
	default:
		// UNEXPECTED
		msg := fmt.Sprintf("Unexpected return code from PEP auth_request endpoint: %v", code)
		logrus.Error(msg)
		WriteHeaderUnauthorized(w)
		fmt.Fprint(w, msg)
	}
}

func processRequestHeaders(w http.ResponseWriter, r *http.Request) (details ClientRequestDetails, err error) {
	details = ClientRequestDetails{}
	err = nil

	// Gather expected info from headers/cookies
	details.OrigUri = r.Header.Get(headerNameXOriginalUri)
	details.OrigMethod = r.Header.Get(headerNameXOriginalMethod)
	details.UserIdToken = r.Header.Get(headerNameXUserId)
	// If no user ID token in header, then fall back to cookie
	if len(details.UserIdToken) == 0 {
		c, err := r.Cookie(config.GetUserIdCookieName())
		if err == nil {
			details.UserIdToken = c.Value
		}
	}

	// Check details are complete
	if len(details.OrigUri) == 0 || len(details.OrigMethod) == 0 || len(details.UserIdToken) == 0 {
		err = fmt.Errorf("mandatory header values missing")
		WriteHeaderUnauthorized(w)
		fmt.Fprintln(w, "ERROR: Expecting non-zero values for the following data...")
		fmt.Fprintf(w, "  Original URI:    %v\n    [header %v]\n", details.OrigUri, headerNameXOriginalUri)
		fmt.Fprintf(w, "  Original Method: %v\n    [header %v]\n", details.OrigMethod, headerNameXOriginalMethod)
		fmt.Fprintf(w, "  User ID Token:   %v\n    [header %v or cookie '%v']\n", details.UserIdToken, headerNameXUserId, config.GetUserIdCookieName())
		return
	}
	logrus.Debug(fmt.Sprintf("Handling request: origUri: %v, origMethod: %v, userIdToken: %v", details.OrigUri, details.OrigMethod, details.UserIdToken))

	return
}

// pepAuthRequest calls the PEP `auth_request` endpoint
func pepAuthRequest(details ClientRequestDetails) (response *http.Response, err error) {
	response = nil
	err = nil

	// Prepare the request
	pepReq, err := http.NewRequest("GET", config.GetPepUrl(), nil)
	if err != nil {
		err = fmt.Errorf("error establishing request for PEP: %w", err)
		logrus.Error(err)
		return
	}
	pepReq.Header.Set(headerNameXOriginalUri, details.OrigUri)
	pepReq.Header.Set(headerNameXOriginalMethod, details.OrigMethod)
	pepReq.Header.Set(headerNameXUserId, details.UserIdToken)
	if len(details.Rpt) > 0 {
		pepReq.Header.Set("Authorization", fmt.Sprintf("Bearer %v", details.Rpt))
	}

	// Send the request
	response, err = uma.HttpClient.Do(pepReq)
	if err != nil {
		response = nil
		err = fmt.Errorf("error requesting auth from PEP: %w", err)
		logrus.Error(err)
	}

	return response, err
}

func handlePepNaiveUnauthorized(clientRequestDetails ClientRequestDetails, pepUnauthResponse *http.Response, w http.ResponseWriter, r *http.Request) {
	// Check that this is a 401 response
	if pepUnauthResponse.StatusCode != http.StatusUnauthorized {
		msg := "not an Unauthorized response"
		logrus.Error(msg)
		WriteHeaderUnauthorized(w)
		fmt.Fprint(w, msg)
		return
	}

	// Get the expected Www-Authenticate header
	wwwAuthHeader := pepUnauthResponse.Header.Get("Www-Authenticate")
	if len(wwwAuthHeader) == 0 {
		msg := "no Www-Authenticate header in PEP response"
		logrus.Error(msg)
		WriteHeaderUnauthorized(w)
		fmt.Fprint(w, msg)
		return
	}

	// Parse the Www-Authenticate header
	authServerUrl, ticket, err := uma.UnpackWwwAuthenticateHeader(wwwAuthHeader)
	if err != nil {
		msg := "could not parse the Www-Authenticate header"
		logrus.Error(msg, ": ", err)
		WriteHeaderUnauthorized(w)
		fmt.Fprint(w, msg)
		return
	}
	// Store the Authorization Server
	authServer, _ := uma.AuthorizationServers.LoadOrStore(authServerUrl, *uma.NewAuthorizationServer(authServerUrl))
	if len(authServer.GetUrl()) == 0 {
		msg := "error getting the Authorization Server details"
		logrus.Error(msg)
		WriteHeaderUnauthorized(w)
		fmt.Fprint(w, msg)
		return
	}

	// Exchange the ticket for an RPT at the Authorization Server
	umaClient := &uma.UmaClient{Id: config.GetClientId(), Secret: config.GetClientSecret()}
	var forbidden bool
	clientRequestDetails.Rpt, forbidden, err = umaClient.ExchangeTicketForRpt(authServer, clientRequestDetails.UserIdToken, ticket)
	if err != nil {
		var msg string
		if forbidden {
			msg = "access request FORBIDDEN by Authorization Server"
			logrus.Warn(msg, ": ", err)
			w.WriteHeader(http.StatusForbidden)
		} else {
			msg = "error getting RPT from Authorization Server"
			logrus.Error(msg, ": ", err)
			WriteHeaderUnauthorized(w)
		}
		fmt.Fprint(w, msg)
		return
	}

	// Call the PEP with the RPT
	pepResponse, err := pepAuthRequest(clientRequestDetails)
	if err != nil {
		msg := "ERROR making call (with RPT) to the pep auth_request endpoint"
		logrus.Error(msg, ": ", err)
		WriteHeaderUnauthorized(w)
		fmt.Fprint(w, msg)
		return
	}

	// Handle the response
	handlePepResponse(clientRequestDetails, pepResponse, nil, w, r)
}

// WriteHeaderUnauthorized writes the header response to indicate unauthorized
func WriteHeaderUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Www-Authenticate", config.GetUnauthorizedResponse())
	w.WriteHeader(http.StatusUnauthorized)
}
