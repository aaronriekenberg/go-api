package connection

import (
	"sync/atomic"
	"time"
)

type connectionMetrics struct {
	maxOpenConnections       int
	minConnectionAge         *time.Duration
	maxConnectionAge         time.Duration
	maxRequestsPerConnection uint64
}

type newConnectionMessage struct {
	currentOpenConnections int
}

type closedConnectionMessage struct {
	closedConnection *connection
}

type connectionMetricsManager struct {
	atomicConnectionMetrics          atomic.Pointer[connectionMetrics]
	updateForNewConnectionChannel    chan newConnectionMessage
	updateForClosedConnectionChannel chan closedConnectionMessage
}

func newConnectionMetricsManager() *connectionMetricsManager {
	cmm := &connectionMetricsManager{
		updateForNewConnectionChannel:    make(chan newConnectionMessage, 10),
		updateForClosedConnectionChannel: make(chan closedConnectionMessage, 10),
	}

	cmm.atomicConnectionMetrics.Store(new(connectionMetrics))

	go cmm.runUpdateMetricsTask()

	return cmm
}

func (cmm *connectionMetricsManager) connectionMetrics() connectionMetrics {
	return *cmm.atomicConnectionMetrics.Load()
}

func (cmm *connectionMetricsManager) runUpdateMetricsTask() {
	for {
		select {
		case newConnectionMessage := <-cmm.updateForNewConnectionChannel:
			newMetrics := cmm.connectionMetrics()

			newMetrics.maxOpenConnections = max(newMetrics.maxOpenConnections, newConnectionMessage.currentOpenConnections)

			cmm.atomicConnectionMetrics.Store(&newMetrics)

		case closedConnectionMessage := <-cmm.updateForClosedConnectionChannel:
			newMetrics := cmm.connectionMetrics()

			closedConnection := closedConnectionMessage.closedConnection

			if newMetrics.minConnectionAge == nil {
				minConnectionAge := closedConnection.openDuration()
				newMetrics.minConnectionAge = &minConnectionAge
			} else {
				minConnectionAge := min(closedConnection.openDuration(), *newMetrics.minConnectionAge)
				newMetrics.minConnectionAge = &minConnectionAge
			}

			newMetrics.maxConnectionAge = max(closedConnection.openDuration(), newMetrics.maxConnectionAge)
			newMetrics.maxRequestsPerConnection = max(closedConnection.Requests(), newMetrics.maxRequestsPerConnection)

			cmm.atomicConnectionMetrics.Store(&newMetrics)
		}
	}
}

func (cmm *connectionMetricsManager) updateForNewConnection(
	currentOpenConnections int,
) {
	cmm.updateForNewConnectionChannel <- newConnectionMessage{
		currentOpenConnections: currentOpenConnections,
	}
}

func (cmm *connectionMetricsManager) updateForClosedConnection(
	closedConnection *connection,
) {
	cmm.updateForClosedConnectionChannel <- closedConnectionMessage{
		closedConnection: closedConnection,
	}
}
