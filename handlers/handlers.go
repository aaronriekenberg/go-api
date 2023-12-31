package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/handlers/command"
	"github.com/aaronriekenberg/go-api/handlers/connectioninfo"
	"github.com/aaronriekenberg/go-api/handlers/requestinfo"
	"github.com/aaronriekenberg/go-api/handlers/versioninfo"
)

func CreateHandlers(
	config config.Configuration,
) http.Handler {
	const get = http.MethodGet

	router := httprouter.New()

	router.Handler(get, "/api/v1/commands", command.NewAllCommandsHandler(config.CommandConfiguration))
	router.Handler(get, "/api/v1/commands/:id", command.NewRunCommandsHandler(config.CommandConfiguration))

	router.Handler(get, "/api/v1/connection_info", connectioninfo.NewConnectionInfoHandler())

	router.Handler(get, "/api/v1/request_info", requestinfo.NewRequestInfoHandler())

	router.Handler(get, "/api/v1/version_info", versioninfo.NewVersionInfoHandler())

	return router
}
