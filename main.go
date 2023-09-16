package main

import (
	"log/slog"
	"os"

	"github.com/julienschmidt/httprouter"

	"github.com/aaronriekenberg/go-api/command"
	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/requestinfo"
	"github.com/aaronriekenberg/go-api/server"
)

func createRouter(config config.Configuration) *httprouter.Router {
	router := httprouter.New()
	router.GET("/commands", command.CreateAllCommandsHandler(config.CommandConfiguration))
	router.GET("/commands/:id", command.CreateRunCommandsHandler(config.CommandConfiguration))
	router.GET("/request_info", requestinfo.CreateHandler())

	return router
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if len(os.Args) != 2 {
		slog.Error("config file required as command line arument")
	}

	configFile := os.Args[1]

	config := config.ReadConfiguration(configFile)

	slog.Info("read configuration",
		"config", config)

	router := createRouter(config)

	server.Run(config.ServerConfiguration, router)
}
