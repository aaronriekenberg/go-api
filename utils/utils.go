package utils

import (
	"bytes"
	"encoding/json"
	"io"
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
		http.Error(w, err.Error(), http.StatusInternalServerError)
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
