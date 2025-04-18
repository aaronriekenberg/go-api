package connection

import (
	"sync/atomic"
	"time"
)

type connectionMetrics struct {
	maxOpenConnections           uint64
	pastMinConnectionAge         *time.Duration
	pastMaxConnectionAge         time.Duration
	pastMaxRequestsPerConnection uint64
}

func (cm *connectionMetrics) clone() *connectionMetrics {
	if cm == nil {
		return new(connectionMetrics)
	}

	cmClone := *cm

	if cm.pastMinConnectionAge != nil {
		cmClone.pastMinConnectionAge = new(time.Duration)
		cmClone.pastMinConnectionAge = cm.pastMinConnectionAge
	}

	return &cmClone
}

type newConnectionMessage struct {
	currentOpenConnections uint64
}

type closedConnectionMessage struct {
	closedConnection ConnectionInfo
}

type updateMetricsMessage struct {
	newConnectionMessage    *newConnectionMessage
	closedConnectionMessage *closedConnectionMessage
}

type connectionMetricsManager struct {
	atomicConnectionMetrics atomic.Pointer[connectionMetrics]
	updateChannel           chan updateMetricsMessage
}

func newConnectionMetricsManager() *connectionMetricsManager {
	cmm := &connectionMetricsManager{
		updateChannel: make(chan updateMetricsMessage, maxConnections),
	}

	go cmm.runUpdateMetricsTask()

	return cmm
}

func (cmm *connectionMetricsManager) connectionMetrics() *connectionMetrics {
	return cmm.atomicConnectionMetrics.Load().clone()
}

func (cmm *connectionMetricsManager) runUpdateMetricsTask() {

	for {
		updateMessage := <-cmm.updateChannel

		newConnectionMessage := updateMessage.newConnectionMessage

		if newConnectionMessage != nil {

			metricsClone := cmm.connectionMetrics()

			metricsClone.maxOpenConnections = max(metricsClone.maxOpenConnections, newConnectionMessage.currentOpenConnections)

			cmm.atomicConnectionMetrics.Store(metricsClone)
		}

		closedConnectionMessage := updateMessage.closedConnectionMessage

		if updateMessage.closedConnectionMessage != nil {

			metricsClone := cmm.connectionMetrics()

			closedConnection := closedConnectionMessage.closedConnection

			if metricsClone.pastMinConnectionAge == nil {
				metricsClone.pastMinConnectionAge = new(time.Duration)
				*metricsClone.pastMinConnectionAge = closedConnection.openDuration()
			} else {
				*metricsClone.pastMinConnectionAge = min(closedConnection.openDuration(), *metricsClone.pastMinConnectionAge)
			}

			metricsClone.pastMaxConnectionAge = max(closedConnection.openDuration(), metricsClone.pastMaxConnectionAge)
			metricsClone.pastMaxRequestsPerConnection = max(closedConnection.Requests(), metricsClone.pastMaxRequestsPerConnection)

			cmm.atomicConnectionMetrics.Store(metricsClone)
		}
	}
}

func (cmm *connectionMetricsManager) updateForNewConnection(
	currentOpenConnections uint64,
) {
	cmm.updateChannel <- updateMetricsMessage{
		newConnectionMessage: &newConnectionMessage{
			currentOpenConnections: currentOpenConnections,
		},
	}

}

func (cmm *connectionMetricsManager) updateForClosedConnection(
	closedConnection ConnectionInfo,
) {
	cmm.updateChannel <- updateMetricsMessage{
		closedConnectionMessage: &closedConnectionMessage{
			closedConnection: closedConnection,
		},
	}
}
