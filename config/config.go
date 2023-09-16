package config

import (
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"
)

type ServerConfiguration struct {
	ListenAddress string
}

type Configuration struct {
	ServerConfiguration ServerConfiguration
}

func ReadConfiguration(configFile string, logger *slog.Logger) *Configuration {

	logger.Info("begin ReadConfiguration")

	var config Configuration
	_, err := toml.DecodeFile(configFile, &config)
	if err != nil {
		logger.Error("toml.DecodeFile error",
			"configFile", configFile,
			"error", err)
		os.Exit(1)
	}

	logger.Info("end ReadConfiguration")

	return &config
}
