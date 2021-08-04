package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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

/*
	Extracts the url parameter from the request and adds it to
	the context so that the handlers have can use it.
*/
func RestaurantCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writter http.ResponseWriter, request *http.Request) {
		body, err := io.ReadAll(request.Body)

		var requestBody RequestBody
		json.Unmarshal(body, &requestBody)

		if err != nil {
			fmt.Println(err)
		}

		if err != nil {
			http.Error(writter, http.StatusText(404), 404)
			return
		}
		ctx := context.WithValue(request.Context(), "date", requestBody.Date)
		next.ServeHTTP(writter, request.WithContext(ctx))
	})
}

func loadRestaurantData(writter http.ResponseWriter, request *http.Request) {
	requestContext := request.Context()
	date, ok := requestContext.Value("date").(string)

	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	dataLoader := &DataLoader{
		dateStr: date,
		txn:     txn,
	}

	writter.Write([]byte(dataLoader.loadRestaurantData()))
	writter.Header().Set("Content-Type", "text/plain")

	if !ok {
		http.Error(writter, http.StatusText(http.StatusUnprocessableEntity),
			http.StatusUnprocessableEntity)
		return
	}
}
