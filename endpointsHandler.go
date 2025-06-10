package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

// Endpoint Auth request
type AuthType int

const (
	AuthNone AuthType = iota
	AuthBearer
	AuthBasic
	AuthAPIKey
)

type APIAuthConfig struct {
	AuthType   AuthType
	Token      string // Bearer or API Key
	Username   string // Basic Auth
	Password   string // Basic Auth
	APIKeyName string // Header for API Key
}

func (app *Application) SendAuthorizationRequest(ctx context.Context, method, url string, body interface{}, auth APIAuthConfig) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBytes)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	switch auth.AuthType {
	case AuthBearer:
		app.Logger.Info("Auth type Bearer for %v", url)
		req.Header.Set("Authorization", "Bearer "+auth.Token)
	case AuthBasic:
		app.Logger.Info("Auth type Basic for %v", url)
		req.Header.Set(auth.Username, auth.Password)
	case AuthAPIKey:
		app.Logger.Info("Auth type ApiKey for %v", url)
		req.Header.Set(auth.APIKeyName, auth.Token)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	return resp, nil

}
