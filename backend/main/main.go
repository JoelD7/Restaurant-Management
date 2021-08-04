package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"google.golang.org/grpc"
)

type RequestBody struct {
	Date string `json:"date"`
}

var ctx context.Context = context.Background()
var dgraphClient *dgo.Dgraph = newClient()

const port string = "9000"

func main() {

	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Route("/restaurant-data", func(router chi.Router) {
		router.Use(RestaurantCtx)

		router.Post("/", loadRestaurantData)
	})

	router.Route("/buyers", func(router chi.Router) {
		router.Get("/", getBuyers)
	})

	fmt.Printf("Server listening on port %s\n", port)
	http.ListenAndServe(":"+port, router)

}

func newClient() *dgo.Dgraph {
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	dc := api.NewDgraphClient(d)
	return dgo.NewDgraphClient(dc)
}
