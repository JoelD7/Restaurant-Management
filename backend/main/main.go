package main

import (
	"context"
	"log"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	d "github.com/shopspring/decimal"
	"google.golang.org/grpc"
)

type Buyer struct {
	BuyerId      string
	Name         string
	Transactions []Transaction
	Type         string `json:"dgraph.type,omitempty"`
}

type Product struct {
	ProductId string
	Name      string
	Date      string
	Price     d.Decimal
}

type Transaction struct {
	TransactionId string
	BuyerId       string
	Ip            string
	Device        string
	Products      []string
	Date          string
	Type          string `json:"dgraph.type,omitempty"`
}

var ctx context.Context = context.Background()
var dgraphClient *dgo.Dgraph = newClient()

func main() {
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	// dateStr:="2020-08-17T00:00:00Z"

}

func newClient() *dgo.Dgraph {
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	dc := api.NewDgraphClient(d)
	return dgo.NewDgraphClient(dc)
}
