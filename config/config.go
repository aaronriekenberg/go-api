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

	slog.Info("begin ReadConfiguration")

	var config Configuration
	_, err := toml.DecodeFile(configFile, &config)
	if err != nil {
		slog.Error("toml.DecodeFile error",
			"configFile", configFile,
			"error", err)
		os.Exit(1)
	}

	slog.Info("end ReadConfiguration")

	return &config
}
