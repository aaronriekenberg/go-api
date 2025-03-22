package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

const (
	ContentTypeHeaderKey       = "Content-Type"
	ContentTypeApplicationJSON = "application/json"
	ContentTypeTextPlain       = "text/plain; charset=utf-8"
)

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
	w.Header().Set(ContentTypeHeaderKey, ContentTypeApplicationJSON)

	err := json.NewEncoder(w).Encode(dto)
	if err != nil {
		slog.Warn("utils.RespondWithJSONDTO: json.Encode error",
			"error", err,
		)
		HTTPErrorStatusCode(w, http.StatusInternalServerError)
		return
	}

}

func JSONBytesHandlerFunc(
	jsonBytes []byte,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(ContentTypeHeaderKey, ContentTypeApplicationJSON)
		io.Copy(w, bytes.NewReader(jsonBytes))
	}
}

func PlainTextHandlerFunc(
	textString string,
) http.HandlerFunc {
	textBytes := []byte(textString)

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set(ContentTypeHeaderKey, ContentTypeTextPlain)

		io.Copy(w, bytes.NewReader(textBytes))
	}
}

func HTTPErrorStatusCode(
	w http.ResponseWriter,
	statusCode int,
) {
	http.Error(w, http.StatusText(statusCode), statusCode)
}
