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
const headerNameXAuthRpt = "X-Auth-Rpt"

// ClientRequestDetails represents the details of the 'incoming' request made by the client
type ClientRequestDetails struct {
	OrigUri     string
	OrigMethod  string
	UserIdToken string
	Rpt         string
	Tries       int
}

// GetRequestLogger returns a logger with fields set from the supplied client request details
func GetRequestLogger(clientRequestDetails *ClientRequestDetails) *logrus.Entry {
	return logrus.StandardLogger().WithFields(logrus.Fields{
		"origUri":    clientRequestDetails.OrigUri,
		"origMethod": clientRequestDetails.OrigMethod,
		"attempt":    clientRequestDetails.Tries,
	})
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
type pepResponseHandlerFunc func(*ClientRequestDetails, *http.Response, http.ResponseWriter, *http.Request)

// NginxAuthRequestHandler is the entrypoint handler for the nginx `auth_request` implementation
func NginxAuthRequestHandler(rw http.ResponseWriter, r *http.Request) {
	var clientRequestDetails *ClientRequestDetails
	// Ensure that request status is logged at completion
	w := &wrappedResponseWriter{rw, http.StatusOK}
	defer func() {
		w.LogRequestCompletion(GetRequestLogger(clientRequestDetails))
	}()

	// Gather expected info from headers/cookies
	r, clientRequestDetails, err := processRequestHeaders(w, r)
	requestLogger := GetRequestLogger(clientRequestDetails)
	if err != nil {
		requestLogger.Error("ERROR processing request headers: ", err)
		return
	}

	// If we are in 'OPEN' mode then the request is simply allowed
	if nginxAuthRequestHandlerOpen(w, r) {
		return
	}

	// Defer the Authorization decision to the PEP
	requestLogger.Debug("START handling new request")
	deferAuthorizationToPep(clientRequestDetails, w, r)
}

func deferAuthorizationToPep(clientRequestDetails *ClientRequestDetails, w http.ResponseWriter, r *http.Request) {
	// Increment the 'try' counter
	clientRequestDetails.Tries += 1
	requestLogger := GetRequestLogger(clientRequestDetails)
	if clientRequestDetails.Tries > 1 {
		requestLogger.Warningf("Authorization retry attempt #%d", clientRequestDetails.Tries-1)
	} else {
		requestLogger.Debug("First Authorization attempt")
	}

	// Naive call to the PEP
	requestLogger.Debug("Calling PEP `auth_request` initial (naive) attempt")
	pepResponse, err := pepAuthRequest(clientRequestDetails, requestLogger)
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

// nginxAuthRequestHandlerOpen provides an nginx `auth_request` handler for OPEN access
func nginxAuthRequestHandlerOpen(w http.ResponseWriter, r *http.Request) (requestHandled bool) {
	requestHandled = config.IsOpenAccess()
	if requestHandled {
		// Pass on the User ID Token if provided in the request
		userIdToken := r.Header.Get(headerNameXUserId)
		// If no user ID token in header, then fall back to cookie
		if len(userIdToken) == 0 {
			c, err := r.Cookie(config.GetUserIdCookieName())
			if err == nil {
				userIdToken = c.Value
			}
		}
		w.Header().Set(headerNameXUserId, userIdToken)

		fmt.Fprintln(w, "Allowing OPEN access")
	}
	return
}

// handlePepResponse is a helper function to handle the response from the PEP's `auth_request` endpoint
func handlePepResponse(clientRequestDetails *ClientRequestDetails, pepResponse *http.Response, unauthResponseHandler pepResponseHandlerFunc, w http.ResponseWriter, r *http.Request) {
	requestLogger := GetRequestLogger(clientRequestDetails)
	switch code := pepResponse.StatusCode; {
	case code >= 200 && code <= 299:
		// AUTHORIZED
		msg := fmt.Sprintf("PEP authorized the request with code: %v", code)
		requestLogger.Debug(msg)
		w.Header().Set(headerNameXUserId, clientRequestDetails.UserIdToken)
		setRptCookieInResponse(clientRequestDetails.Rpt, w)
		w.WriteHeader(code)
		fmt.Fprint(w, msg)
	case code == 401:
		// UNAUTHORIZED
		msg := "PEP responded UNAUTHORIZED"
		requestLogger.Debug(msg)
		// Use specific handler if it's been provided
		if unauthResponseHandler != nil {
			unauthResponseHandler(clientRequestDetails, pepResponse, w, r)
		} else {
			// If) we have remaining retry attempts, then go back around the loop
			// Else) retries are exhausted, so return unauthorized
			if (clientRequestDetails.Tries - 1) < config.GetRetriesAuthorizationAttempt() {
				deferAuthorizationToPep(clientRequestDetails, w, r)
			} else {
				requestLogger.Debugf("RPT was not accepted: %s", clientRequestDetails.Rpt)
				WriteHeaderUnauthorized(w)
				fmt.Fprint(w, msg)
			}
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
func processRequestHeaders(w http.ResponseWriter, r *http.Request) (reqUpdated *http.Request, details *ClientRequestDetails, err error) {
	reqUpdated = r
	details = &ClientRequestDetails{}
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
	// RPT
	{
		c, err := r.Cookie(config.GetAuthRptCookieName())
		if err == nil {
			details.Rpt = c.Value
		}
	}

	// Get the request logger
	requestLogger := GetRequestLogger(details)

	// Some verbose logging
	requestLogger.Tracef("%s: %s", headerNameXOriginalMethod, details.OrigMethod)
	requestLogger.Tracef("%s: %s", headerNameXOriginalUri, details.OrigUri)
	requestLogger.Tracef("%s: %s", headerNameXUserId, details.UserIdToken)
	requestLogger.Tracef("%s: %s", headerNameXAuthRpt, details.Rpt)

	// Check details are complete
	if len(details.OrigUri) == 0 || len(details.OrigMethod) == 0 {
		err = fmt.Errorf("mandatory header values missing")
		WriteHeaderUnauthorized(w)
		fmt.Fprintln(w, "ERROR: Expecting non-zero values for the following data...")
		fmt.Fprintf(w, "  Original URI:    %v\n    [header %v]\n", details.OrigUri, headerNameXOriginalUri)
		fmt.Fprintf(w, "  Original Method: %v\n    [header %v]\n", details.OrigMethod, headerNameXOriginalMethod)
		return
	}

	return
}

// pepAuthRequest calls the PEP `auth_request` endpoint
func pepAuthRequest(details *ClientRequestDetails, requestLogger *logrus.Entry) (response *http.Response, err error) {
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
	if len(details.UserIdToken) > 0 {
		pepReq.Header.Set(headerNameXUserId, details.UserIdToken)
	}
	if len(details.Rpt) > 0 {
		pepReq.Header.Set("Authorization", fmt.Sprintf("Bearer %v", details.Rpt))
	}

	// Send the request
	response, err = uma.MakeResilentRequest(pepReq, requestLogger, "pepAuthRequest")
	if err != nil {
		response = nil
		err = fmt.Errorf("error requesting auth from PEP: %w", err)
	}

	return response, err
}

// handlePepNaiveUnauthorized provides the behaviour that is triggered by a 401 (Unauthorized)
// response to a naive (no RPT) request to the PEP `auth_request` endpoint
func handlePepNaiveUnauthorized(clientRequestDetails *ClientRequestDetails, pepUnauthResponse *http.Response, w http.ResponseWriter, r *http.Request) {
	requestLogger := GetRequestLogger(clientRequestDetails)
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
	requestLogger.Tracef("Obtained RPT: %s", clientRequestDetails.Rpt)

	// Refresh the request logger with updated client details
	requestLogger = GetRequestLogger(clientRequestDetails)

	// Call the PEP with the RPT
	requestLogger.Debug("Calling PEP `auth_request` with RPT")
	pepResponse, err := pepAuthRequest(clientRequestDetails, requestLogger)
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

// setRptCookieInResponse uses an http header to provide the `Set-Cookie` string.
func setRptCookieInResponse(rpt string, w http.ResponseWriter) {
	w.Header().Set(headerNameXAuthRpt,
		fmt.Sprintf("%s=%s; Path=/; Secure; HttpOnly; Max-Age=%d",
			config.GetAuthRptCookieName(),
			rpt, config.GetAuthRptCookieMaxAge()))
}
