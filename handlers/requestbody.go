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
