package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"

	"github.com/guptarohit/asciigraph"
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
	_, err = app.GetValidToken()
	if err != nil {
		log.Fatal(err)
	}
	// Try to fetch data
	params := []string{"2025-06-15T15:30:00", "2025-06-15T16:30:00", "PT1M", "Waiting Time Display", "MAX"}

	var HResponse *HistResponse
	HResponse, err = FetchHistData(params)
	if err != nil {
		log.Fatal(err)
	}
	pred_waiting_time, err := HResponse.GetValuesByKPIType("PREDICTED_WAITING_TIME")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%+v", pred_waiting_time)
	values := make([]float64, len(pred_waiting_time))
	for i, v := range pred_waiting_time {
		values[i] = float64(v)
	}
	graph := asciigraph.Plot(values, asciigraph.Height(10), asciigraph.Caption("Predicted WT"))
	fmt.Println(graph)
}
