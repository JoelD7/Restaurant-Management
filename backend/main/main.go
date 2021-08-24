package main

import (
	"context"
	"encoding/json"
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

type APIDescriptor struct {
	Method      string
	Endpoint    string
	Body        string `json:"Body,omitempty"`
	URLParam    string `json:"URLParam,omitempty"`
	Description string
}

var ctx context.Context = context.Background()
var dgraphClient = newDGraphClient()
var descriptor []APIDescriptor = []APIDescriptor{
	{
		Method:      http.MethodPost,
		Endpoint:    "/restaurant-data",
		Description: "Loads all restaurant related data of the specified date to the database.",
		Body:        "'date' in yyyy-MM-DD format",
	},
	{
		Method:      http.MethodGet,
		Endpoint:    "/buyer/all",
		Description: "Returns all the buyers currently saved on the database.",
	},
	{
		Method:      http.MethodGet,
		Endpoint:    "/buyer/{buyerId}",
		Description: "Returns the buyer with the id 'buyerId'.",
	},
}

const port string = "9000"

func main() {

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(Cors)

	router.Route("/", func(router chi.Router) {
		router.Get("/", describeAPI)
	})

	router.Route("/restaurant-data", func(router chi.Router) {
		router.Use(restaurantCtx)

		router.Post("/", loadRestaurantData)
	})

	router.Route("/buyer", func(router chi.Router) {
		router.Use(BuyersCtx)
		router.Get("/all", getBuyers)

		router.Route("/{buyerId}", func(router chi.Router) {
			router.Use(BuyerCtx)
			router.Get("/", getBuyer)
		})
	})

	router.Route("/products", func(router chi.Router) {
		router.Use(ProductsCtx)

		router.Get("/", getProducts)
	})

	fmt.Printf("Server listening on port %s\n", port)
	// Manejar error de ListenAndServe
	http.ListenAndServe(":"+port, router)

}

func newDGraphClient() *dgo.Dgraph {
	target := "localhost:9080"
	clientConn, err := grpc.Dial(target, grpc.WithInsecure())
	if err != nil {
		log.Fatal(fmt.Errorf("error ocurred while trying to establish connection with '%s': %w", target, err))
	}

	dc := api.NewDgraphClient(clientConn)
	return dgo.NewDgraphClient(dc)
}

func describeAPI(writter http.ResponseWriter, request *http.Request) {
	jsonDescriptor, err := json.Marshal(descriptor)

	if err != nil {
		http.Error(writter, "error while processing response", http.StatusInternalServerError)
		return
	}

	writter.Write(jsonDescriptor)
}
