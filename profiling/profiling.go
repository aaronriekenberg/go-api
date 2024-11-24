package profiling

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime/debug"

	"github.com/aaronriekenberg/go-api/config"
)

func Start(config config.ProfilingConfiguration) {
	if !config.Enabled {
		return
	}

	go runPprofServer(config)
}

func runPprofServer(config config.ProfilingConfiguration) {
	defer func() {
		if err := recover(); err != nil {
			slog.Error("panic in runPprofServer",
				"error", err,
			)
			fmt.Fprintf(os.Stderr, "stack trace:\n%v", string(debug.Stack()))
			os.Exit(1)
		}
	}()

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
	panic(fmt.Errorf("runPprofServer: http.ListenAndServe returned error %w", err))
}
