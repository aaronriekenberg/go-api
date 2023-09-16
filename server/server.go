package server

import (
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/aaronriekenberg/go-api/config"
)

func Run(
	config config.ServerConfiguration,
	handler http.Handler,
) {
	slog.Info("begin server.Run")

	if config.Network == "unix" {
		os.Remove(config.ListenAddress)
	}

	listener, err := net.Listen(config.Network, config.ListenAddress)
	if err != nil {
		slog.Error("net.Listen error",
			"error", err)
		os.Exit(1)
	}

	httpServer := &http.Server{
		Handler: handler,
	}

	err = httpServer.Serve(listener)

	slog.Error("http.ListenAndServe error",
		"error", err)
	os.Exit(1)
}
