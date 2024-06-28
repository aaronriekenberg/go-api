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
	config config.ServerListenerConfiguration,
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
	if connWrapper, ok := c.(*connWrapper); ok {
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

func runListener(
	listenerConfig config.ServerListenerConfiguration,
	serverConfig config.ServerConfiguration,
	handler http.Handler,
	errorChannel chan<- error,
) {
	logger := slog.Default().With(
		"listenerConfig", listenerConfig,
	)

	logger.Info("begin server.runListener")

	listener, err := createListener(
		listenerConfig,
	)
	if err != nil {
		logger.Warn("server.createListener error",
			"error", err,
		)
		errorChannel <- fmt.Errorf("server.createListener error: %w", err)
		return
	}

	handler = updateContextForRequestHandler(handler)

	if serverConfig.H2CEnabled {
		logger.Info("server.runListener enabling h2c")
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

	logger.Warn("httpServer.serve error",
		"error", err,
	)
	errorChannel <- fmt.Errorf("httpServer.Serve error: %w", err)
}

func Run(
	serverConfig config.ServerConfiguration,
	handler http.Handler,
) error {

	slog.Info("begin server.Run")

	if len(serverConfig.Listeners) < 1 {
		return fmt.Errorf("no server configured")
	}

	errorChannel := make(chan error, len(serverConfig.Listeners))

	for _, listenerConfig := range serverConfig.Listeners {
		go runListener(
			listenerConfig,
			serverConfig,
			handler,
			errorChannel,
		)
	}

	err := <-errorChannel
	return fmt.Errorf("server.runListener error: %w", err)
}
