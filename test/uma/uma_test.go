package uma_test

import (
	"fmt"
	"testing"

	"github.com/EOEPCA/uma-user-agent/src/uma"
)

func TestUnpackWwwAuthenticateHeader(t *testing.T) {
	expectedAsUri := "https://test.185.52.193.87.nip.io"
	expectedTicket := "0a6f8d80-d618-44b1-b9e4-77a425d4981d"
	testHeader := fmt.Sprintf("realm=eoepca,as_uri=%v,ticket=%v", expectedAsUri, expectedTicket)

	as_uri, ticket, err := uma.UnpackWwwAuthenticateHeader(testHeader)
	if err != nil {
		t.Error(err)
	}
	if as_uri != expectedAsUri {
		t.Errorf("unexpected value for as_uri: %v", as_uri)
	}
	if ticket != expectedTicket {
		t.Errorf("unexpected value for ticket: %v", as_uri)
	}
}

func TestLookupTokenEndpoint(t *testing.T) {
	expectedTokenEndpointUrl := "https://test.185.52.193.87.nip.io/oxauth/restv1/token"
	authServerUrl := "https://test.185.52.193.87.nip.io"

	tokenEndpointUrl, err := uma.LookupTokenEndpoint(authServerUrl)
	if err != nil {
		t.Error(err)
	}
	if tokenEndpointUrl != expectedTokenEndpointUrl {
		t.Errorf("unexpected value for tokenEndpointUrl: %v", tokenEndpointUrl)
	}
}
