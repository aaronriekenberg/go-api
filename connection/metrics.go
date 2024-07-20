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
	closedConnection *connectionInfo
}

type connectionMetricsManager struct {
	atomicConnectionMetrics          atomic.Pointer[connectionMetrics]
	updateForNewConnectionChannel    chan newConnectionMessage
	updateForClosedConnectionChannel chan closedConnectionMessage
}

func newConnectionMetricsManager() *connectionMetricsManager {
	cmm := &connectionMetricsManager{
		updateForNewConnectionChannel:    make(chan newConnectionMessage, maxConnections),
		updateForClosedConnectionChannel: make(chan closedConnectionMessage, maxConnections),
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
			metricsCopy := cmm.connectionMetrics()

			metricsCopy.maxOpenConnections = max(metricsCopy.maxOpenConnections, newConnectionMessage.currentOpenConnections)

			cmm.atomicConnectionMetrics.Store(&metricsCopy)

		case closedConnectionMessage := <-cmm.updateForClosedConnectionChannel:
			metricsCopy := cmm.connectionMetrics()

			closedConnection := closedConnectionMessage.closedConnection

			if metricsCopy.pastMinConnectionAge == nil {
				metricsCopy.pastMinConnectionAge = new(time.Duration)
				*metricsCopy.pastMinConnectionAge = closedConnection.openDuration()
			} else {
				*metricsCopy.pastMinConnectionAge = min(closedConnection.openDuration(), *metricsCopy.pastMinConnectionAge)
			}

			metricsCopy.pastMaxConnectionAge = max(closedConnection.openDuration(), metricsCopy.pastMaxConnectionAge)
			metricsCopy.pastMaxRequestsPerConnection = max(closedConnection.Requests(), metricsCopy.pastMaxRequestsPerConnection)

			cmm.atomicConnectionMetrics.Store(&metricsCopy)
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
	closedConnection *connectionInfo,
) {
	cmm.updateForClosedConnectionChannel <- closedConnectionMessage{
		closedConnection: closedConnection,
	}
}
