package connection

import (
	"cmp"
	"log/slog"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aaronriekenberg/gsm"
)

var ConnectionManagerInstance = sync.OnceValue(newConnectionManager)

type ConnectionManagerStateSnapshot struct {
	TotalConnections            int
	TotalConnectionsByNetwork   map[string]int
	MaxOpenConnections          int
	MinConnectionLifetime       time.Duration
	MaxConnectionLifetime       time.Duration
	MaxRequestsPerConnection    int
	CurrentConnections          []ConnectionInfo
	CurrentConnectionsByNetwork map[string]int
}

type ConnectionManager interface {
	AddConnection(network string) ConnectionInfo

	RemoveConnection(connectionID ConnectionID)

	StateSnapshot() ConnectionManagerStateSnapshot
}

type connectionManager struct {
	idToConnection     gsm.GenericSyncMap[ConnectionID, ConnectionInfo]
	metricsManager     *connectionMetricsManager
	nextConnectionID   func() ConnectionID
	numOpenConnections atomic.Int64
}

func newConnectionManager() ConnectionManager {
	slog.Info("begin newConnectionManager")
	return &connectionManager{
		metricsManager:   newConnectionMetricsManager(),
		nextConnectionID: connectionIDFactory(),
	}
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

	numOpenConnections := int(cm.numOpenConnections.Add(1))

	slog.Info("connectionManager.AddConnection",
		"connectionID", connectionID,
		"network", network,
		"numOpenConnections", numOpenConnections,
	)

	cm.metricsManager.updateForNewConnection(
		connectionInfo,
		numOpenConnections,
	)

	return connectionInfo
}

func (cm *connectionManager) RemoveConnection(connectionID ConnectionID) {
	connection, loaded := cm.idToConnection.LoadAndDelete(connectionID)
	if !loaded {
		return
	}

	numOpenConnections := cm.numOpenConnections.Add(-1)

	slog.Info("connectionManager.RemoveConnection",
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
	connectionMetrics *connectionMetrics,
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
	connectionsSlice := slices.Collect(cm.idToConnection.Values())

	now := time.Now()

	connectionMetrics := cm.metricsManager.connectionMetrics()

	currentConnectionsByNetwork := make(map[string]int)
	maxConnectionLifetime := connectionMetrics.pastMaxConnectionAge
	maxRequestsPerConnection := connectionMetrics.pastMaxRequestsPerConnection

	for _, c := range connectionsSlice {
		maxConnectionLifetime = max(c.Age(now), maxConnectionLifetime)
		maxRequestsPerConnection = max(c.Requests(), maxRequestsPerConnection)
		currentConnectionsByNetwork[c.Network()]++
	}

	return ConnectionManagerStateSnapshot{
		TotalConnections:            connectionMetrics.totalConnections,
		TotalConnectionsByNetwork:   connectionMetrics.totalConnectionsByNetwork,
		MaxOpenConnections:          connectionMetrics.maxOpenConnections,
		MinConnectionLifetime:       computeMinConnectionLifetime(now, connectionsSlice, connectionMetrics),
		MaxConnectionLifetime:       maxConnectionLifetime,
		MaxRequestsPerConnection:    maxRequestsPerConnection,
		CurrentConnections:          connectionsSlice,
		CurrentConnectionsByNetwork: currentConnectionsByNetwork,
	}
}
