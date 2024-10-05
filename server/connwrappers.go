package server

import (
	"io"
	"log/slog"
	"net"

	"github.com/aaronriekenberg/go-api/connection"
)

type connectionInfoWrapper interface {
	connectionInfo() connection.ConnectionInfo
}

type tcpConnWrapper struct {
	*net.TCPConn
	connInfo connection.ConnectionInfo
}

var _ io.ReaderFrom = (*tcpConnWrapper)(nil)
var _ io.WriterTo = (*tcpConnWrapper)(nil)
var _ connectionInfoWrapper = (*tcpConnWrapper)(nil)

func newTCPConnWrapper(
	conn *net.TCPConn,
) *tcpConnWrapper {
	connInfo := connection.ConnectionManagerInstance().AddConnection("tcp")

	slog.Debug("newTCPConnWrapper",
		"connectionID", connInfo.ID(),
	)

	return &tcpConnWrapper{
		TCPConn:  conn,
		connInfo: connInfo,
	}
}

func (tcw *tcpConnWrapper) Close() error {
	slog.Debug("tcpConnWrapper.Close",
		"connectionID", tcw.connInfo.ID(),
	)

	connection.ConnectionManagerInstance().RemoveConnection(
		tcw.connInfo.ID(),
	)

	return tcw.TCPConn.Close()
}

func (tcw *tcpConnWrapper) connectionInfo() connection.ConnectionInfo {
	return tcw.connInfo
}

type unixConnWrapper struct {
	*net.UnixConn
	connInfo connection.ConnectionInfo
}

var _ connectionInfoWrapper = (*unixConnWrapper)(nil)

func newUnixConnWrapper(
	conn *net.UnixConn,
) *unixConnWrapper {
	connInfo := connection.ConnectionManagerInstance().AddConnection("unix")

	slog.Debug("newUnixConnWrapper",
		"connectionID", connInfo.ID(),
	)

	return &unixConnWrapper{
		UnixConn: conn,
		connInfo: connInfo,
	}
}

func (ucw *unixConnWrapper) Close() error {
	slog.Debug("unixConnWrapper.Close",
		"connectionID", ucw.connInfo.ID(),
	)

	connection.ConnectionManagerInstance().RemoveConnection(
		ucw.connInfo.ID(),
	)

	return ucw.UnixConn.Close()
}

func (ucw *unixConnWrapper) connectionInfo() connection.ConnectionInfo {
	return ucw.connInfo
}
