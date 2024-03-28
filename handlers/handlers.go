package handlers

import (
	"log/slog"
	"net/http"
	"path"

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

	mux.Handle("GET "+path.Join(context, "/commands"), command.NewAllCommandsHandler(config.CommandConfiguration))

	mux.Handle("GET "+path.Join(context, "/commands/{id}"), command.NewRunCommandsHandler(config.CommandConfiguration))

	mux.Handle("GET "+path.Join(context, "/connection_info"), connectioninfo.NewConnectionInfoHandler())

	mux.Handle("GET "+path.Join(context, "/request_info"), requestinfo.NewRequestInfoHandler())

	mux.Handle("GET "+path.Join(context+"/version_info"), versioninfo.NewVersionInfoHandler())

	return mux
}
