package config

import (
	"fmt"
	"log/slog"

	"github.com/BurntSushi/toml"
)

type ServerListenerConfiguration struct {
	Network       string
	ListenAddress string
}

type ServerConfiguration struct {
	Listeners  []ServerListenerConfiguration
	H2CEnabled bool
	Context    string
}

type ProfilingConfiguration struct {
	Enabled       bool
	ListenAddress string
}

type CommandInfo struct {
	ID          string
	Description string
	Command     string
	Args        []string
}

type CommandConfiguration struct {
	MaxConcurrentCommands           int64
	RequestTimeoutDuration          string
	SemaphoreAcquireTimeoutDuration string
	Commands                        []CommandInfo
}

type StaticFileConfiguration struct {
	RootPath string
}

type Configuration struct {
	ServerConfiguration     ServerConfiguration
	ProfilingConfiguration  ProfilingConfiguration
	CommandConfiguration    CommandConfiguration
	StaticFileConfiguration StaticFileConfiguration
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
		"configuration", configuration,
	)

	return &configuration, nil
}
