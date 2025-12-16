package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/connection"
	"github.com/aaronriekenberg/go-api/request"
)

func addConnectionInfoToContext(
	ctx context.Context,
	c net.Conn,
) context.Context {
	if connWrapper, ok := c.(connectionInfoWrapper); ok {
		connectionInfo := connWrapper.connectionInfo()
		return connection.AddConnectionInfoToContext(ctx, connectionInfo)
	}
	return ctx
}

var nextRequestID func() request.RequestID = request.RequestIDFactory()

func updateContextForRequestHandler(
	handler http.Handler,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := nextRequestID()

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

	protocols := new(http.Protocols)
	protocols.SetHTTP1(true)

	if listenerConfig.H2CEnabled {
		logger.Info("server.runListener enabling h2c")

		protocols.SetUnencryptedHTTP2(true)
	}

	logger.Info("creating httpServer",
		"protocols", protocols.String(),
	)

	httpServer := &http.Server{
		IdleTimeout:  5 * time.Minute,
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
		ConnContext:  addConnectionInfoToContext,
		Handler:      handler,
		Protocols:    protocols,
	}

	err = httpServer.Serve(listener)

	logger.Warn("httpServer.serve error",
		"error", err,
	)
	errorChannel <- fmt.Errorf("httpServer.Serve error: %w", err)
}

func Run(
	handler http.Handler,
) error {

	serverConfig := config.Instance().ServerConfiguration

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
