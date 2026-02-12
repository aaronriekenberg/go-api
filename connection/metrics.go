package connection

import (
	"maps"
	"sync/atomic"
	"time"
)

type connectionMetrics struct {
	totalConnections             int
	totalConnectionsByNetwork    map[string]int
	maxOpenConnections           int
	pastMinConnectionAge         *time.Duration
	pastMaxConnectionAge         time.Duration
	pastMaxRequestsPerConnection int
}

func newConnectionMetrics() *connectionMetrics {
	return &connectionMetrics{
		totalConnectionsByNetwork: make(map[string]int),
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
	currentOpenConnections int
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

		if updateMessage.newConnectionMessage != nil {

			newConnectionMessage := updateMessage.newConnectionMessage

			metricsClone := cmm.connectionMetrics()

			metricsClone.totalConnections++

			metricsClone.totalConnectionsByNetwork[newConnectionMessage.newConnection.Network()]++

			metricsClone.maxOpenConnections = max(metricsClone.maxOpenConnections, newConnectionMessage.currentOpenConnections)

			cmm.atomicConnectionMetrics.Store(metricsClone)
		}

		if updateMessage.closedConnectionMessage != nil {

			closedConnectionMessage := updateMessage.closedConnectionMessage

			metricsClone := cmm.connectionMetrics()

			closedConnection := closedConnectionMessage.closedConnection

			if metricsClone.pastMinConnectionAge == nil {
				metricsClone.pastMinConnectionAge = new(closedConnection.openDuration())
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
	currentOpenConnections int,
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
