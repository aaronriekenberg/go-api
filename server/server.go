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
	"github.com/aaronriekenberg/go-api/request"
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

func addConnectionIDToContext(ctx context.Context, c net.Conn) context.Context {
	connWrapper, ok := c.(*connWrapper)
	if ok {
		connectionID := connWrapper.connectionID
		return connection.AddConnectionIDToContext(ctx, connectionID)
	}
	return ctx
}

func updateContextForRequestHandler(
	handler http.Handler,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := request.NextRequestID()

		if connectionID, ok := connection.ConnectionIDFromContext(ctx); ok {
			connection.ConnectionManagerInstance().IncrementRequestsForConnection(connectionID)
		}

		ctx = request.AddRequestIDToContext(ctx, requestID)

		r = r.WithContext(ctx)

		handler.ServeHTTP(w, r)
	}
}

func Run(
	config config.ServerConfiguration,
	handler http.Handler,
) error {
	logger := slog.Default().With(
		"config", config,
	)

	logger.Info("begin server.Run")

	listener, err := createListener(
		config,
	)
	if err != nil {
		return fmt.Errorf("server.Run: createListener error: %w", err)
	}

	handler = updateContextForRequestHandler(handler)

	if config.H2CEnabled {
		h2Server := &http2.Server{
			IdleTimeout: 5 * time.Minute,
		}
		handler = h2c.NewHandler(handler, h2Server)
	}

	httpServer := &http.Server{
		IdleTimeout:  5 * time.Minute,
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
		ConnContext:  addConnectionIDToContext,
		Handler:      handler,
	}

	err = httpServer.Serve(listener)

	return fmt.Errorf("server.Run: httpServer.Serve error: %w", err)
}
