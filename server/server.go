package server

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/aaronriekenberg/go-api/config"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func Run(
	config config.ServerConfiguration,
	handler http.Handler,
) error {
	logger := slog.Default().With(
		slog.Group("config",
			"Network", config.Network,
			"ListenAddress", config.ListenAddress,
		),
	)

	logger.Info("begin server.Run")

	if config.Network == "unix" {
		os.Remove(config.ListenAddress)
	}

	listener, err := net.Listen(config.Network, config.ListenAddress)
	if err != nil {
		logger.Error("net.Listen error",
			"error", err,
		)
		return fmt.Errorf("net.Listen error: %w", err)
	}

	h2Server := &http2.Server{}

	httpServer := &http.Server{
		IdleTimeout:  2 * time.Minute,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
		Handler:      h2c.NewHandler(handler, h2Server),
	}

	err = httpServer.Serve(listener)

	logger.Error("httpServer.Serve error",
		"error", err,
	)
	return fmt.Errorf("httpServer.Serve error: %w", err)
}
