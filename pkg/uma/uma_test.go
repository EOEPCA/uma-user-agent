package uma_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/EOEPCA/uma-user-agent/pkg/config"
	"github.com/EOEPCA/uma-user-agent/pkg/uma"
	log "github.com/sirupsen/logrus"
)

// Server and Client
var authServerUrl = "https://test.185.52.193.87.nip.io"
var authServer = uma.NewAuthorizationServer(authServerUrl)

// var umaClient = uma.UmaClient{Id: "22ba0c56-9780-4b0b-ad71-d745c166ca3b", Secret: "0e3e1d0d-9002-4d44-bbff-a170efa18512"}
var umaClient = uma.UmaClient{Id: config.Config.ClientId, Secret: config.Config.ClientSecret}

// Test data
var username = "eric"
var password = "defaultPWD"
var userIdToken = ""
var testTicket = "bbaed2cf-ae16-433f-862c-d7fbc2d758ef"

// TestMain performs testing setup
func TestMain(m *testing.M) {
	setup()
	os.Exit(m.Run())
}

func setup() {
	var err error
	userIdToken, err = umaClient.GetUserIdTokenBasicAuth(*authServer, username, password)
	if err != nil {
		log.Errorf("Could not initialise user ID token: %v", err)
	}
	log.Debugf("User ID token: %v", userIdToken)
}

// TestGetTokenEndpoint tests getting the Token Endpoint from the Authorization Server
func TestGetTokenEndpoint(t *testing.T) {
	expectedTokenEndpointUrl := authServerUrl + "/oxauth/restv1/token"

	authServer := uma.NewAuthorizationServer(authServerUrl)
	tokenEndpointUrl, err := authServer.GetTokenEndpoint()
	if err != nil {
		t.Error(err)
	}
	if tokenEndpointUrl != expectedTokenEndpointUrl {
		t.Errorf("unexpected value returned for tokenEndpointUrl: %v", tokenEndpointUrl)
	}
}

// TestUnpackWwwAuthenticateHeader tests getting the 'parts' from a Www-Authenticate header
func TestUnpackWwwAuthenticateHeader(t *testing.T) {
	expectedAsUri := authServerUrl
	expectedTicket := testTicket
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

// TestExchangeTicketForRpt tests getting the RPT from the Token Endpoint using a Ticket
func TestExchangeTicketForRpt(t *testing.T) {
	rpt, err := umaClient.ExchangeTicketForRpt(*authServer, userIdToken, testTicket)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("RPT =", rpt)
	}
}
