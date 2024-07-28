package requestlogging

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/connection"
	"github.com/aaronriekenberg/go-api/request"
	"github.com/felixge/httpsnoop"
)

const writeChannelCapacity = 1_000

type RequestLogger interface {
	WrapHttpHandler(handler http.Handler) http.Handler
}

type requestLogger struct {
	writeChannel chan<- []byte
}

func (requestLogger *requestLogger) Write(p []byte) (n int, err error) {
	bufferLength := len(p)
	requestLogger.writeChannel <- p
	return bufferLength, nil
}

func (requestLogger *requestLogger) WrapHttpHandler(
	handler http.Handler,
) http.Handler {
	if requestLogger == nil {
		return handler
	}

	// return gorillaHandlers.CombinedLoggingHandler(requestLogger, handler)
	return newLoggingHandler(requestLogger, handler)
}

type requestLogDTO struct {
	ConnectionID  uint64      `json:"connection_id"`
	RequestID     uint64      `json:"request_id"`
	Close         bool        `json:"close"`
	ContentLength int64       `json:"content_length"`
	Headers       http.Header `json:"headers"`
	Host          string      `json:"host"`
	Method        string      `json:"method"`
	Protocol      string      `json:"protocol"`
	RemoteAddress string      `json:"remote_address"`
	URL           string      `json:"url"`
}

type responseLogDTO struct {
	Headers      http.Header `json:"headers"`
	BytesWritten int64       `json:"bytes_written"`
	Code         int         `json:"code"`
	Duration     string      `json:"duration"`
}

type logDTO struct {
	Timestamp      string         `json:"timestamp"`
	RequestLogDTO  requestLogDTO  `json:"request"`
	ResponseLogDTO responseLogDTO `json:"response"`
}

func newLoggingHandler(
	writer io.Writer,
	nextHandler http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestTime := time.Now()

		ctx := r.Context()

		metrics := httpsnoop.CaptureMetrics(nextHandler, w, r)

		var connectionID connection.ConnectionID
		if connectionInfo, ok := connection.ConnectionInfoFromContext(ctx); ok {
			connectionID = connectionInfo.ID()
		}

		requestID, _ := request.RequestIDFromContext(ctx)

		logDTO := logDTO{
			Timestamp: requestTime.Format("02/Jan/2006:15:04:05.000 -0700"),
			RequestLogDTO: requestLogDTO{
				ConnectionID:  uint64(connectionID),
				RequestID:     uint64(requestID),
				Close:         r.Close,
				ContentLength: r.ContentLength,
				Headers:       r.Header,
				Host:          r.Host,
				Method:        r.Method,
				Protocol:      r.Proto,
				RemoteAddress: r.RemoteAddr,
				URL:           r.URL.String(),
			},
			ResponseLogDTO: responseLogDTO{
				Headers:      w.Header(),
				BytesWritten: metrics.Written,
				Code:         metrics.Code,
				Duration:     metrics.Duration.String(),
			},
		}

		byteBuffer, err := json.Marshal(&logDTO)
		if err != nil {
			slog.Warn("logDTO json.Marshal error",
				"error", err,
			)
		}

		writer.Write(byteBuffer)
	})
}

func NewRequestLogger(
	requestLoggerConfig config.RequestLoggingConfiguration,
) RequestLogger {

	if !requestLoggerConfig.Enabled {
		return (*requestLogger)(nil)
	}

	writer := &lumberjack.Logger{
		Filename:   requestLoggerConfig.RequestLogFile,
		MaxSize:    requestLoggerConfig.MaxSizeMegabytes,
		MaxBackups: requestLoggerConfig.MaxBackups,
	}

	channel := make(chan []byte, writeChannelCapacity)

	requestLogger := &requestLogger{
		writeChannel: channel,
	}

	go runAsyncWriter(
		writer,
		channel,
	)

	return requestLogger
}

func runAsyncWriter(
	writer io.Writer,
	channel <-chan []byte,
) {
	for {
		buffer := <-channel
		writer.Write(buffer)
	}
}
