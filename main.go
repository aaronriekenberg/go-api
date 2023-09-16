package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/requestinfo"
)

// func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	fmt.Fprint(w, "Welcome!\n")
// }

// func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
// }

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	if len(os.Args) != 2 {
		logger.Error("config file required as command line arument")
	}

	configFile := os.Args[1]

	config := config.ReadConfiguration(configFile, logger)

	logger.Info("read configuration",
		"config", config)

	router := httprouter.New()
	router.GET("/request_info", requestinfo.CreateHandler())
	// router.GET("/hello/:name", Hello)

	err := http.ListenAndServe(config.ServerConfiguration.ListenAddress, router)
	logger.Error("http.ListenAndServe error",
		"error", err)
	os.Exit(1)
}
