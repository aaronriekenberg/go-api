package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/BurntSushi/toml"
)

type ServerListenerConfiguration struct {
	Network       string
	ListenAddress string
	H2CEnabled    bool
}

type ServerConfiguration struct {
	Listeners  []ServerListenerConfiguration
	APIContext string
}

type ProfilingConfiguration struct {
	Enabled       bool
	ListenAddress string
}

type RequestLoggingConfiguration struct {
	Enabled          bool
	RequestLogFile   string
	MaxSizeMegabytes int
	MaxBackups       int
}

type CommandInfo struct {
	ID          string
	Description string
	Command     string
	Args        []string
}

type StaticFileConfiguration struct {
	RootPath string
}

type CommandConfiguration struct {
	MaxConcurrentCommands           int64
	RequestTimeoutDuration          time.Duration
	SemaphoreAcquireTimeoutDuration time.Duration
	Commands                        []CommandInfo
}

// Idea from https://choly.ca/post/go-json-marshalling/
func (c *CommandConfiguration) MarshalJSON() ([]byte, error) {
	type Alias CommandConfiguration
	return json.Marshal(&struct {
		RequestTimeoutDuration          string
		SemaphoreAcquireTimeoutDuration string
		*Alias
	}{
		RequestTimeoutDuration:          c.RequestTimeoutDuration.String(),
		SemaphoreAcquireTimeoutDuration: c.SemaphoreAcquireTimeoutDuration.String(),
		Alias:                           (*Alias)(c),
	})
}

type Configuration struct {
	GoMaxProcs                  int
	ServerConfiguration         ServerConfiguration
	ProfilingConfiguration      ProfilingConfiguration
	RequestLoggingConfiguration RequestLoggingConfiguration
	StaticFileConfiguration     StaticFileConfiguration
	CommandConfiguration        CommandConfiguration
}

func ReadConfiguration(configFile string) (*Configuration, error) {

	logger := slog.Default().With("configFile", configFile)

	logger.Info("begin ReadConfiguration")

	var configuration Configuration
	_, err := toml.DecodeFile(configFile, &configuration)
	if err != nil {
		logger.Error("toml.DecodeFile error",
			"error", err,
		)
		return nil, fmt.Errorf("ReadConfiguration error: %w", err)
	}

	logger.Info("end ReadConfiguration",
		"configuration", &configuration,
	)

	return &configuration, nil
}
