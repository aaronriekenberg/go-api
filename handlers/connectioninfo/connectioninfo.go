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
	ID           uint64 `json:"id"`
	Network      string `json:"network"`
	Age          string `json:"age"`
	CreationTime string `json:"creation_time"`
	Requests     uint64 `json:"requests"`
}

func connectionToDTO(
	connection connection.Connection,
	now time.Time,
) *connectionDTO {
	return &connectionDTO{
		ID:           uint64(connection.ID()),
		Network:      connection.Network(),
		Age:          connection.Age(now).Truncate(time.Millisecond).String(),
		CreationTime: utils.FormatTime(connection.CreationTime()),
		Requests:     connection.Requests(),
	}
}

type connectionInfoResponse struct {
	NumCurrentConnections    int              `json:"num_current_connections"`
	MaxOpenConnections       int              `json:"max_open_connections"`
	MinConnectionLifetime    string           `json:"min_connection_lifetime"`
	MaxConnectionLifetime    string           `json:"max_connection_lifetime"`
	MaxRequestsPerConnection uint64           `json:"max_requests_per_connection"`
	Connections              []*connectionDTO `json:"connections"`
}

func connectionInfoHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		connectionManagerState := connection.ConnectionManagerInstance().State()

		connectionDTOs := make([]*connectionDTO, 0, len(connectionManagerState.Connections))

		now := time.Now()

		for _, connection := range connectionManagerState.Connections {
			connectionDTOs = append(connectionDTOs, connectionToDTO(connection, now))
		}

		slices.SortFunc(connectionDTOs, func(cdto1, cdto2 *connectionDTO) int {
			// sort descending
			return -cmp.Compare(cdto1.ID, cdto2.ID)
		})

		response := &connectionInfoResponse{
			NumCurrentConnections:    len(connectionDTOs),
			MaxOpenConnections:       connectionManagerState.MaxOpenConnections,
			MinConnectionLifetime:    connectionManagerState.MinConnectionLifetime.Truncate(time.Millisecond).String(),
			MaxConnectionLifetime:    connectionManagerState.MaxConnectionLifetime.Truncate(time.Millisecond).String(),
			MaxRequestsPerConnection: connectionManagerState.MaxRequestsPerConnection,
			Connections:              connectionDTOs,
		}

		utils.RespondWithJSONDTO(response, w)
	}
}

func NewConnectionInfoHandler() http.Handler {
	return connectionInfoHandlerFunc()
}
