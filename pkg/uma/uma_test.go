package uma_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/EOEPCA/uma-user-agent/pkg/config"
	"github.com/EOEPCA/uma-user-agent/pkg/uma"
	"github.com/sirupsen/logrus"
)

// Server and Client
// var authServerUrl = "https://test.185.52.193.87.nip.io"
var authServerUrl = "https://test.demo.eoepca.org"
var authServer = uma.NewAuthorizationServer(authServerUrl)

// var umaClient = uma.UmaClient{Id: "22ba0c56-9780-4b0b-ad71-d745c166ca3b", Secret: "0e3e1d0d-9002-4d44-bbff-a170efa18512"}
var umaClient = uma.UmaClient{Id: config.GetClientId(), Secret: config.GetClientSecret()}

// Test data
var username = "eric"
var password = "defaultPWD"
var userIdToken = ""
var testTicket = "b33f6aff-ac5c-403f-96aa-b1aff58488cf"

var testLogger = logrus.NewEntry(logrus.New())

// TestMain performs testing setup
func TestMain(m *testing.M) {
	setup()
	os.Exit(m.Run())
}

func setup() {
	var err error
	userIdToken, err = umaClient.GetUserIdTokenBasicAuth(testLogger, *authServer, username, password)
	if err != nil {
		testLogger.Errorf("Could not initialise user ID token: %v", err)
	}
	testLogger.Debugf("User ID token: %v", userIdToken)
}

// TestLookupAuthServer tests store and retrieve from the AuthServer cache
func TestLookupAuthServer(t *testing.T) {
	doLoadOrStore := func(context string, expectLoaded bool) {
		_authServer, loaded := uma.AuthorizationServers.LoadOrStore(testLogger, authServerUrl, *uma.NewAuthorizationServer(authServerUrl))
		if len(_authServer.GetUrl()) == 0 {
			t.Errorf("[%v] error getting the Authorization Server details", context)
		} else if loaded != expectLoaded {
			t.Errorf("[%v] loaded status discrepency=> expected: %v, got: %v", context, expectLoaded, loaded)
		} else {
			t.Logf("[%v] successfully retrieved Authorization Server: %v", context, _authServer)
		}
	}
	doLoadOrStore("Attempt#1", false)
	doLoadOrStore("Attempt#2", true)
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
	rpt, _, err := umaClient.ExchangeTicketForRpt(testLogger, *authServer, userIdToken, testTicket)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("RPT =", rpt)
	}
}
