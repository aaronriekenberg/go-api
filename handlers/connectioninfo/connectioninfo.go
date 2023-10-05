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
	NumConnections int              `json:"num_connections"`
	Connections    []*connectionDTO `json:"connections"`
}

func connectionInfoHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		connections := connection.ConnectionManagerInstance().Connections()

		connectionDTOs := make([]*connectionDTO, 0, len(connections))

		for _, connection := range connections {
			cdto := &connectionDTO{
				ID:           uint64(connection.ID()),
				Age:          time.Since(connection.CreationTime()).Truncate(time.Millisecond).String(),
				CreationTime: utils.FormatTime(connection.CreationTime()),
				Requests:     connection.Requests(),
			}

			connectionDTOs = append(connectionDTOs, cdto)
		}

		slices.SortFunc(connectionDTOs, func(cdto1, cdto2 *connectionDTO) int {
			return cmp.Compare(cdto1.ID, cdto2.ID)
		})

		response := &connectionInfoResponse{
			NumConnections: len(connectionDTOs),
			Connections:    connectionDTOs,
		}

		utils.RespondWithJSONDTO(response, w)
	}
}

func NewConnectionInfoHandler() http.Handler {
	return connectionInfoHandlerFunc()
}
