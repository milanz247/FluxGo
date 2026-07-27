package main

import (
	"log"
	"net/http"

	Route "fluxgo/internal/route"
	"fluxgo/internal/view"
	Routes "fluxgo/route"
)

func main() {
	views, err := view.New(view.Config{Root: "views"})
	if err != nil {
		log.Fatalf("boot views: %v", err)
	}
	Route.SetRenderer(views)

	Routes.Middleware()
	Routes.Web()
	Routes.API()

	addr := ":8080"
	log.Printf("server running at http://localhost%s", addr)
	if err := http.ListenAndServe(addr, Route.HTTPHandler()); err != nil {
		log.Fatal(err)
	}
}
