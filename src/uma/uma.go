package uma

import (
	"fmt"
	"strings"
)

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
