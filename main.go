package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
)

type Application struct {
	Logger       *slog.Logger
	ApiEndpoints []ApiEndpoint
	TokenManager TokenManager
}

func main() {

	logDir := flag.String("logdir", "logs", "Katalog zapisu logów")

	//  ENV
	os.Setenv("AUTH_URL", "https://login.xovis.cloud/oauth/token")
	os.Setenv("CLIENT_ID", "PLACEHOLDER")
	os.Setenv("CLIENT_SECRET", "PLACEHOLDER")
	os.Setenv("AUDIENCE", "https://api.xovis.cloud/aero/")
	os.Setenv("TOKEN_CACHE", "token_cache.json")

	flag.Parse()

	logger, file, err := CreateLogger(*logDir)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var Endpoints []ApiEndpoint

	app := &Application{
		Logger:       logger,
		ApiEndpoints: Endpoints,
		TokenManager: TokenManager{},
	}

	app.CreateEndpointAndAppend("SecurityMaxWaitTime", "/api/v2/sec-max-wait-time")

	app.Logger.Info("Aplikacja uruchomiona")
	app.Logger.Info(app.GetEndpoints())

	token, err := app.TokenManager.GetToken()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(token)
}
