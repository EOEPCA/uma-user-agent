package uma

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/sirupsen/logrus"
)

var AuthorizationServers = NewAuthorizationServerList()

//------------------------------------------------------------------------------

// AuthorizationServer represents a single Authorization Server
type AuthorizationServer struct {
	url           string
	tokenEndpoint string
}

//------------------------------------------------------------------------------

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

//------------------------------------------------------------------------------

// AuthorizationServerList is a thread-safe collection of Authorization Servers
type AuthorizationServerList struct {
	rwMutex     sync.RWMutex
	authServers map[string]AuthorizationServer
}

//------------------------------------------------------------------------------

func NewAuthorizationServerList() *AuthorizationServerList {
	return &AuthorizationServerList{authServers: make(map[string]AuthorizationServer)}
}

// Delete deletes the value for a key
func (asl *AuthorizationServerList) Delete(key string) {
	asl.rwMutex.Lock()
	defer asl.rwMutex.Unlock()
	delete(asl.authServers, key)
}

// Load returns the value stored in the map for a key.
// The ok result indicates whether value was found in the map.
func (asl *AuthorizationServerList) Load(key string) (value AuthorizationServer, ok bool) {
	asl.rwMutex.RLock()
	defer asl.rwMutex.RUnlock()
	value, ok = asl.authServers[key]
	return
}

// LoadAndDelete deletes the value for a key, returning the previous value if any.
// The loaded result reports whether the key was present.
func (asl *AuthorizationServerList) LoadAndDelete(key string) (value AuthorizationServer, loaded bool) {
	if value, loaded = asl.Load(key); loaded {
		asl.Delete(key)
	}
	return
}

// LoadOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (asl *AuthorizationServerList) LoadOrStore(requestLogger *logrus.Entry, key string, value AuthorizationServer) (actual AuthorizationServer, loaded bool) {
	if actual, loaded = asl.Load(key); !loaded {
		actual = value
		asl.Store(requestLogger, key, actual)
	} else {
		requestLogger.Tracef("Using existing cache entry for Authorization Server: %v", actual.url)
	}
	return
}

// Store sets the value for a key.
func (asl *AuthorizationServerList) Store(requestLogger *logrus.Entry, key string, value AuthorizationServer) {
	asl.rwMutex.Lock()
	defer asl.rwMutex.Unlock()
	asl.authServers[key] = value
	requestLogger.Infof("Authorization Server stored in the cache: %v", value.url)
}

//------------------------------------------------------------------------------
