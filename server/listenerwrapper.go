package server

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/aaronriekenberg/go-api/config"
)

type listenerWrapper struct {
	net.Listener
}

func (lw *listenerWrapper) Accept() (net.Conn, error) {
	conn, err := lw.Listener.Accept()

	if err != nil {
		slog.Warn("listenerWrapper.Accept error",
			"error", err,
		)
		return conn, err
	}

	switch conn := conn.(type) {
	case *net.TCPConn:
		return newTCPConnWrapper(conn), nil

	case *net.UnixConn:
		return newUnixConnWrapper(conn), nil

	default:
		slog.Warn("listenerWrapper.Accept got unknown conn type",
			"conn", conn,
		)
		return conn, nil
	}
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
