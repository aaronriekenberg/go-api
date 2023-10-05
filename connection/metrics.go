package connection

import (
	"sync/atomic"
	"time"
)

type connectionMetrics struct {
	maxConnectionAge         time.Duration
	maxRequestsPerConnection uint64
}

func (cm *connectionMetrics) updateForClosedConnection(
	closedConnection Connection,
) connectionMetrics {
	if cm == nil {
		cm = &connectionMetrics{}
	}
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

	go cmm.runUpdateMetricsTask()

	return cmm
}

func (cmm *connectionMetricsManager) connectionMetrics() connectionMetrics {
	currentMetrics := cmm.atomicConnectionMetrics.Load()
	if currentMetrics == nil {
		return connectionMetrics{}
	}
	return *currentMetrics
}

func (cmm *connectionMetricsManager) runUpdateMetricsTask() {
	for {
		closedConnection := <-cmm.updateForClosedConnectionChannel

		currentMetrics := cmm.atomicConnectionMetrics.Load()

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
