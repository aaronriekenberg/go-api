package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"path"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/handlers/command"
	"github.com/aaronriekenberg/go-api/handlers/connectioninfo"
	"github.com/aaronriekenberg/go-api/handlers/requestinfo"
	"github.com/aaronriekenberg/go-api/handlers/staticfile"
	"github.com/aaronriekenberg/go-api/handlers/versioninfo"
	"github.com/aaronriekenberg/go-api/utils"
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

	return maxBodyLengthHandler(mux)
}

func maxBodyLengthHandler(
	nextHandler http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r2 := *r

		r2.Body = http.MaxBytesReader(w, r.Body, 0)

		_, err := io.ReadAll(r2.Body)
		if err != nil {
			slog.Warn("request body read error",
				"error", err,
				"content-length", r2.Header.Get("Content-Length"),
			)
			utils.HTTPErrorStatusCode(w, http.StatusBadRequest)
			return
		}

		nextHandler.ServeHTTP(w, &r2)
	})
}
