package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
)

type Buyer struct {
	BuyerId      string
	Name         string
	Transactions []Transaction
	Type         string `json:"dgraph.type,omitempty"`
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

func main() {
	ctx := context.Background()
	dgraphClient := newClient()
	txn := dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	fetchTransactions(txn)

}

func newClient() *dgo.Dgraph {
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	dc := api.NewDgraphClient(d)
	return dgo.NewDgraphClient(dc)
}

func fetchTransactions(txn *dgo.Txn) {
	jsonTransactions := parseTransactions()

	mutation := &api.Mutation{
		SetJson:   jsonTransactions,
		CommitNow: true,
	}

	_, err := txn.Mutate(context.Background(), mutation)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("New transactions saved.")
}

func parseTransactions() []byte {
	var transactions []Transaction

	req, err := http.NewRequest("GET", "https://kqxty15mpg.execute-api.us-east-1.amazonaws.com/transactions", nil)

	if err != nil {
		fmt.Println(err)
	}

	q := req.URL.Query()
	q.Add("date", "1624680000")
	req.URL.RawQuery = q.Encode()
	query := req.URL.String()

	resp, err := http.Get(query)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	rawTransactions := strings.Split(string(body), "#")
	transactionsQty := len(rawTransactions)

	for i := 0; i < transactionsQty; i++ {
		line := rawTransactions[i]
		size := len(line)

		if size == 0 {
			continue
		}

		transactionId := line[0:12]
		buyerId := line[12:21]
		deviceRgx := regexp.MustCompile(`[a-z]`)
		productRgx := regexp.MustCompile(`\(`)

		deviceIndex := deviceRgx.FindStringIndex(line[21 : size-1])[0]
		if deviceIndex == 0 {
			continue
		}

		productIndex := productRgx.FindStringIndex(line[21 : size-1])[0]

		ip := line[21 : deviceIndex+21]
		device := line[deviceIndex+21 : productIndex+21]
		products := line[productIndex+21:]
		products = products[1 : len(products)-3]

		newTransaction := Transaction{
			BuyerId:       buyerId,
			TransactionId: transactionId,
			Ip:            ip,
			Device:        device,
			Products:      strings.Split(products, ","),
			Date:          "2020-08-17",
			Type:          "Transaction",
		}

		transactions = append(transactions, newTransaction)

	}

	rawJsonTransactions, err := json.Marshal(transactions)

	if err != nil {
		fmt.Println(err)
	}

	/*
		Replace all appearances of the unicode null character: \u0000 with an
		empty string.
	*/
	jsonTransactions := strings.Replace(string(rawJsonTransactions), "\\u0000", "", -1)
	return []byte(jsonTransactions)

}
