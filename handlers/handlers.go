package handlers

import (
	"log/slog"
	"net/http"
	"path"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/handlers/command"
	"github.com/aaronriekenberg/go-api/handlers/connectioninfo"
	"github.com/aaronriekenberg/go-api/handlers/health"
	"github.com/aaronriekenberg/go-api/handlers/requestinfo"
	"github.com/aaronriekenberg/go-api/handlers/requestlogging"
	"github.com/aaronriekenberg/go-api/handlers/versioninfo"
)

func CreateHandlers(
	config config.Configuration,
) http.Handler {
	mux := http.NewServeMux()

	apiContext := config.ServerConfiguration.APIContext

	slog.Info("CreateHandlers",
		"apiContext", apiContext,
	)

	mux.Handle("GET /health", health.NewHealthHandler())

	handleAPIGET := func(
		relativePath string,
		handler http.Handler,
	) {
		mux.Handle("GET "+path.Join(apiContext, relativePath), handler)
	}

	handleAPIGET("/commands", command.NewAllCommandsHandler(config.CommandConfiguration))

	handleAPIGET("/commands/{id}", command.NewRunCommandsHandler(config.CommandConfiguration))

	handleAPIGET("/connection_info", connectioninfo.NewConnectionInfoHandler())

	handleAPIGET("/request_info", requestinfo.NewRequestInfoHandler())

	handleAPIGET("/version_info", versioninfo.NewVersionInfoHandler())

	return requestlogging.NewRequestLogger(config.RequestLoggingConfiguration, mux)
}
