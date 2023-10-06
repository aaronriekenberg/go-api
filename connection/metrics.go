package connection

import (
	"sync/atomic"
	"time"
)

type connectionMetrics struct {
	maxOpenConnections       int
	maxConnectionAge         time.Duration
	maxRequestsPerConnection uint64
}

type connectionMetricsManager struct {
	atomicConnectionMetrics          atomic.Pointer[connectionMetrics]
	updateForNewConnectionChannel    chan int
	updateForClosedConnectionChannel chan Connection
}

func newConnectionMetricsManager() *connectionMetricsManager {
	cmm := &connectionMetricsManager{
		updateForNewConnectionChannel:    make(chan int),
		updateForClosedConnectionChannel: make(chan Connection),
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
		var newMetrics *connectionMetrics

		select {
		case openConnections := <-cmm.updateForNewConnectionChannel:
			newMetrics = new(connectionMetrics)
			*newMetrics = cmm.connectionMetrics()

			newMetrics.maxOpenConnections = max(newMetrics.maxOpenConnections, openConnections)

			cmm.atomicConnectionMetrics.Store(newMetrics)
			newMetrics = nil
		case closedConnection := <-cmm.updateForClosedConnectionChannel:
			newMetrics = new(connectionMetrics)
			*newMetrics = cmm.connectionMetrics()

			newMetrics.maxConnectionAge = max(closedConnection.Age(time.Now()), newMetrics.maxConnectionAge)
			newMetrics.maxRequestsPerConnection = max(closedConnection.Requests(), newMetrics.maxRequestsPerConnection)

			cmm.atomicConnectionMetrics.Store(newMetrics)
			newMetrics = nil
		}
	}
}

func (cmm *connectionMetricsManager) updateForNewConnection(
	currentOpenConnections int,
) {
	cmm.updateForNewConnectionChannel <- currentOpenConnections
}

func (cmm *connectionMetricsManager) updateForClosedConnection(
	closedConnection Connection,
) {
	cmm.updateForClosedConnectionChannel <- closedConnection
}
