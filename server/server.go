package server

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/aaronriekenberg/go-api/config"
)

func Run(
	config config.ServerConfiguration,
	handler http.Handler,
) error {
	logger := slog.Default().With(slog.Group("config",
		"Network", config.Network,
		"ListenAddress", config.ListenAddress,
	))

	logger.Info("begin server.Run")

	if config.Network == "unix" {
		os.Remove(config.ListenAddress)
	}

	listener, err := net.Listen(config.Network, config.ListenAddress)
	if err != nil {
		logger.Error("server.Run net.Listen error",
			"error", err,
		)
		return fmt.Errorf("server.Run net.Listen error: %w", err)
	}

	httpServer := &http.Server{
		Handler: handler,
	}

	err = httpServer.Serve(listener)

	logger.Error("server.Run http.ListenAndServe error",
		"error", err,
	)
	return fmt.Errorf("server.Run http.ListenAndServe error: %w", err)
}
