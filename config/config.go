package config

import (
	"log/slog"
	"os"

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

func ReadConfiguration(configFile string) *Configuration {

	logger := slog.Default().With("configFile", configFile)

	logger.Info("begin ReadConfiguration")

	var config Configuration
	_, err := toml.DecodeFile(configFile, &config)
	if err != nil {
		logger.Error("toml.DecodeFile error",
			"error", err)
		os.Exit(1)
	}

	logger.Info("end ReadConfiguration")

	return &config
}
