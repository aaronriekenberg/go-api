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

	return maxRequestBodyLengthHandler(mux)
}

func maxRequestBodyLengthHandler(
	nextHandler http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		bodyReader := http.MaxBytesReader(w, r.Body, 0)

		_, err := io.ReadAll(bodyReader)
		if err != nil {
			slog.Warn("request body read error",
				"error", err,
				"url", r.URL.String(),
				"method", r.Method,
				"proto", r.Proto,
				"header", r.Header,
				"remote_addr", r.RemoteAddr,
			)
			utils.HTTPErrorStatusCode(w, http.StatusBadRequest)
			return
		}

		r2 := *r

		r2.Body = noopReader{}

		nextHandler.ServeHTTP(w, &r2)
	})
}

type noopReader struct{}

func (noopReader noopReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (noopReader noopReader) Close() error {
	return nil
}
