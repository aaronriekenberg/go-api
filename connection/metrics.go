package connection

import (
	"sync/atomic"
	"time"
)

type connectionMetrics struct {
	maxConnectionAge         time.Duration
	maxRequestsPerConnection uint64
}

func (cm connectionMetrics) updateForClosedConnection(
	closedConnection Connection,
) connectionMetrics {
	return connectionMetrics{
		maxConnectionAge:         max(cm.maxConnectionAge, closedConnection.Age(time.Now())),
		maxRequestsPerConnection: max(cm.maxRequestsPerConnection, closedConnection.Requests()),
	}
}

type connectionMetricsManager struct {
	atomicConnectionMetrics          atomic.Pointer[connectionMetrics]
	updateForClosedConnectionChannel chan Connection
}

func newConnectionMetricsManager() *connectionMetricsManager {
	cmm := &connectionMetricsManager{
		updateForClosedConnectionChannel: make(chan Connection),
	}

	cmm.atomicConnectionMetrics.Store(&connectionMetrics{})

	go cmm.runUpdateMetricsTask()

	return cmm
}

func (cmm *connectionMetricsManager) connectionMetrics() connectionMetrics {
	return *cmm.atomicConnectionMetrics.Load()
}

func (cmm *connectionMetricsManager) runUpdateMetricsTask() {
	for {
		closedConnection := <-cmm.updateForClosedConnectionChannel

		currentMetrics := cmm.connectionMetrics()

		newMetrics := currentMetrics.updateForClosedConnection(closedConnection)

		cmm.atomicConnectionMetrics.Store(&newMetrics)
	}
}

func (cmm *connectionMetricsManager) updateForClosedConnection(
	closedConnection Connection,
) {
	go func() {
		cmm.updateForClosedConnectionChannel <- closedConnection
	}()
}
