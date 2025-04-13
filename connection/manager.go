package connection

import (
	"cmp"
	"log/slog"
	"slices"
	"sync/atomic"
	"time"

	"github.com/aaronriekenberg/gsm"
)

type ConnectionManagerStateSnapshot struct {
	MaxOpenConnections       uint64
	MinConnectionLifetime    time.Duration
	MaxConnectionLifetime    time.Duration
	MaxRequestsPerConnection uint64
	Connections              []ConnectionInfo
}

type ConnectionManager interface {
	AddConnection(network string) ConnectionInfo

	RemoveConnection(connectionID ConnectionID)

	StateSnapshot() ConnectionManagerStateSnapshot
}

type connectionManager struct {
	idToConnection       gsm.GenericSyncMap[ConnectionID, ConnectionInfo]
	numOpenConnections   atomic.Uint64
	previousConnectionID atomic.Uint64
	metricsManager       *connectionMetricsManager
}

func newConnectionManager() *connectionManager {
	return &connectionManager{
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

	numOpenConnections := cm.numOpenConnections.Add(1)

	slog.Debug("connectionManager.AddConnection",
		"connectionID", connectionID,
		"network", network,
		"numOpenConnections", numOpenConnections,
	)

	cm.metricsManager.updateForNewConnection(numOpenConnections)

	return connectionInfo
}

func (cm *connectionManager) RemoveConnection(connectionID ConnectionID) {
	connection, loaded := cm.idToConnection.LoadAndDelete(connectionID)
	if !loaded {
		return
	}

	// Decrement idea from https://pkg.go.dev/sync/atomic@go1.24.0#AddUint64
	numOpenConnections := cm.numOpenConnections.Add(^uint64(0))

	slog.Debug("connectionManager.RemoveConnection",
		"connectionID", connection.ID(),
		"requests", connection.Requests(),
		"numOpenConnections", numOpenConnections,
	)

	connection.markClosed()

	cm.metricsManager.updateForClosedConnection(connection)
}

func computeMinConnectionLifetime(
	now time.Time,
	connections []ConnectionInfo,
	connectionMetrics connectionMetrics,
) time.Duration {
	if connectionMetrics.pastMinConnectionAgeExists {
		return connectionMetrics.pastMinConnectionAge
	}

	if len(connections) > 0 {
		minAgeConnection := slices.MinFunc(connections, func(c1, c2 ConnectionInfo) int {
			return cmp.Compare(c1.Age(now), c2.Age(now))
		})
		return minAgeConnection.Age(now)
	}

	return 0
}

func (cm *connectionManager) StateSnapshot() ConnectionManagerStateSnapshot {
	connectionsSlice := slices.Collect(cm.idToConnection.Values())

	now := time.Now()

	connectionMetrics := cm.metricsManager.connectionMetrics()

	maxConnectionLifetime := connectionMetrics.pastMaxConnectionAge
	maxRequestsPerConnection := connectionMetrics.pastMaxRequestsPerConnection

	for _, c := range connectionsSlice {
		maxConnectionLifetime = max(c.Age(now), maxConnectionLifetime)
		maxRequestsPerConnection = max(c.Requests(), maxRequestsPerConnection)
	}

	return ConnectionManagerStateSnapshot{
		MaxOpenConnections:       connectionMetrics.maxOpenConnections,
		MinConnectionLifetime:    computeMinConnectionLifetime(now, connectionsSlice, connectionMetrics),
		MaxConnectionLifetime:    maxConnectionLifetime,
		MaxRequestsPerConnection: maxRequestsPerConnection,
		Connections:              connectionsSlice,
	}
}

var connectionManagerInstance = newConnectionManager()

func ConnectionManagerInstance() ConnectionManager {
	return connectionManagerInstance
}
