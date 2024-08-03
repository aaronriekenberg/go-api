package connection

import (
	"cmp"
	"log/slog"
	"net"
	"slices"
	"sync/atomic"
	"time"

	"github.com/puzpuzpuz/xsync/v3"
)

type ConnectionManagerState struct {
	MaxOpenConnections       int
	MinConnectionLifetime    time.Duration
	MaxConnectionLifetime    time.Duration
	MaxRequestsPerConnection uint64
	Connections              []ConnectionInfo
}

type ConnectionManager interface {
	AddConnection(conn net.Conn) (connectionInfo ConnectionInfo, added bool)

	RemoveConnection(conn net.Conn)

	State() ConnectionManagerState
}

type connectionManager struct {
	idToConnection       *xsync.MapOf[connKey, *connectionInfo]
	previousConnectionID atomic.Uint64
	metricsManager       *connectionMetricsManager
}

func newConnectionManager() *connectionManager {
	return &connectionManager{
		idToConnection: xsync.NewMapOf[connKey, *connectionInfo](
			xsync.WithPresize(maxConnections),
			xsync.WithGrowOnly(),
		),
		metricsManager: newConnectionMetricsManager(),
	}
}

func (cm *connectionManager) nextConnectionID() ConnectionID {
	return ConnectionID(cm.previousConnectionID.Add(1))
}

func (cm *connectionManager) AddConnection(
	conn net.Conn,
) (connectionInfo ConnectionInfo, added bool) {
	connKey, ok := newConnKey(conn)
	if !ok {
		return
	}

	connectionID := cm.nextConnectionID()
	newConnectionInfo := newConnection(connectionID, connKey.network)

	cm.idToConnection.Store(
		connKey,
		newConnectionInfo,
	)

	slog.Debug("connectionManager.AddConnection",
		"connectionID", connectionID,
		"network", newConnectionInfo.Network(),
	)

	numOpenConnections := cm.idToConnection.Size()
	cm.metricsManager.updateForNewConnection(numOpenConnections)

	connectionInfo = newConnectionInfo
	added = true

	return
}

func (cm *connectionManager) RemoveConnection(conn net.Conn) {
	connKey, ok := newConnKey(conn)
	if !ok {
		return
	}

	connection, loaded := cm.idToConnection.LoadAndDelete(connKey)
	if !loaded {
		return
	}

	slog.Debug("connectionManager.RemoveConnection",
		"connectionID", connection.ID(),
		"requests", connection.Requests(),
	)

	connection.markClosed()

	cm.metricsManager.updateForClosedConnection(connection)
}

func (cm *connectionManager) connections() []ConnectionInfo {
	connections := make([]ConnectionInfo, 0, cm.idToConnection.Size())

	cm.idToConnection.Range(
		func(key connKey, value *connectionInfo) bool {
			connections = append(connections, value)
			return true
		},
	)

	return connections
}

func computeMinConnectionLifetime(
	now time.Time,
	connections []ConnectionInfo,
	connectionMetrics connectionMetrics,
) time.Duration {
	if connectionMetrics.pastMinConnectionAge != nil {
		return *connectionMetrics.pastMinConnectionAge
	}

	if len(connections) > 0 {
		minAgeConnection := slices.MinFunc(connections, func(c1, c2 ConnectionInfo) int {
			return cmp.Compare(c1.Age(now), c2.Age(now))
		})
		return minAgeConnection.Age(now)
	}

	return 0
}

func (cm *connectionManager) State() ConnectionManagerState {
	connections := cm.connections()

	now := time.Now()

	connectionMetrics := cm.metricsManager.connectionMetrics()

	maxConnectionLifetime := connectionMetrics.pastMaxConnectionAge
	maxRequestsPerConnection := connectionMetrics.pastMaxRequestsPerConnection

	for _, c := range connections {
		maxConnectionLifetime = max(c.Age(now), maxConnectionLifetime)
		maxRequestsPerConnection = max(c.Requests(), maxRequestsPerConnection)
	}

	return ConnectionManagerState{
		MaxOpenConnections:       connectionMetrics.maxOpenConnections,
		MinConnectionLifetime:    computeMinConnectionLifetime(now, connections, connectionMetrics),
		MaxConnectionLifetime:    maxConnectionLifetime,
		MaxRequestsPerConnection: maxRequestsPerConnection,
		Connections:              connections,
	}
}

var connectionManagerInstance = newConnectionManager()

func ConnectionManagerInstance() ConnectionManager {
	return connectionManagerInstance
}
