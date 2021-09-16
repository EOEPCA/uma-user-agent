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

// ClientRequestDetails represents the details of the 'incoming' request made by the client
type ClientRequestDetails struct {
	OrigUri     string
	OrigMethod  string
	UserIdToken string
	Rpt         string
}

// wrappedResponseWriter provides access to the StatusCode that is written to the http header,
// for the purposes of request logging
type wrappedResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// Override (wrap) the WriteHeader function to take a note of the status code - exposed via
// a public (exported) value.
func (rl *wrappedResponseWriter) WriteHeader(statusCode int) {
	rl.ResponseWriter.WriteHeader(statusCode)
	rl.StatusCode = statusCode
}

// Helper function to log the request completion including the status code.
func (rl *wrappedResponseWriter) LogRequestCompletion(requestLogger *logrus.Entry) {
	requestLogger.WithField("statusCode", rl.StatusCode).Info("Request complete")
}

// pepResponseHandlerFunc defines the function prototype for functions able to
// handle the response to an `auth_request` made to the PEP
type pepResponseHandlerFunc func(ClientRequestDetails, *http.Response, http.ResponseWriter, *http.Request)

// NginxAuthRequestHandler is the entrypoint handler for the nginx `auth_request` implementation
func NginxAuthRequestHandler(rw http.ResponseWriter, r *http.Request) {
	w := &wrappedResponseWriter{rw, http.StatusOK}

	// Gather expected info from headers/cookies
	r, clientRequestDetails, err := processRequestHeaders(w, r)
	requestLogger := GetRequestLogger(r.Context())
	defer w.LogRequestCompletion(requestLogger)
	if err != nil {
		requestLogger.Error("ERROR processing request headers: ", err)
		return
	}
	requestLogger.Debug("START handling new request")

	// Naive call to the PEP
	requestLogger.Debug("Calling PEP `auth_request` initial (naive) attempt")
	pepResponse, err := pepAuthRequest(clientRequestDetails)
	if err != nil {
		msg := "ERROR making naive call to the pep auth_request endpoint"
		requestLogger.Error(fmt.Errorf("%s: %w", msg, err))
		WriteHeaderUnauthorized(w)
		fmt.Fprint(w, msg)
		return
	}

	// Handle the response
	handlePepResponse(clientRequestDetails, pepResponse, handlePepNaiveUnauthorized, w, r)
}

// handlePepResponse is a helper function to handle the response from the PEP's `auth_request` endpoint
func handlePepResponse(clientRequestDetails ClientRequestDetails, pepResponse *http.Response, unauthResponseHandler pepResponseHandlerFunc, w http.ResponseWriter, r *http.Request) {
	requestLogger := GetRequestLogger(r.Context())
	switch code := pepResponse.StatusCode; {
	case code >= 200 && code <= 299:
		// AUTHORIZED
		msg := fmt.Sprintf("PEP authorized the request with code: %v", code)
		requestLogger.Debug(msg)
		w.Header().Set(headerNameXUserId, clientRequestDetails.UserIdToken)
		w.WriteHeader(code)
		fmt.Fprint(w, msg)
	case code == 401:
		// UNAUTHORIZED
		msg := "PEP responded UNAUTHORIZED"
		requestLogger.Debug(msg)
		if unauthResponseHandler != nil {
			unauthResponseHandler(clientRequestDetails, pepResponse, w, r)
		} else {
			requestLogger.Debugf("RPT was not accepted: %s", clientRequestDetails.Rpt)
			WriteHeaderUnauthorized(w)
			fmt.Fprint(w, msg)
		}
	case code == 403:
		// FORBIDDEN
		msg := "PEP responded FORBIDDEN"
		requestLogger.Debug(msg)
		w.WriteHeader(code)
		fmt.Fprint(w, msg)
	default:
		// UNEXPECTED
		msg := fmt.Sprintf("Unexpected return code from PEP auth_request endpoint: %v", code)
		requestLogger.Error(msg)
		WriteHeaderUnauthorized(w)
		fmt.Fprint(w, msg)
	}
}

// processRequestHeaders is a helper function to extract the expected information from the
// http headers of the received `auth_request`
func processRequestHeaders(w http.ResponseWriter, r *http.Request) (reqUpdated *http.Request, details ClientRequestDetails, err error) {
	reqUpdated = r
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

	// Update the request context with the supplied headers
	r = UpdateRequestWithDetails(r, details)
	reqUpdated = r

	// Get the request logger
	requestLogger := GetRequestLogger(reqUpdated.Context())

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
	requestLogger.Trace(fmt.Sprintf("Handling request: origUri: %v, origMethod: %v, userIdToken: %v", details.OrigUri, details.OrigMethod, details.UserIdToken))

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
	}

	return response, err
}

// handlePepNaiveUnauthorized provides the behaviour that is triggered by a 401 (Unauthorized)
// response to a naive (no RPT) request to the PEP `auth_request` endpoint
func handlePepNaiveUnauthorized(clientRequestDetails ClientRequestDetails, pepUnauthResponse *http.Response, w http.ResponseWriter, r *http.Request) {
	requestLogger := GetRequestLogger(r.Context())
	// Check that this is a 401 response
	if pepUnauthResponse.StatusCode != http.StatusUnauthorized {
		msg := "not an Unauthorized response"
		requestLogger.Error(msg)
		WriteHeaderUnauthorized(w)
		fmt.Fprint(w, msg)
		return
	}

	// Get the expected Www-Authenticate header
	wwwAuthHeader := pepUnauthResponse.Header.Get("Www-Authenticate")
	if len(wwwAuthHeader) == 0 {
		msg := "no Www-Authenticate header in PEP response"
		requestLogger.Error(msg)
		WriteHeaderUnauthorized(w)
		fmt.Fprint(w, msg)
		return
	}

	// Parse the Www-Authenticate header
	authServerUrl, ticket, err := uma.UnpackWwwAuthenticateHeader(wwwAuthHeader)
	if err != nil {
		msg := "could not parse the Www-Authenticate header"
		requestLogger.Error(fmt.Errorf("%s: %w", msg, err))
		WriteHeaderUnauthorized(w)
		fmt.Fprint(w, msg)
		return
	}
	// Store the Authorization Server
	authServer, _ := uma.AuthorizationServers.LoadOrStore(requestLogger, authServerUrl, *uma.NewAuthorizationServer(authServerUrl))
	if len(authServer.GetUrl()) == 0 {
		msg := "error getting the Authorization Server details"
		requestLogger.Error(msg)
		WriteHeaderUnauthorized(w)
		fmt.Fprint(w, msg)
		return
	}

	// Exchange the ticket for an RPT at the Authorization Server
	umaClient := &uma.UmaClient{Id: config.GetClientId(), Secret: config.GetClientSecret()}
	var forbidden bool
	clientRequestDetails.Rpt, forbidden, err = umaClient.ExchangeTicketForRpt(requestLogger, authServer, clientRequestDetails.UserIdToken, ticket)
	if err != nil {
		var msg string
		if forbidden {
			msg = "access request FORBIDDEN by Authorization Server"
			requestLogger.Warn(fmt.Errorf("%s: %w", msg, err))
			w.WriteHeader(http.StatusForbidden)
		} else {
			msg = "error getting RPT from Authorization Server"
			requestLogger.Error(fmt.Errorf("%s: %w", msg, err))
			WriteHeaderUnauthorized(w)
		}
		fmt.Fprint(w, msg)
		return
	}

	// Check the RPT looks OK
	if len(clientRequestDetails.Rpt) == 0 {
		msg := "the RPT obtained is blank"
		requestLogger.Error(msg)
		WriteHeaderUnauthorized(w)
		fmt.Fprint(w, msg)
		return
	}

	// Call the PEP with the RPT
	requestLogger.Debug("Calling PEP `auth_request` with RPT")
	pepResponse, err := pepAuthRequest(clientRequestDetails)
	if err != nil {
		msg := "ERROR making call (with RPT) to the pep auth_request endpoint"
		requestLogger.Error(fmt.Errorf("%s: %w", msg, err))
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
