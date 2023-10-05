package connection

import (
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

type ConnectionID uint64

type connectionIDContextKey struct {
}

var ConnectionIDContextKey = &connectionIDContextKey{}

type Connection interface {
	ID() ConnectionID
	CreationTime() time.Time
	Requests() uint64
}

type connection struct {
	id           ConnectionID
	creationTime time.Time
	requests     atomic.Uint64
}

func (c *connection) ID() ConnectionID {
	return c.id
}

func (c *connection) CreationTime() time.Time {
	return c.creationTime
}

func (c *connection) Requests() uint64 {
	return c.requests.Load()
}

type ConnectionManagerState struct {
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

func newConnectionManager() connectionManager {
	return connectionManager{
		idToConnection:   make(map[ConnectionID]*connection),
		nextConnectionID: 1,
		metricsManager:   newConnectionMetricsManager(),
	}
}

func (cm *connectionManager) AddConnection() ConnectionID {
	cm.mutex.Lock()
	defer cm.mutex.Unlock()

	id := cm.nextConnectionID
	cm.nextConnectionID++

	cm.idToConnection[id] = &connection{
		id:           id,
		creationTime: time.Now(),
	}

	return id
}

func (cm *connectionManager) IncrementRequestsForConnection(connectionID ConnectionID) {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	connection := cm.idToConnection[connectionID]
	if connection != nil {
		connection.requests.Add(1)
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
		maxConnectionAge = max(now.Sub(c.CreationTime()), maxConnectionAge)
		maxRequestsPerConnection = max(c.Requests(), maxRequestsPerConnection)
	}

	return ConnectionManagerState{
		MaxConnectionAge:         maxConnectionAge,
		MaxRequestsPerConnection: maxRequestsPerConnection,
		Connections:              connections,
	}
}

var connectionManagerInstance = newConnectionManager()

func ConnectionManagerInstance() ConnectionManager {
	return &connectionManagerInstance
}
