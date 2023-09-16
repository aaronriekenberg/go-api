package main

import (
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/aaronriekenberg/go-api/requestinfo"
)

// func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	fmt.Fprint(w, "Welcome!\n")
// }

// func Hello(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))
// }

func main() {
	router := httprouter.New()
	router.GET("/request_info", requestinfo.CreateHandler())
	// router.GET("/hello/:name", Hello)

	log.Fatal(http.ListenAndServe(":8080", router))
}
