package uma

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

type UmaClient struct {
	Id     string
	Secret string
}

// Exchange the ticket for an RPT at the Authorization Server
func (umaClient *UmaClient) ExchangeTicketForRpt(authServer AuthorizationServer, userIdToken string, ticket string) (rpt string, err error) {
	rpt = ""
	err = nil

	// Get the token endpoint
	tokenEndpoint, err := authServer.GetTokenEndpoint()
	if err != nil {
		msg := "error getting token endpoint for Authorization Server: " + authServer.url
		log.Error(msg, ": ", err)
		err = fmt.Errorf(msg+": %w", err)
		return
	}

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
		log.Error(msg, ": ", err)
		err = fmt.Errorf(msg+": %w", err)
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Cache-Control", "no-cache")

	// Make the request
	log.Debug("Requesting RPT from token endpoint: ", tokenEndpoint)
	response, err := HttpClient.Do(request)
	if err != nil {
		msg := "error making request to Token Endpoint: " + tokenEndpoint
		log.Error(msg, ": ", err)
		err = fmt.Errorf(msg+": %w", err)
		return
	}
	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("unexpected response code '%v' from Token Endpoint: %v", response.StatusCode, tokenEndpoint)
		log.Error(msg)
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

	return rpt, err
}
