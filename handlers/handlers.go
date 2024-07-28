package handlers

import (
	"log/slog"
	"net/http"
	"path"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/handlers/command"
	"github.com/aaronriekenberg/go-api/handlers/connectioninfo"
	"github.com/aaronriekenberg/go-api/handlers/requestbody"
	"github.com/aaronriekenberg/go-api/handlers/requestinfo"
	"github.com/aaronriekenberg/go-api/handlers/requestlogging"
	"github.com/aaronriekenberg/go-api/handlers/staticfile"
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

	mux.Handle("GET /", staticfile.NewStaticFileHandler(config.StaticFileConfiguration))

	handler := requestbody.EmptyRequestBodyHandler(mux)

	requestLogger := requestlogging.NewRequestLogger(config.RequestLoggingConfiguration)
	return requestLogger.WrapHttpHandler(handler)
}
