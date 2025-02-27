package connection

import (
	"cmp"
	"iter"
	"log/slog"
	"slices"
	"sync/atomic"
	"time"

	"github.com/aaronriekenberg/go-api/utils"
)

type ConnectionManagerStateSnapshot struct {
	MaxOpenConnections       int32
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
	idToConnection       utils.GenericSyncMap[ConnectionID, *connectionInfo]
	numOpenConnections   atomic.Int32
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

	numOpenConnections := cm.numOpenConnections.Add(-1)

	slog.Debug("connectionManager.RemoveConnection",
		"connectionID", connection.ID(),
		"requests", connection.Requests(),
		"numOpenConnections", numOpenConnections,
	)

	connection.markClosed()

	cm.metricsManager.updateForClosedConnection(connection)
}

func (cm *connectionManager) connectionInfoSeq() iter.Seq[ConnectionInfo] {
	return func(yield func(ConnectionInfo) bool) {
		for v := range cm.idToConnection.ValueRange {
			if !yield(v) {
				return
			}
		}
	}
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

func (cm *connectionManager) StateSnapshot() ConnectionManagerStateSnapshot {
	connectionsSlice := slices.Collect(cm.connectionInfoSeq())

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
