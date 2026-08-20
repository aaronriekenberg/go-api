package utils

import (
	"bytes"
	"encoding/json/v2"
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
	opts ...json.Options,
) []byte {
	jsonBytes, err := json.Marshal(dto, opts...)
	if err != nil {
		panic(fmt.Errorf("utils.MustMarshalJSON: json.Marshal error: %w", err))
	}
	return jsonBytes
}

func RespondWithJSONDTO(
	dto any,
	w http.ResponseWriter,
	opts ...json.Options,
) {
	w.Header().Set(ContentTypeHeaderKey, ContentTypeApplicationJSON)

	err := json.MarshalWrite(w, dto, opts...)
	if err != nil {
		slog.Warn("utils.RespondWithJSONDTO: json.MarshalWrite error",
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
