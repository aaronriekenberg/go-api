package requestlogging

import (
	"io"
	"net/http"

	gorillaHandlers "github.com/gorilla/handlers"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/aaronriekenberg/go-api/config"
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

	return gorillaHandlers.CombinedLoggingHandler(requestLogger, handler)
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
