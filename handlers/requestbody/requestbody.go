package requestbody

import (
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/aaronriekenberg/go-api/utils"
)

func EmptyRequestBodyHandler(
	nextHandler http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Body == http.NoBody {
			nextHandler.ServeHTTP(w, r)
			return
		}

		logger := slog.Default().With(
			"handler", "emptyrequestbody",
			"urlPath", r.URL.Path,
			"method", r.Method,
			"content_length", r.ContentLength,
		)

		logger.Debug("reading r.body")
		bodyReader := http.MaxBytesReader(w, r.Body, 0)
		_, err := io.ReadAll(bodyReader)

		if err == nil {
			nextHandler.ServeHTTP(w, r)
			return
		}

		logger.Warn("request body read error",
			"error", err,
			"url", r.URL.String(),
			"proto", r.Proto,
			"header", r.Header,
			"remoteAddr", r.RemoteAddr,
		)

		var maxBytesError *http.MaxBytesError
		switch {
		case errors.As(err, &maxBytesError):
			logger.Debug("got maxBytesError",
				"maxBytesError", maxBytesError,
				"limit", maxBytesError.Limit,
			)
			utils.HTTPErrorStatusCode(w, http.StatusRequestEntityTooLarge)

		default:
			logger.Debug("got other error",
				"err", err,
			)
			utils.HTTPErrorStatusCode(w, http.StatusBadRequest)
		}
	})
}
