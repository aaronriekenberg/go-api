package main

import (
	"log/slog"
	"os"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/handlers"
	"github.com/aaronriekenberg/go-api/server"
)

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if len(os.Args) != 2 {
		slog.Error("config file required as command line arument")
		os.Exit(1)
	}

	configFile := os.Args[1]

	config := config.ReadConfiguration(configFile)

	slog.Info("read configuration",
		"config", config)

	handlers := handlers.CreateHandlers(config)

	server.Run(config.ServerConfiguration, handlers)
}
