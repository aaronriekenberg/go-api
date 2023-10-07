package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/connection"
)

type connWrapper struct {
	connectionID connection.ConnectionID
	net.Conn
}

func (cw *connWrapper) Close() error {
	slog.Debug("connWrapper.Close",
		"connectionID", cw.connectionID,
	)

	connection.ConnectionManagerInstance().RemoveConnection(
		cw.connectionID)

	return cw.Conn.Close()
}

type listenerWrapper struct {
	net.Listener
}

func (lw *listenerWrapper) Accept() (net.Conn, error) {
	conn, err := lw.Listener.Accept()

	if err != nil {
		return conn, err
	}

	connectionID := connection.ConnectionManagerInstance().AddConnection()

	slog.Debug("listenerWrapper.Accept got new connection",
		"connectionID", connectionID,
	)

	return &connWrapper{
		connectionID: connectionID,
		Conn:         conn,
	}, nil
}

func createListener(
	config config.ServerConfiguration,
) (net.Listener, error) {
	if config.Network == "unix" {
		os.Remove(config.ListenAddress)
	}

	listener, err := net.Listen(config.Network, config.ListenAddress)
	if err != nil {
		return nil, fmt.Errorf("net.Listen error: %w", err)
	}

	listenerWrapper := &listenerWrapper{
		Listener: listener,
	}

	return listenerWrapper, nil
}

func incrementRequestsForConnectionHandler(
	handler http.Handler,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connectionID, ok := r.Context().Value(connection.ConnectionIDContextKey).(connection.ConnectionID)
		if ok {
			connection.ConnectionManagerInstance().IncrementRequestsForConnection(connectionID)
		}
		handler.ServeHTTP(w, r)
	}
}

func addConnectionIDToContext(ctx context.Context, c net.Conn) context.Context {
	connWrapper, ok := c.(*connWrapper)
	if ok {
		connectionID := connWrapper.connectionID
		return context.WithValue(ctx, connection.ConnectionIDContextKey, connectionID)
	}
	return ctx
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

	listener, err := createListener(
		config,
	)
	if err != nil {
		logger.Error("createListener error",
			"error", err,
		)
		return fmt.Errorf("createListener error: %w", err)
	}

	handler = incrementRequestsForConnectionHandler(handler)

	h2Server := &http2.Server{
		IdleTimeout: 5 * time.Minute,
	}

	httpServer := &http.Server{
		IdleTimeout:  5 * time.Minute,
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
		ConnContext:  addConnectionIDToContext,
		Handler:      h2c.NewHandler(handler, h2Server),
	}

	err = httpServer.Serve(listener)

	logger.Error("httpServer.Serve error",
		"error", err,
	)
	return fmt.Errorf("httpServer.Serve error: %w", err)
}
