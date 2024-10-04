package server

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/connection"
)

type listenerWrapper struct {
	net.Listener
}

func (lw *listenerWrapper) Accept() (net.Conn, error) {
	conn, err := lw.Listener.Accept()

	if err != nil {
		return conn, err
	}

	switch conn := conn.(type) {
	case *net.TCPConn:
		connectionInfo := connection.ConnectionManagerInstance().AddConnection("tcp")

		slog.Debug("listenerWrapper.Accept got new tcp connection",
			"connectionID", connectionInfo.ID(),
		)

		return newTCPConnWrapper(conn, connectionInfo), nil

	case *net.UnixConn:
		connectionInfo := connection.ConnectionManagerInstance().AddConnection("unix")

		slog.Debug("listenerWrapper.Accept got new unix connection",
			"connectionID", connectionInfo.ID(),
		)

		return newUnixConnWrapper(conn, connectionInfo), nil

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
