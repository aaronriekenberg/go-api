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

func CreateHandlers() http.Handler {

	mux := http.NewServeMux()

	config := config.ConfigurationInstance()

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

	handleAPIGET("/commands", command.NewAllCommandsHandler())

	handleAPIGET("/commands/{id}", command.NewRunCommandsHandler())

	handleAPIGET("/connection_info", connectioninfo.NewConnectionInfoHandler())

	handleAPIGET("/request_info", requestinfo.NewRequestInfoHandler())

	handleAPIGET("/version_info", versioninfo.NewVersionInfoHandler())

	return requestlogging.NewRequestLogger(mux)
}
