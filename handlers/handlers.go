package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/handlers/command"
	"github.com/aaronriekenberg/go-api/handlers/requestinfo"
	"github.com/aaronriekenberg/go-api/handlers/versioninfo"
)

func CreateHandlers(
	config config.Configuration,
) (handler http.Handler, err error) {
	const get = http.MethodGet

	router := httprouter.New()

	h, err := command.NewAllCommandsHandler(config.CommandConfiguration)
	if err != nil {
		return
	}
	router.Handler(get, "/api/v1/commands", h)

	h, err = command.NewRunCommandsHandler(config.CommandConfiguration)
	if err != nil {
		return
	}
	router.Handler(get, "/api/v1/commands/:id", h)

	router.Handler(get, "/api/v1/request_info", requestinfo.NewRequestInfoHandler())

	h, err = versioninfo.NewVersionInfoHandler()
	if err != nil {
		return
	}
	router.Handler(get, "/api/v1/version_info", h)

	handler = router
	return
}
