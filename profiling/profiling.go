package profiling

import (
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"

	"github.com/aaronriekenberg/go-api/config"
)

func Start(config config.ProfilingConfiguration) {
	if !config.Enabled {
		return
	}

	go runPprofServer(config)
}

func runPprofServer(config config.ProfilingConfiguration) {
	slog.Info("begin runPprofServer",
		"config", config,
	)

	serveMux := http.NewServeMux()

	serveMux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	serveMux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	serveMux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	serveMux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	serveMux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	err := http.ListenAndServe(config.ListenAddress, serveMux)
	slog.Error("pprof server error",
		"error", err,
	)
	os.Exit(1)
}
