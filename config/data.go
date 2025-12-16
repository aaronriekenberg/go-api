package config

import (
	"encoding/json"
	"time"
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

type RequestConfiguration struct {
	ExternalHost string
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
	ID           string
	InternalOnly bool
	Description  string
	Command      string
	Args         []string
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
	ServerConfiguration         ServerConfiguration
	RequestConfiguration        RequestConfiguration
	ProfilingConfiguration      ProfilingConfiguration
	RequestLoggingConfiguration RequestLoggingConfiguration
	CommandConfiguration        CommandConfiguration
}
