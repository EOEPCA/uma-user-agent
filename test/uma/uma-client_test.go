package uma_test

import (
	"testing"

	"github.com/EOEPCA/uma-user-agent/src/uma"
)

func TestExchangeTicketForRpt(t *testing.T) {
	// Test data
	userIdToken := "eyJraWQiOiJkYjVlMGM5Ni0yZGQwLTQyNjktYTA5OC02NWZkNzIyNjFjMGFfc2lnX3JzMjU2IiwidHlwIjoiSldUIiwiYWxnIjoiUlMyNTYifQ.eyJhdWQiOiI2OTE4OTJmZC0xZTU4LTRlNDQtODM1NS0xZjNiNzYzNGFmOGYiLCJzdWIiOiJkMzY4OGRhYS0zODVkLTQ1YjAtOGUwNC0yMDYyZTNlMmNkODYiLCJ1c2VyX25hbWUiOiJlcmljIiwiaXNzIjoiaHR0cHM6Ly90ZXN0LmRlbW8uZW9lcGNhLm9yZyIsImV4cCI6MTYyNzkyNTA5NSwiaWF0IjoxNjI3OTIxNDk1LCJveE9wZW5JRENvbm5lY3RWZXJzaW9uIjoib3BlbmlkY29ubmVjdC0xLjAifQ.oIb8sfDXBxdO1wsXOpYqhp_zhk1oCtWdTQqGDHQ0H2KKMM4rujHtLqFwidK3-pq2sOl-xMr4heZIWngwlxRTGj3aSy4DOKm64h7PIC1Ho2K-KtHC55yPgwEeLsLsFVEQ6uvd1HooqYokYIfEGjvv6elybrT3u46ScPS8MBTjP9vCPpWPBAJtXcms2yc531eJv4sXG-M-9QAhTzVZv99e5wVTa8Nh1CmnkpoWNcaPEiMLHP_4fXH8s_C3yOUs_NwKGmQ20MqXYQEC2KD6AlmGQy4UEJ13WcSk3ImQcqJsWTLJxqwKvJIsBAxwzGxZXeqbSgrkzs6_-Eu01FNreVmfLQ"
	authServerUrl := "https://test.demo.eoepca.org"
	ticket := "e24108e6-2a38-4908-bb9f-f5a8f15cec3f"

	authServer := uma.NewAuthorizationServer(authServerUrl)

	clientId := "691892fd-1e58-4e44-8355-1f3b7634af8f"
	clientSecret := "eb28831f-3e0f-4580-a103-8fd1e0adbb3c"
	umaClient := uma.UmaClient{Id: clientId, Secret: clientSecret}

	rpt, err := umaClient.ExchangeTicketForRpt(*authServer, userIdToken, ticket)
	if err != nil {
		t.Error(err)
	} else {
		t.Log("RPT =", rpt)
	}
}
