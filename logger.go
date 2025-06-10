package main

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

func CreateLogger(logDir string) (*slog.Logger, *os.File, error) {
	// Osobny log dla kazdego dnia
	logDay := time.Now().Local().Format("2006-01-02")

	fullPath := filepath.Join(logDir, (logDay + ".log"))

	err := os.MkdirAll(logDir, os.ModePerm)
	if err != nil {
		return nil, nil, errors.New("Nie można utworzyć katalogu logów: " + err.Error())
	}
	file, err := os.OpenFile(fullPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, nil, errors.New("Nie można otworzyć pliku logów: " + err.Error())
	}

	// Multiwriter do zapisu jednoczesnie w terminalu i pliku
	writer := io.MultiWriter(os.Stdout, file)

	// JSON logger
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key != "time" {
				return a
			}
			t := a.Value.Time()
			a.Value = slog.StringValue(t.Format("2006-01-02 15:04:05"))

			return a
		},
	})

	logger := slog.New(handler)

	return logger, file, nil

}
