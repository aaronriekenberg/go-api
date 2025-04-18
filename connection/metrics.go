package connection

import (
	"maps"
	"sync/atomic"
	"time"
)

type connectionMetrics struct {
	totalConnections             uint64
	totalConnectionsByNetwork    map[string]uint64
	maxOpenConnections           uint64
	pastMinConnectionAge         *time.Duration
	pastMaxConnectionAge         time.Duration
	pastMaxRequestsPerConnection uint64
}

func newConnectionMetrics() *connectionMetrics {
	return &connectionMetrics{
		totalConnectionsByNetwork: make(map[string]uint64),
	}
}

func (cm *connectionMetrics) clone() *connectionMetrics {
	if cm == nil {
		return newConnectionMetrics()
	}

	cmClone := *cm

	cmClone.totalConnectionsByNetwork = maps.Clone(cm.totalConnectionsByNetwork)

	if cm.pastMinConnectionAge != nil {
		cmClone.pastMinConnectionAge = new(time.Duration)
		cmClone.pastMinConnectionAge = cm.pastMinConnectionAge
	}

	return &cmClone
}

type newConnectionMessage struct {
	newConnection          ConnectionInfo
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

	cmm.atomicConnectionMetrics.Store(newConnectionMetrics())

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

			metricsClone.totalConnections++

			metricsClone.totalConnectionsByNetwork[newConnectionMessage.newConnection.Network()]++

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
	newConnection ConnectionInfo,
	currentOpenConnections uint64,
) {
	cmm.updateChannel <- updateMetricsMessage{
		newConnectionMessage: &newConnectionMessage{
			newConnection:          newConnection,
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
