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

type Configuration struct {
	ServerConfiguration ServerConfiguration
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
