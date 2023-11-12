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
	ContentTypeHeaderKey       = "Content-Type"
	ContentTypeApplicationJSON = "application/json"
)

func FormatTime(t time.Time) string {
	const timeFormat = "Mon Jan 2 15:04:05.000000 -0700 MST 2006"

	return t.Format(timeFormat)
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

	w.Header().Add(ContentTypeHeaderKey, ContentTypeApplicationJSON)
	io.Copy(w, bytes.NewReader(jsonBytes))
}

func JSONBytesHandlerFunc(
	jsonBytes []byte,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(ContentTypeHeaderKey, ContentTypeApplicationJSON)
		io.Copy(w, bytes.NewReader(jsonBytes))
	}
}

func HTTPErrorStatusCode(
	w http.ResponseWriter,
	statusCode int,
) {
	http.Error(w, http.StatusText(statusCode), statusCode)
}
