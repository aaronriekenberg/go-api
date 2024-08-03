package connection

import (
	"fmt"
	"log/slog"
	"net"
)

type connKey struct {
	tcpConn  *net.TCPConn
	unixConn *net.UnixConn
}

func buildConnKeyAndNetwork(
	conn net.Conn,
) (key connKey, network string, ok bool) {
	switch conn := conn.(type) {
	case *net.TCPConn:
		ok = true
		network = "tcp"
		key = connKey{
			tcpConn: conn,
		}
	case *net.UnixConn:
		ok = true
		network = "unix"
		key = connKey{
			unixConn: conn,
		}
	default:
		slog.Warn("buildConnKeyAndNetwork unknown conn type",
			"conn", fmt.Sprintf("%T %v", conn, conn),
		)
		ok = false
	}
	return
}
