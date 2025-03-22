package health

import (
	"net/http"

	"github.com/aaronriekenberg/go-api/utils"
)

const (
	responseBodyString = "all good"
)

func healthHandlerFunc() http.HandlerFunc {
	return utils.PlainTextHandlerFunc(responseBodyString)
}

func NewHealthHandler() http.Handler {
	return healthHandlerFunc()
}
