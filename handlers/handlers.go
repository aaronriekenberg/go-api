package handlers

import (
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

	mux.Handle("GET /api/v1/commands", command.NewAllCommandsHandler(config.CommandConfiguration))

	mux.Handle("GET /api/v1/commands/{id}", command.NewRunCommandsHandler(config.CommandConfiguration))

	mux.Handle("GET /api/v1/connection_info", connectioninfo.NewConnectionInfoHandler())

	mux.Handle("GET /api/v1/request_info", requestinfo.NewRequestInfoHandler())

	mux.Handle("GET /api/v1/version_info", versioninfo.NewVersionInfoHandler())

	return mux
}
