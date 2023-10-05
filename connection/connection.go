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

type ConnectionManager interface {
	AddConnection() ConnectionID

	IncrementRequestsForConnection(connectionID ConnectionID)

	RemoveConnection(connectionID ConnectionID)

	Connections() []Connection
}

type connectionManager struct {
	mutex            sync.RWMutex
	idToConnection   map[ConnectionID]*connection
	nextConnectionID ConnectionID
}

func newConnectionManager() connectionManager {
	return connectionManager{
		idToConnection:   make(map[ConnectionID]*connection),
		nextConnectionID: 1,
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
	}

	delete(cm.idToConnection, connectionID)
}

func (cm *connectionManager) Connections() []Connection {
	cm.mutex.RLock()
	defer cm.mutex.RUnlock()

	connections := make([]Connection, 0, len(cm.idToConnection))

	for _, connection := range cm.idToConnection {
		connections = append(connections, connection)
	}

	return connections
}

var connectionManagerInstance = newConnectionManager()

func ConnectionManagerInstance() ConnectionManager {
	return &connectionManagerInstance
}