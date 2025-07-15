package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"

	"github.com/joho/godotenv"
)

type Application struct {
	Logger *slog.Logger
	Token  *AuthToken
}

func main() {

	logDir := flag.String("logdir", "logs", "Katalog zapisu logów")

	flag.Parse()

	// load dot env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	logger, file, err := CreateLogger(*logDir)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	app := &Application{
		Logger: logger,
	}
	token, err := app.GetValidToken()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Access Token:", token.Token)
}
