package uma

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

//------------------------------------------------------------------------------

// UnpackWwwAuthenticateHeader gets the 'Auth Server URL' and Ticket from the
// supplied string which is interpreted as a Www-Authentication header from a
// UMA flow 401 response
func UnpackWwwAuthenticateHeader(wwwAuthenticate string) (authServerUrl string, ticket string, err error) {
	authServerUrl = ""
	ticket = ""
	err = nil

	// Split the header into parts delimted by commas
	parts := strings.Split(wwwAuthenticate, ",")
	for _, part := range parts {
		// Split the part into fields delimted by 'equals' sign
		fields := strings.Split(part, "=")
		// Pull out the bits we want
		if len(fields) == 2 {
			switch fields[0] {
			case "as_uri":
				authServerUrl = fields[1]
			case "ticket":
				ticket = fields[1]
			}
		}
	}

	// If we don't have the ticket and the App Server Uri then error
	if len(authServerUrl) == 0 || len(ticket) == 0 {
		authServerUrl, ticket = "", ""
		err = fmt.Errorf("failed to get as_uri and/or ticket")
	}

	return authServerUrl, ticket, err
}

//------------------------------------------------------------------------------

type UmaClient struct {
	Id     string
	Secret string
}

// ExchangeTicketForRpt exchanges the ticket for an RPT at the Authorization Server
func (umaClient *UmaClient) ExchangeTicketForRpt(requestLogger *logrus.Entry, authServer AuthorizationServer, userIdToken string, ticket string) (rpt string, forbidden bool, err error) {
	rpt = ""
	forbidden = false
	err = nil

	// Check we have a User ID Token
	if len(userIdToken) == 0 {
		err = fmt.Errorf("missing User ID Token to exchange ticket for RPT")
		requestLogger.Error(err)
		return
	}

	// Get the token endpoint
	tokenEndpoint, err := authServer.GetTokenEndpoint()
	if err != nil {
		msg := "error getting token endpoint for Authorization Server: " + authServer.url
		err = fmt.Errorf("%s: %w", msg, err)
		requestLogger.Error(err)
		return
	}
	requestLogger.Debug("Sucessfully retrieved URL for Token Endpoint: ", tokenEndpoint)

	// Prepare the request
	data := url.Values{}
	data.Set("claim_token_format", "http://openid.net/specs/openid-connect-core-1_0.html#IDToken")
	data.Set("claim_token", userIdToken)
	data.Set("ticket", ticket)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:uma-ticket")
	data.Set("client_id", umaClient.Id)
	data.Set("client_secret", umaClient.Secret)
	data.Set("scope", "openid")
	request, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		msg := "error preparing request to Token Endpoint: " + tokenEndpoint
		err = fmt.Errorf("%s: %w", msg, err)
		requestLogger.Error(err)
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Cache-Control", "no-cache")

	// Make the request
	requestLogger.Debug("Requesting RPT from token endpoint: ", tokenEndpoint)
	response, err := MakeResilentRequest(request, requestLogger, "ExchangeTicketForRpt")
	if err != nil {
		msg := "error making request to Token Endpoint: " + tokenEndpoint
		err = fmt.Errorf("%s: %w", msg, err)
		requestLogger.Error(err)
		return
	}
	if response.StatusCode != http.StatusOK {
		var msg string
		if response.StatusCode == http.StatusForbidden {
			forbidden = true
			msg = fmt.Sprintf("access request is FORBIDDEN (403) by Token Endpoint: %v", tokenEndpoint)
			requestLogger.Warn(msg)
		} else {
			msg = fmt.Sprintf("unexpected response code '%v' from Token Endpoint: %v", response.StatusCode, tokenEndpoint)
			requestLogger.Error(msg)
		}
		err = fmt.Errorf(msg)
		return
	}
	requestLogger.Debug("Token endpoint replied with 200 (OK)")

	// Read the response body
	body := response.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		err = fmt.Errorf("could not read response data from Token Endpoint %v: %w", tokenEndpoint, err)
		return
	}

	// Get the RPT from the json response
	bodyJson := struct {
		AccessToken string `json:"access_token"`
	}{}
	err = json.Unmarshal(bodyBytes, &bodyJson)
	if err != nil {
		err = fmt.Errorf("could not interpret json response from Token Endpoint %v: %w", tokenEndpoint, err)
		return
	}
	rpt = bodyJson.AccessToken
	requestLogger.Debug("Successfully extracted RPT from token endpoint response")

	return rpt, forbidden, err
}

// GetUserIdTokenBasicAuth performs basic auth to obtain an ID token with the supplied credentials
func (umaClient *UmaClient) GetUserIdTokenBasicAuth(requestLogger *logrus.Entry, authServer AuthorizationServer, username string, password string) (userIdToken string, err error) {
	userIdToken = ""
	err = nil

	// Get the token endpoint
	tokenEndpoint, err := authServer.GetTokenEndpoint()
	if err != nil {
		msg := "error getting token endpoint for Authorization Server: " + authServer.url
		requestLogger.Error(msg, ": ", err)
		err = fmt.Errorf(msg+": %w", err)
		return
	}
	requestLogger.Debug("Sucessfully retrieved URL for Token Endpoint: ", tokenEndpoint)

	// Prepare the request
	data := url.Values{}
	data.Set("scope", "openid user_name")
	data.Set("grant_type", "password")
	data.Set("username", username)
	data.Set("password", password)
	data.Set("client_id", umaClient.Id)
	data.Set("client_secret", umaClient.Secret)
	request, err := http.NewRequest("POST", tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		msg := "error preparing request to Token Endpoint: " + tokenEndpoint
		requestLogger.Error(msg, ": ", err)
		err = fmt.Errorf(msg+": %w", err)
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Cache-Control", "no-cache")

	// Make the request
	requestLogger.Debug("Requesting User ID Token from token endpoint: ", tokenEndpoint)
	response, err := HttpClient.Do(request)
	if err != nil {
		msg := "error making request to Token Endpoint: " + tokenEndpoint
		requestLogger.Error(msg, ": ", err)
		err = fmt.Errorf(msg+": %w", err)
		return
	}
	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("unexpected response code '%v' from Token Endpoint: %v", response.StatusCode, tokenEndpoint)
		requestLogger.Error(msg)
		err = fmt.Errorf(msg)
		return
	}

	// Read the response body
	body := response.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		err = fmt.Errorf("could not read response data from Token Endpoint %v: %w", tokenEndpoint, err)
		return
	}

	// Get the ID token from the json response
	bodyJson := struct {
		IdToken string `json:"id_token"`
	}{}
	err = json.Unmarshal(bodyBytes, &bodyJson)
	if err != nil {
		err = fmt.Errorf("could not interpret json response from Token Endpoint %v: %w", tokenEndpoint, err)
		return
	}
	userIdToken = bodyJson.IdToken

	return userIdToken, err
}

//------------------------------------------------------------------------------
