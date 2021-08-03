package main

import (
	"context"
	"fmt"
	"log"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"

	"google.golang.org/grpc"
)

var ctx context.Context = context.Background()
var dgraphClient *dgo.Dgraph = newClient()

func main() {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	dataLoader := &DataLoader{
		dateStr: "2020-08-17T00:00:00Z",
		txn:     txn,
	}

	fmt.Println(string(dataLoader.fetchBuyers()))

	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

	// 	json := dataLoader.fetchTransactions()

	// 	w.Header().Set("Content-Type", "application/json")
	// 	fmt.Fprint(w, string(json))
	// })

	// http.ListenAndServe(":7070", nil)

}

func newClient() *dgo.Dgraph {
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	dc := api.NewDgraphClient(d)
	return dgo.NewDgraphClient(dc)
}
