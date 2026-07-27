package main

import (
	"log"
	"net/http"

	Route "fluxgo/internal/route"
	Routes "fluxgo/route"
)

func main() {
	Routes.Middleware()
	Routes.Web()
	Routes.API()

	addr := ":8080"
	log.Printf("server running at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, Route.HTTPHandler()); err != nil {
		log.Fatal(err)
	}
}
