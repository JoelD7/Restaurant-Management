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
	name         string        `json:"name,omitempty"`
	transactions []Transaction `json:"transactions,omitempty"`
}

type Transaction struct {
	device string `json:"device,omitempty"`
}

func main() {
	// ctx := context.Background()
	dgraphClient := newClient()
	// txn := dgraphClient.NewTxn()
	// defer txn.Discard(ctx)

	const q = `{
		buyer(func: has(name)) {
			name
			transactions
		}
	}
`

	resp, err := dgraphClient.NewTxn().Query(context.Background(), q)
	if err != nil {
		log.Fatal(err)
	}

	type Root struct {
		Buyers []Buyer `json:"buyer"`
	}

	var r Root
	err = json.Unmarshal(resp.Json, &r)
	if err != nil {
		log.Fatal(err)
	}

	out, _ := json.MarshalIndent(r, "", "\t")
	fmt.Printf("%s\n", out)

}

func newClient() *dgo.Dgraph {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	d, err := grpc.Dial("127.0.0.1:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	dc := api.NewDgraphClient(d)
	return dgo.NewDgraphClient(dc)
}

func parseTransactions() {
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
	body, err := io.ReadAll(resp.Body)

	transactions := strings.Split(string(body), "#")

	for i := 0; i < len(transactions); i++ {
		line := transactions[i]
		size := len(line)

		if size == 0 {
			continue
		}

		transactionId := line[0:12]
		buyerId := line[12:21]
		deviceRgx := regexp.MustCompile(`[a-z]`)
		productRgx := regexp.MustCompile(`\(`)

		deviceIndex := deviceRgx.FindStringIndex(line[21 : size-1])[0]
		fmt.Println(deviceIndex)
		if deviceIndex == 0 {
			continue
		}

		productIndex := productRgx.FindStringIndex(line[21 : size-1])[0]

		ip := line[21 : deviceIndex+21]
		device := line[deviceIndex+21 : productIndex+21]
		products := line[productIndex+21:]

		fmt.Printf("transactionId: %s, buyerId: %s, ip: %s, device: %s, products: %s\n", transactionId, buyerId, ip, device, products)
	}
}
