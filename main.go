package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"

	"github.com/aaronriekenberg/go-api/command"
	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/requestinfo"
	"github.com/aaronriekenberg/go-api/server"
)

func createRouter(config config.Configuration) *httprouter.Router {
	router := httprouter.New()

	router.Handler(http.MethodGet, "/commands", command.NewAllCommandsHandler(config.CommandConfiguration))
	router.Handler(http.MethodGet, "/commands/:id", command.NewRunCommandsHandler(config.CommandConfiguration))

	router.Handler(http.MethodGet, "/request_info", requestinfo.NewHandler())

	return router
}

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

	router := createRouter(config)

	server.Run(config.ServerConfiguration, router)
}
