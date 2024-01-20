package connection

import (
	"log/slog"
	"sync"
	"time"
)

type ConnectionManagerState struct {
	MaxOpenConnections       int
	MaxConnectionAge         time.Duration
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

func (cm *connectionManager) State() ConnectionManagerState {
	connections := cm.connections()

	connectionMetrics := cm.metricsManager.connectionMetrics()

	maxConnectionAge := connectionMetrics.maxConnectionAge
	maxRequestsPerConnection := connectionMetrics.maxRequestsPerConnection

	now := time.Now()

	for _, c := range connections {
		maxConnectionAge = max(c.Age(now), maxConnectionAge)
		maxRequestsPerConnection = max(c.Requests(), maxRequestsPerConnection)
	}

	return ConnectionManagerState{
		MaxOpenConnections:       connectionMetrics.maxOpenConnections,
		MaxConnectionAge:         maxConnectionAge,
		MaxRequestsPerConnection: maxRequestsPerConnection,
		Connections:              connections,
	}
}

var connectionManagerInstance = newConnectionManager()

func ConnectionManagerInstance() ConnectionManager {
	return connectionManagerInstance
}
