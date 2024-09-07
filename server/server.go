package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"golang.org/x/sys/unix"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/connection"
	"github.com/aaronriekenberg/go-api/request"
)

func createListener(
	config config.ServerListenerConfiguration,
) (net.Listener, error) {
	if config.Network == "unix" {
		os.Remove(config.ListenAddress)
	}

	if config.Network != "tcp" {
		return net.Listen(config.Network, config.ListenAddress)
	}

	listenConfig := net.ListenConfig{
		Control: func(network, address string, c syscall.RawConn) error {
			var opErr error
			if err := c.Control(func(fd uintptr) {
				slog.Info("setting SO_REUSEPORT for socket")
				opErr = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
			}); err != nil {
				slog.Warn("error setting SO_REUSEPORT",
					"error", err,
				)
				return err
			}
			return opErr
		},
	}
	return listenConfig.Listen(context.Background(), config.Network, config.ListenAddress)
}

func createConnectionContext(
	ctx context.Context,
	conn net.Conn,
) context.Context {
	if connectionInfo, added := connection.ConnectionManagerInstance().AddConnection(conn); added {
		return connection.AddConnectionInfoToContext(ctx, connectionInfo)
	}

	return ctx
}

func serverConnStateChanged(
	conn net.Conn,
	connState http.ConnState,
) {
	slog.Debug("serverConnStateChanged",
		"connState", connState.String(),
	)

	switch connState {
	case http.StateClosed:
		connection.ConnectionManagerInstance().RemoveConnection(conn)
	}
}

func updateContextForRequestHandler(
	handler http.Handler,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := request.NextRequestID()

		if connectionInfo, ok := connection.ConnectionInfoFromContext(ctx); ok {
			connectionInfo.IncrementRequests()
		}

		ctx = request.AddRequestIDToContext(ctx, requestID)

		r = r.WithContext(ctx)

		handler.ServeHTTP(w, r)
	}
}

func runListener(
	listenerConfig config.ServerListenerConfiguration,
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

	if listenerConfig.H2CEnabled {
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
		ConnContext:  createConnectionContext,
		ConnState:    serverConnStateChanged,
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
		return fmt.Errorf("no listeners configured")
	}

	errorChannel := make(chan error, len(serverConfig.Listeners))

	for _, listenerConfig := range serverConfig.Listeners {
		go runListener(
			listenerConfig,
			handler,
			errorChannel,
		)
	}

	err := <-errorChannel
	return fmt.Errorf("server.runListener error: %w", err)
}
