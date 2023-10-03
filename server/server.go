package server

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/aaronriekenberg/go-api/config"
	"golang.org/x/net/http2"
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

	http2Server := &http2.Server{}

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("listener.Accept error",
				"error", err,
			)
			return fmt.Errorf("listener.Accept error: %w", err)
		}

		go runConnectionHandler(logger, conn, handler, http2Server)
	}
}

func runConnectionHandler(
	logger *slog.Logger,
	conn net.Conn,
	handler http.Handler,
	http2Server *http2.Server,
) {
	defer conn.Close()

	logger.Info("begin h2cserver.runConnectionHandler")

	http2Server.ServeConn(
		conn,
		&http2.ServeConnOpts{
			Handler: handler,
		},
	)

	logger.Info("end h2cserver.runConnectionHandler")
}
