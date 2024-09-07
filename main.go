package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/handlers"
	"github.com/aaronriekenberg/go-api/profiling"
	"github.com/aaronriekenberg/go-api/server"
	"github.com/aaronriekenberg/go-api/version"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic in main",
				"error", err,
			)
			os.Exit(1)
		}
	}()

	setupSlog()

	slog.Info("begin main",
		"buildInfoMap", version.BuildInfoMap(),
	)

	if len(os.Args) != 2 {
		panic("config file required as command line arument")
	}

	configFile := os.Args[1]

	config, err := config.ReadConfiguration(configFile)
	if err != nil {
		panic(fmt.Errorf("main: config.ReadConfiguration error: %w", err))
	}

	if config.GoMaxProcs == -1 {
		maxProcs := runtime.GOMAXPROCS(-1)
		slog.Info("using default maxProcs",
			"maxProcs", maxProcs,
		)
	} else {
		runtime.GOMAXPROCS(config.GoMaxProcs)
		maxProcs := config.GoMaxProcs
		slog.Info("set maxProcs",
			"maxProcs", maxProcs,
		)
	}

	profiling.Start(config.ProfilingConfiguration)

	handlers := handlers.CreateHandlers(*config)

	err = server.Run(config.ServerConfiguration, handlers)
	panic(fmt.Errorf("main: server.Run error: %w", err))
}

func setupSlog() {
	level := slog.LevelInfo

	if levelString, ok := os.LookupEnv("LOG_LEVEL"); ok {
		err := level.UnmarshalText([]byte(levelString))
		if err != nil {
			panic(fmt.Errorf("level.UnmarshalText error %w", err))
		}
	}

	logToStdout := false
	if logToStdoutString, ok := os.LookupEnv("DEFAULT_LOG_TO_STDOUT"); ok {
		if strings.ToLower(logToStdoutString) == "true" {
			logToStdout = true
		}
	}

	var writer io.Writer
	if logToStdout {
		writer = os.Stdout
	} else {
		writer = &lumberjack.Logger{
			Filename:   "logs/default.log",
			MaxSize:    10,
			MaxBackups: 10,
		}
	}

	slog.SetDefault(
		slog.New(
			slog.NewJSONHandler(
				writer,
				&slog.HandlerOptions{
					Level: level,
				},
			),
		),
	)

	slog.Info("setupSlog",
		"configuredLevel", level,
	)
}
