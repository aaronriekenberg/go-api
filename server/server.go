package server

import (
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"sync/atomic"
	"time"

	"github.com/aaronriekenberg/go-api/config"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type connWrapper struct {
	connectionID uint64
	net.Conn
}

func (cw *connWrapper) Close() error {
	slog.Info("connWrapper.Close",
		"connectionID", cw.connectionID,
	)

	return cw.Conn.Close()
}

type listenerWrapper struct {
	net.Listener
	lastConnID atomic.Uint64
}

func (lw *listenerWrapper) Accept() (net.Conn, error) {
	conn, err := lw.Listener.Accept()

	if err != nil {
		return conn, err
	}

	connectionID := lw.lastConnID.Add(1)

	slog.Info("listenerWrapper.Accept got new connection",
		"connectionID", connectionID,
	)

	return &connWrapper{
		connectionID: connectionID,
		Conn:         conn,
	}, nil
}

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

	h2Server := &http2.Server{
		IdleTimeout: 2 * time.Minute,
	}

	httpServer := &http.Server{
		IdleTimeout:  2 * time.Minute,
		ReadTimeout:  2 * time.Minute,
		WriteTimeout: 2 * time.Minute,
		Handler:      h2c.NewHandler(handler, h2Server),
	}

	listenerWrapper := &listenerWrapper{
		Listener: listener,
	}

	err = httpServer.Serve(listenerWrapper)

	logger.Error("httpServer.Serve error",
		"error", err,
	)
	return fmt.Errorf("httpServer.Serve error: %w", err)
}
