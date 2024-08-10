package requestlogging

import (
	"net/http"

	"github.com/aaronriekenberg/go-api/connection"
	"github.com/aaronriekenberg/go-api/request"
)

type requestLogData struct {
	ConnectionID  connection.ConnectionID `json:"connection_id"`
	RequestID     request.RequestID       `json:"request_id"`
	Close         bool                    `json:"close"`
	ContentLength int64                   `json:"content_length"`
	Headers       http.Header             `json:"headers"`
	Host          string                  `json:"host"`
	Method        string                  `json:"method"`
	Protocol      string                  `json:"protocol"`
	RemoteAddress string                  `json:"remote_address"`
	URL           string                  `json:"url"`
}

type responseLogData struct {
	Headers      http.Header `json:"headers"`
	BytesWritten int64       `json:"bytes_written"`
	Code         int         `json:"code"`
}

type logData struct {
	Timestamp       string          `json:"timestamp"`
	RequestLogData  requestLogData  `json:"request"`
	ResponseLogData responseLogData `json:"response"`
	Duration        string          `json:"duration"`
}
