package uma

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

type AuthorizationServer struct {
	url           string
	tokenEndpoint string
}

type AuthorizationServerList struct {
	sync.Map
}

var AuthorizationServers = AuthorizationServerList{}

func NewAuthorizationServer(url string) *AuthorizationServer {
	return &AuthorizationServer{url: url}
}

func (authServer *AuthorizationServer) GetUrl() string {
	return authServer.url
}

// GetTokenEndpoint performs a lookup (HTTP GET) on the Authorization Server via
// its AS URL, to retrieve the Token Endpoint from the UMA configuration endpoint
func (authServer *AuthorizationServer) GetTokenEndpoint() (tokenEndpointUrl string, err error) {
	tokenEndpointUrl = ""
	err = nil

	// If we have it already, then return it
	if len(authServer.tokenEndpoint) > 0 {
		tokenEndpointUrl = authServer.tokenEndpoint
		return
	}

	// Fetch the UMA configuration from the Auth Server
	umaConfigUrl := authServer.url + "/.well-known/uma2-configuration"
	response, err := HttpClient.Get(umaConfigUrl)
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
	authServer.tokenEndpoint = bodyJson.TokenEndpoint
	tokenEndpointUrl = authServer.tokenEndpoint
	return
}
