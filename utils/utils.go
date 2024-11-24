package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	ContentTypeHeaderKey       = "Content-Type"
	ContentTypeApplicationJSON = "application/json"
	ContentTypeTextPlain       = "text/plain; charset=utf-8"
)

func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339Nano)
}

func MustMarshalJSON(
	dto any,
) []byte {
	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		panic(fmt.Errorf("utils.MustMarshalJSON: json.Marshal error: %w", err))
	}
	return jsonBytes
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

	w.Header().Set(ContentTypeHeaderKey, ContentTypeApplicationJSON)
	io.Copy(w, bytes.NewReader(jsonBytes))
}

func JSONBytesHandlerFunc(
	jsonBytes []byte,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
