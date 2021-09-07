package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	f "module/utils"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"google.golang.org/grpc"
)

type RequestBody struct {
	Date string `json:"date,omitempty"`
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

var port string = f.GoDotEnvVariable("BACKEND_PORT")

func main() {

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(corsMiddleware)

	router.Route("/", func(router chi.Router) {
		router.Get("/", describeAPI)
	})

	router.Route("/restaurant-data", func(router chi.Router) {
		router.Use(restaurantCtx)

		router.Post("/", loadRestaurantData)
	})

	router.Route("/buyer", func(router chi.Router) {
		router.Use(buyersCtx)
		router.Get("/all", getBuyers)

		router.Route("/{buyerId}", func(router chi.Router) {
			router.Use(buyerCtx)
			router.Get("/", getBuyer)
		})
	})

	router.Route("/products", func(router chi.Router) {
		router.Use(productsCtx)

		router.Get("/", getProducts)
	})

	fmt.Printf("Server listening on port %s\n", port)

	err := http.ListenAndServe(":"+port, router)
	if err != nil {
		log.Fatal(err)
	}
}

func newDGraphClient() *dgo.Dgraph {
	target := f.GoDotEnvVariable("DGRAPH_ALPHA")
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
