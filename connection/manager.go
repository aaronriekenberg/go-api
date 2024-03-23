package connection

import (
	"cmp"
	"log/slog"
	"slices"
	"sync"
	"time"
)

type ConnectionManagerState struct {
	MaxOpenConnections       int
	MinConnectionLifetime    time.Duration
	MaxConnectionLifetime    time.Duration
	MaxRequestsPerConnection uint64
	Connections              []Connection
}

type ConnectionManager interface {
	AddConnection() ConnectionID

	IncrementRequestsForConnection(connectionID ConnectionID)

	RemoveConnection(connectionID ConnectionID)

	State() ConnectionManagerState
}

type connectionManager struct {
	mutex            sync.RWMutex
	idToConnection   map[ConnectionID]*connection
	nextConnectionID ConnectionID
	metricsManager   *connectionMetricsManager
}

func newConnectionManager() *connectionManager {
	return &connectionManager{
		idToConnection:   make(map[ConnectionID]*connection),
		nextConnectionID: 1,
		metricsManager:   newConnectionMetricsManager(),
	}
}

func (cm *connectionManager) AddConnection() ConnectionID {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	connectionID := cm.nextConnectionID
	cm.nextConnectionID++

	cm.idToConnection[connectionID] = newConnection(connectionID)

	slog.Info("connectionManager.AddConnection",
		"connectionID", connectionID,
	)

	numOpenConnections := len(cm.idToConnection)
	cm.metricsManager.updateForNewConnection(numOpenConnections)

	return connectionID
}

func (cm *connectionManager) IncrementRequestsForConnection(connectionID ConnectionID) {
	connection := cm.connection(connectionID)

	if connection != nil {
		connection.incrementRequests()
	}
}

func (cm *connectionManager) RemoveConnection(connectionID ConnectionID) {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	connection := cm.idToConnection[connectionID]
	if connection != nil {
		slog.Info("connectionManager.RemoveConnection",
			"connectionID", connectionID,
			"requests", connection.Requests(),
		)

		connection.markClosed()

		cm.metricsManager.updateForClosedConnection(connection)
	}

	delete(cm.idToConnection, connectionID)
}

func (cm *connectionManager) connection(connectionID ConnectionID) *connection {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	return cm.idToConnection[connectionID]
}

func (cm *connectionManager) connections() []Connection {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	connections := make([]Connection, 0, len(cm.idToConnection))

	for _, connection := range cm.idToConnection {
		connections = append(connections, connection)
	}

	return connections
}

func computeMinConnectionLifetime(
	now time.Time,
	connections []Connection,
	connectionMetrics connectionMetrics,
) (minConnectionLifetime time.Duration) {
	if connectionMetrics.pastMinConnectionAge != nil {
		minConnectionLifetime = *connectionMetrics.pastMinConnectionAge
	} else if len(connections) > 0 {
		minAgeConnection := slices.MinFunc(connections, func(c1, c2 Connection) int {
			return cmp.Compare(c1.Age(now), c2.Age(now))
		})
		minConnectionLifetime = minAgeConnection.Age(now)
	}
	return
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
