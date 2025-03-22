package health

import (
	"net/http"

	"github.com/aaronriekenberg/go-api/utils"
)

const (
	responseBodyString = "all good"
)

func NewHealthHandler() http.Handler {
	return utils.PlainTextHandlerFunc(responseBodyString)
}
