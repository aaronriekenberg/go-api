package main

import (
	"log/slog"
	"os"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/handlers"
	"github.com/aaronriekenberg/go-api/server"
)

func fatalError(
	message string,
	err error,
) {
	slog.Error(message,
		"error", err,
	)
	os.Exit(1)
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if len(os.Args) != 2 {
		fatalError("config file required as command line arument", nil)
	}

	configFile := os.Args[1]

	config, err := config.ReadConfiguration(configFile)

	if err != nil {
		fatalError("config.ReadConfiguration error", err)
	}

	handlers, err := handlers.CreateHandlers(*config)

	if err != nil {
		fatalError("handlers.CreateHandlers error", err)
	}

	err = server.Run(config.ServerConfiguration, handlers)
	fatalError("server.Run error", err)
}
