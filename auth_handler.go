package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type AuthToken struct {
	Token     string `json:"access_token"`
	ExpiresIn int64  `json:"expires_in"`
	CreatedAt time.Time
}

// Auth for receving bearer token
func (app *Application) Auth() (*AuthToken, error) {
	client := &http.Client{}
	url := os.Getenv("AUTH_URL")
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	payload := map[string]string{
		"client_id":     os.Getenv("CLIENT_ID"),
		"client_secret": os.Getenv("CLIENT_SECRET"),
		"audience":      os.Getenv("AUDIENCE"),
		"grant_type":    "client_credentials",
	}
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		app.Logger.Error("Błąd podczas serializacji payloadu", slog.String("error", err.Error()))
	}
	app.Logger.Info("JSON Payload", slog.String("payload", string(jsonPayload)))

	// build request
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonPayload))
	if err != nil {
		app.Logger.Error("Błąd podczas tworzenia żądania HTTP", slog.String("error", err.Error()))
		return nil, err
	}
	// set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		app.Logger.Error("Błąd podczas uzyskiwania tokena", slog.String("status", resp.Status))
		return nil, err
	}

	// Log response
	body, _ := io.ReadAll(resp.Body)
	app.Logger.Info("Odpowiedź z serwera", slog.String("status", resp.Status), slog.String("body", string(body)))
	var token AuthToken
	if err := json.Unmarshal(body, &token); err != nil {
		app.Logger.Error("Błąd podczas deserializacji odpowiedzi", slog.String("error", err.Error()))
		return nil, err
	}
	token.CreatedAt = time.Now()
	return &token, nil
}

func saveTokenToFile(token *AuthToken, filename string) error {
	data, err := json.Marshal(token)
	if err != nil {
		return err
	}
	return os.WriteFile(filename, data, 0600)
}
func loadTokenFromFile(filename string) (*AuthToken, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var token AuthToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, err
	}
	return &token, nil
}

func tokenExpired(token *AuthToken) bool {
	return time.Since(token.CreatedAt) > time.Duration(token.ExpiresIn)*time.Second
}

func (app *Application) GetValidToken() (*AuthToken, error) {
	token, err := loadTokenFromFile(os.Getenv("TOKEN_CACHE"))
	if err != nil || tokenExpired(token) {
		app.Logger.Info("Token wygasł lub nie istnieje, uzyskiwanie nowego tokena")
		token, err = app.Auth()
		if err != nil {
			return nil, err
		}
		if err := saveTokenToFile(token, os.Getenv("TOKEN_CACHE")); err != nil {
			app.Logger.Error("Błąd podczas zapisywania tokena do pliku", slog.String("error", err.Error()))
			return nil, err
		}
	} else {
		app.Logger.Info("Używanie istniejącego tokena", slog.String("token", token.Token))
	}
	return token, nil
}
