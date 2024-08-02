package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	CacheControlHeaderKey = "Cache-Control"
	CacheControlNoCache   = "public, no-cache"

	ContentTypeHeaderKey       = "Content-Type"
	ContentTypeApplicationJSON = "application/json"
)

func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

func RespondWithJSONDTO(
	dto any,
	w http.ResponseWriter,
) {
	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		slog.Warn("utils.RespondWithJSONDTO: json.Marshal error",
			"error", err,
		)
		HTTPErrorStatusCode(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set(CacheControlHeaderKey, CacheControlNoCache)
	w.Header().Set(ContentTypeHeaderKey, ContentTypeApplicationJSON)
	io.Copy(w, bytes.NewReader(jsonBytes))
}

func JSONBytesHandlerFunc(
	jsonBytes []byte,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(CacheControlHeaderKey, CacheControlNoCache)
		w.Header().Set(ContentTypeHeaderKey, ContentTypeApplicationJSON)
		io.Copy(w, bytes.NewReader(jsonBytes))
	}
}

func HTTPErrorStatusCode(
	w http.ResponseWriter,
	statusCode int,
) {
	http.Error(w, http.StatusText(statusCode), statusCode)
}
