package connection

import (
	"sync/atomic"
	"time"
)

type connectionMetrics struct {
	maxOpenConnections           int
	pastMinConnectionAge         *time.Duration
	pastMaxConnectionAge         time.Duration
	pastMaxRequestsPerConnection uint64
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

			if newMetrics.pastMinConnectionAge == nil {
				pastMinConnectionAge := closedConnection.openDuration()
				newMetrics.pastMinConnectionAge = &pastMinConnectionAge
			} else {
				pastMinConnectionAge := min(closedConnection.openDuration(), *newMetrics.pastMinConnectionAge)
				newMetrics.pastMinConnectionAge = &pastMinConnectionAge
			}

			newMetrics.pastMaxConnectionAge = max(closedConnection.openDuration(), newMetrics.pastMaxConnectionAge)
			newMetrics.pastMaxRequestsPerConnection = max(closedConnection.Requests(), newMetrics.pastMaxRequestsPerConnection)

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
