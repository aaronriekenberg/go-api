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
	Connections              []Connection
}

type ConnectionManager interface {
	AddConnection(network string) ConnectionID

	IncrementRequestsForConnection(connectionID ConnectionID)

	RemoveConnection(connectionID ConnectionID)

	State() ConnectionManagerState
}

type connectionManager struct {
	idToConnection       *xsync.MapOf[ConnectionID, *connection]
	previousConnectionID atomic.Uint64
	metricsManager       *connectionMetricsManager
}

func newConnectionManager() *connectionManager {
	return &connectionManager{
		idToConnection: xsync.NewMapOf[ConnectionID, *connection](
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
) ConnectionID {
	connectionID := cm.nextConnectionID()

	cm.idToConnection.Store(
		connectionID,
		newConnection(connectionID, network),
	)

	slog.Debug("connectionManager.AddConnection",
		"connectionID", connectionID,
		"network", network,
	)

	numOpenConnections := cm.idToConnection.Size()
	cm.metricsManager.updateForNewConnection(numOpenConnections)

	return connectionID
}

func (cm *connectionManager) IncrementRequestsForConnection(connectionID ConnectionID) {
	if connection, loaded := cm.idToConnection.Load(connectionID); loaded {
		connection.incrementRequests()
	}
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

func (cm *connectionManager) connections() []Connection {
	connections := make([]Connection, 0, cm.idToConnection.Size())

	cm.idToConnection.Range(
		func(key ConnectionID, value *connection) bool {
			connections = append(connections, value)
			return true
		},
	)

	return connections
}

func computeMinConnectionLifetime(
	now time.Time,
	connections []Connection,
	connectionMetrics connectionMetrics,
) time.Duration {
	if connectionMetrics.pastMinConnectionAge != nil {
		return *connectionMetrics.pastMinConnectionAge
	}

	if len(connections) > 0 {
		minAgeConnection := slices.MinFunc(connections, func(c1, c2 Connection) int {
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
