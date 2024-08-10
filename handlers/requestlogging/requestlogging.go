package requestlogging

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/felixge/httpsnoop"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/connection"
	"github.com/aaronriekenberg/go-api/request"
)

const writeChannelCapacity = 1_000

func NewRequestLogger(
	requestLoggerConfig config.RequestLoggingConfiguration,
	nextHandler http.Handler,
) http.Handler {

	if !requestLoggerConfig.Enabled {
		return nextHandler
	}

	fileWriter := &lumberjack.Logger{
		Filename:   requestLoggerConfig.RequestLogFile,
		MaxSize:    requestLoggerConfig.MaxSizeMegabytes,
		MaxBackups: requestLoggerConfig.MaxBackups,
	}

	channel := make(chan []byte, writeChannelCapacity)

	channelWriter := &channelWriter{
		writeChannel: channel,
	}

	go runAsyncWriter(
		channel,
		fileWriter,
	)

	go channelWriter.runLogDropMonitor()

	return newLoggingHandler(channelWriter, nextHandler)
}

func runAsyncWriter(
	channel <-chan []byte,
	writer io.Writer,
) {
	for {
		buffer := <-channel
		writer.Write(buffer)
	}
}

type channelWriter struct {
	writeChannel chan<- []byte
	numLogDrops  atomic.Uint64
}

func (channelWriter *channelWriter) Write(p []byte) (n int, err error) {
	bufferLength := len(p)

	select {
	case channelWriter.writeChannel <- p:

	default:
		channelWriter.numLogDrops.Add(1)
	}

	return bufferLength, nil
}

func (channelWriter *channelWriter) runLogDropMonitor() {
	ticker := time.NewTicker(5 * time.Second)

	var previousLogDrops uint64 = 0

	for {
		<-ticker.C

		currentLogDrops := channelWriter.numLogDrops.Load()

		if previousLogDrops != currentLogDrops {
			slog.Warn("log drops increased",
				"previousLogDrops", previousLogDrops,
				"currentLogDrops", currentLogDrops,
			)
			previousLogDrops = currentLogDrops
		} else {
			slog.Debug("no change in log drops",
				"previousLogDrops", previousLogDrops,
				"currentLogDrops", currentLogDrops,
			)
		}
	}

}

func newLoggingHandler(
	writer io.Writer,
	nextHandler http.Handler,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestTime := time.Now()

		ctx := r.Context()

		metrics := httpsnoop.CaptureMetrics(nextHandler, w, r)

		logData := logData{
			Timestamp: requestTime.Format(time.RFC3339Nano),
			RequestLogData: requestLogData{
				ConnectionID:  connection.ConnectionIDFromContext(ctx),
				RequestID:     request.RequestIDFromContext(ctx),
				Close:         r.Close,
				ContentLength: r.ContentLength,
				Headers:       r.Header,
				Host:          r.Host,
				Method:        r.Method,
				Protocol:      r.Proto,
				RemoteAddress: r.RemoteAddr,
				URL:           r.URL.String(),
			},
			ResponseLogData: responseLogData{
				Headers:      w.Header(),
				BytesWritten: metrics.Written,
				Code:         metrics.Code,
			},
			Duration: metrics.Duration.String(),
		}

		byteBuffer, err := json.Marshal(&logData)
		if err != nil {
			slog.Warn("logData json.Marshal error",
				"error", err,
			)
			return
		}
		byteBuffer = append(byteBuffer, '\n')

		writer.Write(byteBuffer)
	})
}
