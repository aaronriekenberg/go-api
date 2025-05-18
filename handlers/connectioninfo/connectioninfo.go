package connectioninfo

import (
	"cmp"
	"net/http"
	"slices"
	"time"

	"github.com/aaronriekenberg/go-api/connection"
	"github.com/aaronriekenberg/go-api/utils"
)

type connectionDTO struct {
	ID           uint64    `json:"id"`
	Network      string    `json:"network"`
	Age          string    `json:"age"`
	CreationTime time.Time `json:"creation_time"`
	Requests     uint64    `json:"requests"`
}

func connectionInfoToDTO(
	connectionInfo connection.ConnectionInfo,
	now time.Time,
) connectionDTO {
	return connectionDTO{
		ID:           uint64(connectionInfo.ID()),
		Network:      connectionInfo.Network(),
		Age:          connectionInfo.Age(now).Truncate(time.Millisecond).String(),
		CreationTime: connectionInfo.CreationTime(),
		Requests:     connectionInfo.Requests(),
	}
}

type connectionCountsDTO struct {
	Total     uint64            `json:"total"`
	ByNetwork map[string]uint64 `json:"by_network"`
}

type connectionInfoDTO struct {
	MaxOpenConnections       uint64              `json:"max_open_connections"`
	MinConnectionLifetime    string              `json:"min_connection_lifetime"`
	MaxConnectionLifetime    string              `json:"max_connection_lifetime"`
	MaxRequestsPerConnection uint64              `json:"max_requests_per_connection"`
	CurrentConnectionCounts  connectionCountsDTO `json:"current_connection_counts"`
	TotalConnectionCounts    connectionCountsDTO `json:"total_connection_counts"`
	CurrentConnections       []connectionDTO     `json:"current_connections"`
}

func connectionInfoHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		connectionManagerStateSnapshot := connection.ConnectionManagerInstance().StateSnapshot()

		connectionDTOs := make([]connectionDTO, 0, len(connectionManagerStateSnapshot.CurrentConnections))

		now := time.Now()

		for _, connection := range connectionManagerStateSnapshot.CurrentConnections {
			connectionDTO := connectionInfoToDTO(connection, now)
			connectionDTOs = append(connectionDTOs, connectionDTO)
		}

		slices.SortFunc(connectionDTOs, func(cdto1, cdto2 connectionDTO) int {
			// sort descending
			return -cmp.Compare(cdto1.ID, cdto2.ID)
		})

		response := connectionInfoDTO{
			MaxOpenConnections:       connectionManagerStateSnapshot.MaxOpenConnections,
			MinConnectionLifetime:    connectionManagerStateSnapshot.MinConnectionLifetime.Truncate(time.Millisecond).String(),
			MaxConnectionLifetime:    connectionManagerStateSnapshot.MaxConnectionLifetime.Truncate(time.Millisecond).String(),
			MaxRequestsPerConnection: connectionManagerStateSnapshot.MaxRequestsPerConnection,
			CurrentConnectionCounts: connectionCountsDTO{
				Total:     uint64(len(connectionDTOs)),
				ByNetwork: connectionManagerStateSnapshot.CurrentConnectionsByNetwork,
			},
			TotalConnectionCounts: connectionCountsDTO{
				Total:     connectionManagerStateSnapshot.TotalConnections,
				ByNetwork: connectionManagerStateSnapshot.TotalConnectionsByNetwork,
			},
			CurrentConnections: connectionDTOs,
		}

		utils.RespondWithJSONDTO(&response, w)
	}
}

func NewConnectionInfoHandler() http.Handler {
	return connectionInfoHandlerFunc()
}
