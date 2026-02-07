package config

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
)

var Instance = sync.OnceValue(readConfiguration)

func readConfiguration() *Configuration {

	if len(os.Args) != 2 {
		panic("config file required as command line arument")
	}

	configFile := os.Args[1]

	logger := slog.Default().With("configFile", configFile)

	logger.Info("begin readConfiguration")

	var configuration Configuration
	_, err := toml.DecodeFile(configFile, &configuration)
	if err != nil {
		logger.Error("toml.DecodeFile error",
			"error", err,
		)
		panic(fmt.Errorf("readConfiguration error: %w", err))
	}

	logger.Info("end readConfiguration",
		"configuration", &configuration,
	)

	return &configuration
}
