package handlers

import (
	"log/slog"
	"net/http"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/handlers/command"
	"github.com/aaronriekenberg/go-api/handlers/connectioninfo"
	"github.com/aaronriekenberg/go-api/handlers/requestinfo"
	"github.com/aaronriekenberg/go-api/handlers/versioninfo"
)

func CreateHandlers(
	config config.Configuration,
) http.Handler {
	mux := http.NewServeMux()

	context := config.ServerConfiguration.Context

	slog.Info("CreateHandlers",
		"context", context,
	)

	mux.Handle("GET "+context+"/commands", command.NewAllCommandsHandler(config.CommandConfiguration))

	mux.Handle("GET "+context+"/commands/{id}", command.NewRunCommandsHandler(config.CommandConfiguration))

	mux.Handle("GET "+context+"/connection_info", connectioninfo.NewConnectionInfoHandler())

	mux.Handle("GET "+context+"/request_info", requestinfo.NewRequestInfoHandler())

	mux.Handle("GET "+context+"/version_info", versioninfo.NewVersionInfoHandler())

	return mux
}
