package connection

import (
	"cmp"
	"log/slog"
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
	AddConnection(network string) ConnectionInfo

	RemoveConnection(connectionID ConnectionID)

	State() ConnectionManagerState
}

type connectionManager struct {
	idToConnection       *xsync.MapOf[ConnectionID, *connectionInfo]
	previousConnectionID atomic.Uint64
	metricsManager       *connectionMetricsManager
}

func newConnectionManager() *connectionManager {
	return &connectionManager{
		idToConnection: xsync.NewMapOf[ConnectionID, *connectionInfo](
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
	network string,
) ConnectionInfo {
	connectionID := cm.nextConnectionID()
	connectionInfo := newConnection(connectionID, network)

	cm.idToConnection.Store(
		connectionID,
		connectionInfo,
	)

	slog.Debug("connectionManager.AddConnection",
		"connectionID", connectionID,
		"network", network,
	)

	numOpenConnections := cm.idToConnection.Size()
	cm.metricsManager.updateForNewConnection(numOpenConnections)

	return connectionInfo
}

func (cm *connectionManager) RemoveConnection(connectionID ConnectionID) {
	connection, loaded := cm.idToConnection.LoadAndDelete(connectionID)
	if !loaded {
		return
	}

	slog.Debug("connectionManager.RemoveConnection",
		"connectionID", connectionID,
		"requests", connection.Requests(),
	)

	connection.markClosed()

	cm.metricsManager.updateForClosedConnection(connection)

}

func (cm *connectionManager) connections() []ConnectionInfo {
	connections := make([]ConnectionInfo, 0, cm.idToConnection.Size())

	cm.idToConnection.Range(
		func(key ConnectionID, value *connectionInfo) bool {
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
