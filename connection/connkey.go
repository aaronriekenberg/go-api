package connection

import (
	"fmt"
	"log/slog"
	"net"
)

type connKey struct {
	tcpConn  *net.TCPConn
	unixConn *net.UnixConn
	network  string
}

func newConnKey(
	conn net.Conn,
) (key connKey, ok bool) {
	switch conn := conn.(type) {
	case *net.TCPConn:
		ok = true
		key = connKey{
			tcpConn: conn,
			network: "tcp",
		}
	case *net.UnixConn:
		ok = true
		key = connKey{
			unixConn: conn,
			network:  "unix",
		}
	default:
		slog.Warn("newConnKey unknown conn type",
			"type", fmt.Sprintf("%T", conn),
			"conn", conn,
		)
		ok = false
	}
	return
}
