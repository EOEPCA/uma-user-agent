package uma_test

import (
	"os"
	"testing"

	"github.com/EOEPCA/uma-user-agent/src/uma"
	log "github.com/sirupsen/logrus"
)

// Server and Client
var authServer = uma.NewAuthorizationServer("https://test.185.52.193.87.nip.io")
var umaClient = uma.UmaClient{Id: "22ba0c56-9780-4b0b-ad71-d745c166ca3b", Secret: "0e3e1d0d-9002-4d44-bbff-a170efa18512"}

// Test user
var username = "eric"
var password = "defaultPWD"
var userIdToken = ""

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

func TestExchangeTicketForRpt(t *testing.T) {
	ticket := "6b6d590b-a56a-4c7f-bbe6-19a8b6554aa4"

	rpt, err := umaClient.ExchangeTicketForRpt(*authServer, userIdToken, ticket)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("RPT =", rpt)
	}
}
