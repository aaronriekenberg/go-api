package server

import (
	"io"
	"log/slog"
	"net"

	"github.com/aaronriekenberg/go-api/connection"
)

type connectionInfoWrapper interface {
	GetConnectionInfo() connection.ConnectionInfo
}

type tcpConnWrapper struct {
	*net.TCPConn
	connectionInfo connection.ConnectionInfo
}

var _ io.ReaderFrom = (*tcpConnWrapper)(nil)
var _ io.WriterTo = (*tcpConnWrapper)(nil)

func newTCPConnWrapper(
	conn *net.TCPConn,
) *tcpConnWrapper {
	connectionInfo := connection.ConnectionManagerInstance().AddConnection("tcp")

	slog.Debug("newTCPConnWrapper",
		"connectionID", connectionInfo.ID(),
	)

	return &tcpConnWrapper{
		TCPConn:        conn,
		connectionInfo: connectionInfo,
	}
}

func (tcw *tcpConnWrapper) Close() error {
	slog.Debug("tcpConnWrapper.Close",
		"connectionID", tcw.connectionInfo.ID(),
	)

	connection.ConnectionManagerInstance().RemoveConnection(
		tcw.connectionInfo.ID(),
	)

	return tcw.TCPConn.Close()
}

func (tcw *tcpConnWrapper) GetConnectionInfo() connection.ConnectionInfo {
	return tcw.connectionInfo
}

type unixConnWrapper struct {
	*net.UnixConn
	connectionInfo connection.ConnectionInfo
}

func newUnixConnWrapper(
	conn *net.UnixConn,
) *unixConnWrapper {
	connectionInfo := connection.ConnectionManagerInstance().AddConnection("unix")

	slog.Debug("newUnixConnWrapper",
		"connectionID", connectionInfo.ID(),
	)

	return &unixConnWrapper{
		UnixConn:       conn,
		connectionInfo: connectionInfo,
	}
}

func (ucw *unixConnWrapper) Close() error {
	slog.Debug("unixConnWrapper.Close",
		"connectionID", ucw.connectionInfo.ID(),
	)

	connection.ConnectionManagerInstance().RemoveConnection(
		ucw.connectionInfo.ID(),
	)

	return ucw.UnixConn.Close()
}

func (ucw *unixConnWrapper) GetConnectionInfo() connection.ConnectionInfo {
	return ucw.connectionInfo
}
