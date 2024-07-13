package handlers

import (
	"log/slog"
	"net/http"
	"path"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/handlers/command"
	"github.com/aaronriekenberg/go-api/handlers/connectioninfo"
	"github.com/aaronriekenberg/go-api/handlers/requestinfo"
	"github.com/aaronriekenberg/go-api/handlers/staticfile"
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

	handleGET := func(
		relativePath string,
		handler http.Handler,
	) {
		mux.Handle("GET "+path.Join(context, relativePath), handler)
	}

	handleGET("/commands", command.NewAllCommandsHandler(config.CommandConfiguration))

	handleGET("/commands/{id}", command.NewRunCommandsHandler(config.CommandConfiguration))

	handleGET("/connection_info", connectioninfo.NewConnectionInfoHandler())

	handleGET("/request_info", requestinfo.NewRequestInfoHandler())

	handleGET("/version_info", versioninfo.NewVersionInfoHandler())

	mux.Handle("GET /", staticfile.StaticFileHandler(config.StaticFileConfiguration))

	return mux
}
