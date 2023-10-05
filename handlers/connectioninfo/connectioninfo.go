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
	Age          string `json:"age"`
	CreationTime string `json:"creation_time"`
	Requests     uint64 `json:"requests"`
}

type connectionInfoResponse struct {
	NumCurrentConnections    int              `json:"num_current_connections"`
	MaxConnectionAge         string           `json:"max_connection_age"`
	MaxRequestsPerConnection uint64           `json:"max_requests_per_connection"`
	Connections              []*connectionDTO `json:"connections"`
}

func connectionInfoHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		connectionManagerState := connection.ConnectionManagerInstance().State()

		connectionDTOs := make([]*connectionDTO, 0, len(connectionManagerState.Connections))

		now := time.Now()

		for _, connection := range connectionManagerState.Connections {
			cdto := &connectionDTO{
				ID:           uint64(connection.ID()),
				Age:          connection.Age(now).Truncate(time.Millisecond).String(),
				CreationTime: utils.FormatTime(connection.CreationTime()),
				Requests:     connection.Requests(),
			}

			connectionDTOs = append(connectionDTOs, cdto)
		}

		slices.SortFunc(connectionDTOs, func(cdto1, cdto2 *connectionDTO) int {
			// sort descending
			return -cmp.Compare(cdto1.ID, cdto2.ID)
		})

		response := &connectionInfoResponse{
			NumCurrentConnections:    len(connectionDTOs),
			MaxConnectionAge:         connectionManagerState.MaxConnectionAge.Truncate(time.Millisecond).String(),
			MaxRequestsPerConnection: connectionManagerState.MaxRequestsPerConnection,
			Connections:              connectionDTOs,
		}

		utils.RespondWithJSONDTO(response, w)
	}
}

func NewConnectionInfoHandler() http.Handler {
	return connectionInfoHandlerFunc()
}
