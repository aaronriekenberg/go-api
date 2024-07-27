package handlers

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/aaronriekenberg/go-api/utils"
)

func maxRequestBodyLengthHandler(
	nextHandler http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		bodyReader := http.MaxBytesReader(w, r.Body, 0)

		if _, err := io.ReadAll(bodyReader); err != nil {
			slog.Warn("request body read error",
				"error", err,
				"url", r.URL.String(),
				"method", r.Method,
				"proto", r.Proto,
				"header", r.Header,
				"remote_addr", r.RemoteAddr,
				"content_length", r.ContentLength,
			)
			utils.HTTPErrorStatusCode(w, http.StatusBadRequest)
			return
		}

		nextHandler.ServeHTTP(w, r)
	})
}
