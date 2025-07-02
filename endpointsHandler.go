package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type ApiEndpoint struct {
	Name            string
	EndpointAddress string
}

func (app *Application) GetEndpoints() string {
	var sb strings.Builder
	sb.WriteString("Endpoints list:\n")

	for _, ep := range app.ApiEndpoints {
		sb.WriteString("Name: " + ep.Name + " -- Address: " + ep.EndpointAddress + "\n")
	}

	return sb.String()
}
func NewApiEndPoint(Name, Address string) *ApiEndpoint {
	return &ApiEndpoint{
		Name:            Name,
		EndpointAddress: Address,
	}
}
func (ae *ApiEndpoint) GetAddress() string {
	return ae.EndpointAddress
}
func (app *Application) GetAddressByName(name string) string {
	for _, endpoint := range app.ApiEndpoints {
		if endpoint.Name == name {
			return endpoint.EndpointAddress
		}
	}
	app.Logger.Warn("API Endpoint %v not found in endpoints list", name, "")
	return "NOT FOUND"
}

func (app *Application) CreateEndpointAndAppend(name, address string) {
	ep := NewApiEndPoint(name, address)
	app.ApiEndpoints = append(app.ApiEndpoints, *ep)
}

// authorization method
// Struct for keeping the information about token

type TokenManager struct {
	Token  string
	mu     sync.RWMutex
	expiry time.Time
}

type TokenCache struct {
	Token  string    `json:"token"`
	Expiry time.Time `json:"expiry"`
}

// Get New Token only used by GetToken if necessary
func (tm *TokenManager) getNewToken() (string, time.Time, error) {
	payload := map[string]string{
		"client_id":     os.Getenv("CLIENT_ID"),
		"client_secret": os.Getenv("CLIENT_SECRET"),
		"audience":      os.Getenv("AUDIENCE"),
		"grant_type":    "client_credentials",
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", time.Time{}, err
	}

	req, err := http.NewRequest("POST", os.Getenv("AUTH_URL"), bytes.NewReader(bodyBytes))
	if err != nil {
		return "", time.Time{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("auth server returned status %d", resp.StatusCode)
	}

	var respData struct {
		AccessToken string `json:"access_token`
		TokenType   string `json:"token_type`
	}

	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", time.Time{}, err
	}

	expiryTime := time.Now().Add(24 * time.Hour)

	cache := TokenCache{
		Token:  respData.AccessToken,
		Expiry: expiryTime,
	}
	cacheBytes, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		_ = os.WriteFile(os.Getenv("TOKEN_CACHE"), cacheBytes, 0600)
	}

	return respData.AccessToken, expiryTime, nil
}

func (tm *TokenManager) GetToken() (string, error) {
	tm.mu.RLock()

	if tm.Token != "" && time.Now().Before(tm.expiry) {
		return tm.Token, nil
	}

	newToken, newExpiry, err := tm.getNewToken()
	if err != nil {
		return "", err
	}

	tm.Token = newToken
	tm.expiry = newExpiry
	return tm.Token, nil
}
