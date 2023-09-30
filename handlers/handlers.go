package handlers

import (
	"net/http"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/handlers/command"
	"github.com/aaronriekenberg/go-api/handlers/requestinfo"
	"github.com/aaronriekenberg/go-api/handlers/versioninfo"
	"github.com/julienschmidt/httprouter"
)

func CreateHandlers(
	config config.Configuration,
) (http.Handler, error) {
	router := httprouter.New()

	allCommandsHandler, err := command.NewAllCommandsHandler(config.CommandConfiguration)
	if err != nil {
		return nil, err
	}
	router.Handler(http.MethodGet, "/api/v1/commands", allCommandsHandler)

	runCommandsHandler, err := command.NewRunCommandsHandler(config.CommandConfiguration)
	if err != nil {
		return nil, err
	}
	router.Handler(http.MethodGet, "/api/v1/commands/:id", runCommandsHandler)

	router.Handler(http.MethodGet, "/api/v1/request_info", requestinfo.NewRequestInfoHandler())

	versionInfoHandler, err := versioninfo.NewVersionInfoHandler()
	if err != nil {
		return nil, err
	}
	router.Handler(http.MethodGet, "/api/v1/version_info", versionInfoHandler)

	return router, nil
}
