package config

import (
	"fmt"
	"log/slog"

	"github.com/BurntSushi/toml"
)

type ServerConfiguration struct {
	Network       string
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

type Configuration struct {
	ServerConfiguration  ServerConfiguration
	CommandConfiguration CommandConfiguration
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
