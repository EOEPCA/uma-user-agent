package uma

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type AuthorizationServer struct {
	Url           string
	TokenEndpoint string
}

type AuthorizationServerList sync.Map
var AuthorizationServers = AuthorizationServerList{}

// GetTokenEndpoint performs a lookup (HTTP GET) on the Authorization Server via
// its AS URL, to retrieve the Token Endpoint from the UMA configuration endpoint
func (authServer *AuthorizationServer) GetTokenEndpoint() (tokenEndpointUrl string, err error) {
	tokenEndpointUrl = ""
	err = nil

	// If we have it already, then return it
	if len(authServer.TokenEndpoint) > 0 {
		tokenEndpointUrl = authServer.TokenEndpoint
		return
	}

	// Fetch the UMA configuration from the Auth Server
	umaConfigUrl := authServer.Url + "/.well-known/uma2-configuration"
	response, err := http.Get(umaConfigUrl)
	if err != nil {
		err = fmt.Errorf("could not retieve UMA service details from %v: %w", umaConfigUrl, err)
		return
	}

	// Read the response body
	body := response.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		err = fmt.Errorf("could not read response data from %v: %w", umaConfigUrl, err)
		return
	}

	// Interpret as json response
	bodyJson := struct {
		TokenEndpoint string `json:"token_endpoint"`
	}{}
	err = json.Unmarshal(bodyBytes, &bodyJson)
	if err != nil {
		err = fmt.Errorf("could not interpret json response from %v: %w", umaConfigUrl, err)
		return
	}

	// Check the Token Endpoint is non-empty
	if len(bodyJson.TokenEndpoint) == 0 {
		err = fmt.Errorf("blank Token Endpoint retrieved from %v", umaConfigUrl)
		return
	}

	// Record the retrieved Url and return it
	authServer.TokenEndpoint = bodyJson.TokenEndpoint
	tokenEndpointUrl = authServer.TokenEndpoint
	return
}
