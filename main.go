package main

import (
	"log/slog"
	"os"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/handlers"
	"github.com/aaronriekenberg/go-api/profiling"
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

func setupSlog() {
	level := slog.LevelInfo

	if levelString, ok := os.LookupEnv("LOG_LEVEL"); ok {
		err := level.UnmarshalText([]byte(levelString))
		if err != nil {
			fatalError("level.UnmarshalText error", err)
		}
	}

	slog.SetDefault(
		slog.New(
			slog.NewJSONHandler(
				os.Stdout, &slog.HandlerOptions{
					Level: level,
				},
			),
		),
	)

	slog.Info("setupSlog",
		"configuredLevel", level,
	)
}

func main() {
	setupSlog()

	if len(os.Args) != 2 {
		fatalError("config file required as command line arument", nil)
	}

	configFile := os.Args[1]

	config, err := config.ReadConfiguration(configFile)

	if err != nil {
		fatalError("main: config.ReadConfiguration error", err)
	}

	profiling.Start(config.ProfilingConfiguration)

	handlers, err := handlers.CreateHandlers(*config)

	if err != nil {
		fatalError("main: handlers.CreateHandlers error", err)
	}

	err = server.Run(config.ServerConfiguration, handlers)
	fatalError("main: server.Run returned error", err)
}
