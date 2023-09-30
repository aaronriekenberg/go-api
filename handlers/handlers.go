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
) http.Handler {
	router := httprouter.New()

	router.Handler(http.MethodGet, "/api/v1/commands", command.NewAllCommandsHandler(config.CommandConfiguration))
	router.Handler(http.MethodGet, "/api/v1/commands/:id", command.NewRunCommandsHandler(config.CommandConfiguration))

	router.Handler(http.MethodGet, "/api/v1/request_info", requestinfo.NewRequestInfoHandler())

	router.Handler(http.MethodGet, "/api/v1/version_info", versioninfo.NewVersionInfoHandler())

	return router
}
