package health

import (
	"bytes"
	"io"
	"net/http"

	"github.com/aaronriekenberg/go-api/utils"
)

const (
	responseBodyString = "all good"
)

func healthHandlerFunc() http.HandlerFunc {
	responseBodyBytes := []byte(responseBodyString)

	return func(w http.ResponseWriter, r *http.Request) {

		w.Header().Set(utils.ContentTypeHeaderKey, utils.ContentTypeTextPlain)

		io.Copy(w, bytes.NewReader(responseBodyBytes))
	}
}

func NewHealthHandler() http.Handler {
	return healthHandlerFunc()
}
