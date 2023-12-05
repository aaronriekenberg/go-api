package requestinfo

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/aaronriekenberg/go-api/connection"
	"github.com/aaronriekenberg/go-api/utils"
)

type requestFields struct {
	ConnectionID  string `json:"connection_id"`
	Close         bool   `json:"close"`
	ContentLength int64  `json:"content_length"`
	Host          string `json:"host"`
	Method        string `json:"method"`
	Protocol      string `json:"protocol"`
	RemoteAddress string `json:"remote_address"`
	URL           string `json:"url"`
}

type requestInfoData struct {
	RequestFields  requestFields     `json:"request_fields"`
	RequestHeaders map[string]string `json:"request_headers"`
}

func httpHeaderToRequestHeaders(headers http.Header) map[string]string {

	requestHeaders := make(map[string]string)

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

		connectionIDString := "(nil)"
		if connectionID := connection.GetConnectionIDFromContext(ctx); connectionID != nil {
			connectionIDString = strconv.FormatUint(uint64(*connectionID), 10)
		}

		response := &requestInfoData{
			RequestFields: requestFields{
				ConnectionID:  connectionIDString,
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

		utils.RespondWithJSONDTO(response, w)
	}
}

func NewRequestInfoHandler() http.Handler {
	return requestInfoHandlerFunc()
}
