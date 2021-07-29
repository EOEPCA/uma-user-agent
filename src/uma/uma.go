package uma

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// UMA realm=eoepca,as_uri=https://test.185.52.193.87.nip.io,ticket=0a6f8d80-d618-44b1-b9e4-77a425d4981d
//
// /.well-known/uma2-configuration

func UnpackWwwAuthenticateHeader(wwwAuthenticate string) (as_uri string, ticket string, err error) {
	as_uri = ""
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
				as_uri = fields[1]
			case "ticket":
				ticket = fields[1]
			}
		}
	}

	// If we don't have the ticket and the App Server Uri then error
	if len(as_uri) == 0 || len(ticket) == 0 {
		as_uri, ticket = "", ""
		err = fmt.Errorf("failed to get as_uri and/or ticket")
	}

	return as_uri, ticket, err
}

func LookupTokenEndpoint(authServerUrl string) (tokenEndpointUrl string, err error) {
	tokenEndpointUrl = ""
	err = nil

	// Fetch the UMA configuration from the Auth Server
	url := authServerUrl + "/.well-known/uma2-configuration"
	response, err := http.Get(url)
	if err != nil {
		err = fmt.Errorf("could not retieve UMA service details from %v: %w", url, err)
		return
	}

	// Read the response body
	body := response.Body
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		err = fmt.Errorf("could not read response data from %v: %w", url, err)
		return
	}

	// Interpret as json response
	bodyJson := struct {
		TokenEndpoint string `json:"token_endpoint"`
	}{}
	err = json.Unmarshal(bodyBytes, &bodyJson)
	if err != nil {
		err = fmt.Errorf("could interpret json response from %v: %w", url, err)
		return
	}

	tokenEndpointUrl = bodyJson.TokenEndpoint
	return
}
