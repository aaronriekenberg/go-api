package main

import (
	"log/slog"
	"os"

	"github.com/julienschmidt/httprouter"

	"github.com/aaronriekenberg/go-api/config"
	"github.com/aaronriekenberg/go-api/requestinfo"
	"github.com/aaronriekenberg/go-api/server"
)

// func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	fmt.Fprint(w, "Welcome!\n")
// }

// func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
// }

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	if len(os.Args) != 2 {
		slog.Error("config file required as command line arument")
	}

	configFile := os.Args[1]

	config := config.ReadConfiguration(configFile)

	slog.Info("read configuration",
		"config", config)

	router := httprouter.New()
	router.GET("/request_info", requestinfo.CreateHandler())
	// router.GET("/hello/:name", Hello)

	server.Run(config.ServerConfiguration, router)
}
