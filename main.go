package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/http"
)

type Application struct {
	Logger       *slog.Logger
	ApiEndpoints []ApiEndpoint
}

func main() {

	logDir := flag.String("logdir", "logs", "Katalog zapisu logów")

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
	}

	app.CreateEndpointAndAppend("SecurityMaxWaitTime", "/api/v2/sec-max-wait-time")

	app.Logger.Info("Aplikacja uruchomiona")
	app.Logger.Info(app.GetEndpoints())

	endAddress := "https:/" + app.GetAddressByName("SecurityMaxWaitTime")

	app.SendAuthorizationRequest(context.Background(), http.MethodGet, endAddress, nil, APIAuthConfig{AuthType: AuthBasic, Username: "user", Password: "pass"})

}
