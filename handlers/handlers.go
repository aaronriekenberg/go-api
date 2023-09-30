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

	allCommandsHandler, err := command.NewAllCommandsHandler(config.CommandConfiguration)
	if err != nil {
		return
	}
	router.Handler(get, "/api/v1/commands", allCommandsHandler)

	runCommandsHandler, err := command.NewRunCommandsHandler(config.CommandConfiguration)
	if err != nil {
		return
	}
	router.Handler(get, "/api/v1/commands/:id", runCommandsHandler)

	router.Handler(get, "/api/v1/request_info", requestinfo.NewRequestInfoHandler())

	versionInfoHandler, err := versioninfo.NewVersionInfoHandler()
	if err != nil {
		return
	}
	router.Handler(get, "/api/v1/version_info", versionInfoHandler)

	handler = router
	return
}
