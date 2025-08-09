package main

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/handlers"
	"github.com/aaronriekenberg/go-api/profiling"
	"github.com/aaronriekenberg/go-api/server"
	"github.com/aaronriekenberg/go-api/version"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic in main",
				"error", err,
			)
			fmt.Fprintf(os.Stderr, "stack trace:\n%v", string(debug.Stack()))
			os.Exit(1)
		}
	}()

	setupSlog()

	slog.Info("begin main",
		"buildInfoMap", version.BuildInfoMap(),
		"goEnvironVariables", goEnvironVariables(),
		"GOMAXPROCS", runtime.GOMAXPROCS(0),
		"NumCPU", runtime.NumCPU(),
	)

	if len(os.Args) != 2 {
		panic("config file required as command line arument")
	}

	configFile := os.Args[1]

	config, err := config.ReadConfiguration(configFile)
	if err != nil {
		panic(fmt.Errorf("main: config.ReadConfiguration error: %w", err))
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

	slog.SetDefault(
		slog.New(
			slog.NewJSONHandler(
				os.Stdout,
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

func goEnvironVariables() []string {
	var goVars []string
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "GO") {
			goVars = append(goVars, env)
		}
	}
	return goVars
}
