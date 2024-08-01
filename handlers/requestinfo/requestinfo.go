package requestinfo

import (
	"net/http"
	"strings"

	"github.com/aaronriekenberg/go-api/connection"
	"github.com/aaronriekenberg/go-api/request"
	"github.com/aaronriekenberg/go-api/utils"
)

type requestFieldsDTO struct {
	ConnectionID  connection.ConnectionID `json:"connection_id"`
	RequestID     request.RequestID       `json:"request_id"`
	Close         bool                    `json:"close"`
	ContentLength int64                   `json:"content_length"`
	Host          string                  `json:"host"`
	Method        string                  `json:"method"`
	Protocol      string                  `json:"protocol"`
	RemoteAddress string                  `json:"remote_address"`
	URL           string                  `json:"url"`
}

type requestInfoDTO struct {
	RequestFields  requestFieldsDTO  `json:"request_fields"`
	RequestHeaders map[string]string `json:"request_headers"`
}

func httpHeaderToRequestHeaders(headers http.Header) map[string]string {

	requestHeaders := make(map[string]string, len(headers))

	for key, value := range headers {
		requestHeaders[key] = strings.Join(value, "; ")
	}

	return requestHeaders
}

func requestInfoHandlerFunc() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()

		var urlString string
		if r.URL != nil {
			urlString = r.URL.String()
		} else {
			urlString = "(nil)"
		}

		response := requestInfoDTO{
			RequestFields: requestFieldsDTO{
				ConnectionID:  connection.ConnectionIDFromContext(ctx),
				RequestID:     request.RequestIDFromContext(ctx),
				Close:         r.Close,
				ContentLength: r.ContentLength,
				Host:          r.Host,
				Method:        r.Method,
				Protocol:      r.Proto,
				RemoteAddress: r.RemoteAddr,
				URL:           urlString,
			},
			RequestHeaders: httpHeaderToRequestHeaders(r.Header),
		}

		utils.RespondWithJSONDTO(&response, w)
	}
}

func NewRequestInfoHandler() http.Handler {
	return requestInfoHandlerFunc()
}
